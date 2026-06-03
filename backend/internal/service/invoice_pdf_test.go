//go:build unit

package service

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 发票号码识别：从去空白后的发票文本中抽取 8~20 位号码。
func TestReInvoiceNumber(t *testing.T) {
	cases := map[string]string{
		"发票号码:25117000000123456789": "25117000000123456789",
		"发票号码：12345678":             "12345678",
		"其他字段发票号码20240115001结束":     "20240115001",
		"没有号码这一行":                   "",
	}
	for text, want := range cases {
		compact := strings.ReplaceAll(text, " ", "")
		m := reInvoiceNumber.FindStringSubmatch(compact)
		if want == "" {
			require.Len(t, m, 0, "input=%q", text)
			continue
		}
		require.Len(t, m, 2, "input=%q", text)
		require.Equal(t, want, m[1])
	}
}

// 价税合计金额识别：支持 ￥/¥ 前缀及任意分隔符。
func TestReInvoiceAmount(t *testing.T) {
	cases := map[string]string{
		"价税合计（小写）￥1234.56": "1234.56",
		"价税合计¥99.00":       "99.00",
		"价税合计 100.00":      "100.00",
	}
	for text, want := range cases {
		compact := strings.ReplaceAll(text, " ", "")
		m := reInvoiceAmount.FindStringSubmatch(compact)
		require.Len(t, m, 2, "input=%q", text)
		require.Equal(t, want, m[1])
		_, err := strconv.ParseFloat(m[1], 64)
		require.NoError(t, err)
	}
}

// 开票日期识别：同时兼容中文与 ISO 写法。
func TestParseInvoiceDate(t *testing.T) {
	cn, ok := parseInvoiceDate("开票日期:2024年01月15日")
	require.True(t, ok)
	require.Equal(t, 2024, cn.Year())
	require.Equal(t, 1, int(cn.Month()))
	require.Equal(t, 15, cn.Day())

	iso, ok := parseInvoiceDate("开票日期：2024-12-31")
	require.True(t, ok)
	require.Equal(t, 12, int(iso.Month()))
	require.Equal(t, 31, iso.Day())

	_, ok = parseInvoiceDate("开票日期:2024年13月40日")
	require.False(t, ok)

	_, ok = parseInvoiceDate("无日期")
	require.False(t, ok)
}

// 解析非法 PDF 字节不应 panic，返回空文本。
func TestExtractInvoicePDFText_GarbageIsSafe(t *testing.T) {
	require.Equal(t, "", extractInvoicePDFText([]byte("not a pdf at all")))
	require.Equal(t, "", extractInvoicePDFText(nil))
}

// parseInvoicePDF 对非 PDF 输入返回空结构（不报错）。
func TestParseInvoicePDF_EmptyOnGarbage(t *testing.T) {
	got := parseInvoicePDF([]byte("garbage"))
	require.Empty(t, got.InvoiceNumber)
	require.Nil(t, got.InvoiceDate)
	require.Nil(t, got.InvoiceAmount)
}
