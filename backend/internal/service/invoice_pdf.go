package service

import (
	"bytes"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

// ParsedInvoice 从发票 PDF 中尽力识别出的元数据。
// 识别是 best-effort：任一字段都可能为空，由管理员在开票时人工确认/补填。
type ParsedInvoice struct {
	InvoiceNumber string
	InvoiceDate   *time.Time
	InvoiceAmount *float64
}

// 发票 PDF 文本里的关键字段正则（先把空白去掉再匹配，兼容不同排版）。
var (
	reInvoiceNumber  = regexp.MustCompile(`发票号码[：:]*([0-9]{8,20})`)
	reInvoiceDateCN  = regexp.MustCompile(`开票日期[：:]*([0-9]{4})年([0-9]{1,2})月([0-9]{1,2})日`)
	reInvoiceDateISO = regexp.MustCompile(`开票日期[：:]*([0-9]{4})[-/]([0-9]{1,2})[-/]([0-9]{1,2})`)
	reInvoiceAmount  = regexp.MustCompile(`价税合计[^0-9]*[￥¥]?([0-9]+\.[0-9]{2})`)
)

// extractInvoicePDFText 尽力抽取 PDF 全文。识别失败返回空串（不报错，best-effort）。
// 用 recover 兜底，防止第三方解析库在异常 PDF 上 panic 影响开票主流程。
func extractInvoicePDFText(data []byte) (text string) {
	defer func() { _ = recover() }()
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ""
	}
	plain, err := reader.GetPlainText()
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, plain); err != nil {
		return ""
	}
	return buf.String()
}

// parseInvoicePDF 从发票 PDF 中尽力识别发票号码、开票日期、价税合计金额。
func parseInvoicePDF(data []byte) ParsedInvoice {
	var result ParsedInvoice
	compact := strings.ReplaceAll(extractInvoicePDFText(data), " ", "")
	if compact == "" {
		return result
	}
	if m := reInvoiceNumber.FindStringSubmatch(compact); len(m) == 2 {
		result.InvoiceNumber = m[1]
	}
	if t, ok := parseInvoiceDate(compact); ok {
		result.InvoiceDate = &t
	}
	if m := reInvoiceAmount.FindStringSubmatch(compact); len(m) == 2 {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			result.InvoiceAmount = &v
		}
	}
	return result
}

// parseInvoiceDate 同时兼容 "2024年01月15日" 和 "2024-01-15" / "2024/01/15" 两种开票日期写法。
func parseInvoiceDate(compact string) (time.Time, bool) {
	if m := reInvoiceDateCN.FindStringSubmatch(compact); len(m) == 4 {
		return buildInvoiceDate(m[1], m[2], m[3])
	}
	if m := reInvoiceDateISO.FindStringSubmatch(compact); len(m) == 4 {
		return buildInvoiceDate(m[1], m[2], m[3])
	}
	return time.Time{}, false
}

func buildInvoiceDate(year, month, day string) (time.Time, bool) {
	y, err1 := strconv.Atoi(year)
	mo, err2 := strconv.Atoi(month)
	d, err3 := strconv.Atoi(day)
	if err1 != nil || err2 != nil || err3 != nil {
		return time.Time{}, false
	}
	if mo < 1 || mo > 12 || d < 1 || d > 31 {
		return time.Time{}, false
	}
	return time.Date(y, time.Month(mo), d, 0, 0, 0, 0, time.UTC), true
}
