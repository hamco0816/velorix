//go:build unit

package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

// --- 测试替身 ---

// invoiceTestMailer 记录发票邮件发送，可注入失败。
type invoiceTestMailer struct {
	calls     int
	lastTo    string
	lastBytes []byte
	failWith  error
}

func (m *invoiceTestMailer) SendEmailWithAttachment(_ context.Context, to, _, _, _ string, attachment []byte, _ string) error {
	m.calls++
	m.lastTo = to
	m.lastBytes = attachment
	return m.failWith
}

// invoiceTestSettingRepo 仅用于发票开关：按构造时的 enabled 返回 invoice_enabled。
type invoiceTestSettingRepo struct{ enabled bool }

func (r *invoiceTestSettingRepo) Get(context.Context, string) (*service.Setting, error) {
	return nil, nil
}
func (r *invoiceTestSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	if key == service.SettingKeyInvoiceEnabled {
		if r.enabled {
			return "true", nil
		}
		return "false", nil
	}
	return "", nil
}
func (r *invoiceTestSettingRepo) Set(context.Context, string, string) error { return nil }
func (r *invoiceTestSettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (r *invoiceTestSettingRepo) SetMultiple(context.Context, map[string]string) error { return nil }
func (r *invoiceTestSettingRepo) GetAll(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}
func (r *invoiceTestSettingRepo) Delete(context.Context, string) error { return nil }

// --- 测试环境搭建 ---

type invoiceTestEnv struct {
	router *gin.Engine
	client *dbent.Client
	mailer *invoiceTestMailer
}

func newInvoiceTestEnv(t *testing.T, invoiceEnabled bool) *invoiceTestEnv {
	t.Helper()

	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", t.Name())
	db, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	mailer := &invoiceTestMailer{}
	settingService := service.NewSettingService(&invoiceTestSettingRepo{enabled: invoiceEnabled}, nil)
	invoiceService := service.NewInvoiceService(client, mailer, settingService)
	userHandler := NewInvoiceHandler(invoiceService, settingService)
	adminHandler := admin.NewInvoiceHandler(invoiceService)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 测试鉴权中间件：从 X-Test-User-ID 注入调用者身份
	r.Use(func(c *gin.Context) {
		if v := c.GetHeader("X-Test-User-ID"); v != "" {
			id, _ := strconv.ParseInt(v, 10, 64)
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: id})
		}
		c.Next()
	})

	v1 := r.Group("/api/v1")
	inv := v1.Group("/invoices")
	{
		inv.GET("/invoiceable-orders", userHandler.GetInvoiceableOrders)
		inv.GET("/my", userHandler.GetMyInvoices)
		inv.POST("", userHandler.ApplyInvoice)
		inv.GET("/:id", userHandler.GetMyInvoice)
		inv.POST("/:id/cancel", userHandler.CancelInvoice)
	}
	adm := v1.Group("/admin/invoices")
	{
		adm.GET("", adminHandler.List)
		adm.GET("/:id", adminHandler.GetDetail)
		adm.POST("/:id/parse-pdf", adminHandler.ParsePDF)
		adm.POST("/:id/issue", adminHandler.Issue)
		adm.POST("/:id/reject", adminHandler.Reject)
	}

	return &invoiceTestEnv{router: r, client: client, mailer: mailer}
}

// seedUser 创建测试用户。
func (e *invoiceTestEnv) seedUser(t *testing.T, email string) *dbent.User {
	t.Helper()
	u, err := e.client.User.Create().
		SetEmail(email).SetPasswordHash("hash").SetUsername(email).
		Save(context.Background())
	require.NoError(t, err)
	return u
}

// seedCompletedOrder 创建已完成付费订单。
func (e *invoiceTestEnv) seedCompletedOrder(t *testing.T, user *dbent.User, payAmount float64, suffix string) *dbent.PaymentOrder {
	t.Helper()
	o, err := e.client.PaymentOrder.Create().
		SetUserID(user.ID).SetUserEmail(user.Email).SetUserName(user.Username).
		SetAmount(payAmount).SetPayAmount(payAmount).SetFeeRate(0).
		SetRechargeCode("HINV-" + suffix).SetOutTradeNo("sub2_hinv_" + suffix).
		SetPaymentType(payment.TypeAlipay).SetPaymentTradeNo("htrade-" + suffix).
		SetOrderType(payment.OrderTypeBalance).SetStatus(service.OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").SetSrcHost("api.example.com").
		Save(context.Background())
	require.NoError(t, err)
	return o
}

// --- HTTP 辅助 ---

// envelope 统一响应体 {code,message,data}。
type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (e *invoiceTestEnv) do(t *testing.T, method, path string, userID int64, body *bytes.Buffer, contentType string) (int, envelope) {
	t.Helper()
	var reader *bytes.Buffer = body
	if reader == nil {
		reader = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, path, reader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if userID > 0 {
		req.Header.Set("X-Test-User-ID", strconv.FormatInt(userID, 10))
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	var env envelope
	if rec.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	}
	return rec.Code, env
}

func (e *invoiceTestEnv) postJSON(t *testing.T, path string, userID int64, payload any) (int, envelope) {
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	return e.do(t, http.MethodPost, path, userID, bytes.NewBuffer(raw), "application/json")
}

// buildInvoiceMultipart 构造带文件字段 file 的 multipart 表单。filename 为空表示不带文件。
func buildInvoiceMultipart(t *testing.T, filename string, content []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	if filename != "" {
		fw, err := w.CreateFormFile("file", filename)
		require.NoError(t, err)
		_, err = fw.Write(content)
		require.NoError(t, err)
	}
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	require.NoError(t, w.Close())
	return body, w.FormDataContentType()
}

// --- 用例 ---

// 完整流程：用户申请 → 管理员上传 PDF 开票（发邮件）→ 状态变为已开票。
func TestInvoiceFlow_ApplyThenIssue(t *testing.T) {
	env := newInvoiceTestEnv(t, true)
	user := env.seedUser(t, "flow@example.com")
	o1 := env.seedCompletedOrder(t, user, 100, "f1")
	o2 := env.seedCompletedOrder(t, user, 50, "f2")

	// 用户申请开票
	status, env1 := env.postJSON(t, "/api/v1/invoices", user.ID, map[string]any{
		"recipient_email": "recipient@example.com",
		"title_type":      "personal",
		"title_name":      "个人",
		"order_ids":       []int64{o1.ID, o2.ID},
	})
	require.Equal(t, http.StatusCreated, status)
	require.Equal(t, 0, env1.Code)
	var applied struct {
		ID     int64   `json:"id"`
		Amount float64 `json:"amount"`
		Status string  `json:"status"`
	}
	require.NoError(t, json.Unmarshal(env1.Data, &applied))
	require.InDelta(t, 150.0, applied.Amount, 0.001)
	require.Equal(t, "pending", applied.Status)

	// 管理员上传 PDF 开票
	body, ct := buildInvoiceMultipart(t, "invoice.pdf", []byte("%PDF-1.4 fake"), map[string]string{
		"invoice_number": "25117000000123456789",
		"invoice_date":   "2026-01-15",
		"invoice_amount": "150.00",
	})
	status, env2 := env.do(t, http.MethodPost, fmt.Sprintf("/api/v1/admin/invoices/%d/issue", applied.ID), 999, body, ct)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, env2.Code)

	// 邮件已发送给接收邮箱
	require.Equal(t, 1, env.mailer.calls)
	require.Equal(t, "recipient@example.com", env.mailer.lastTo)

	// DB 状态已开票
	ir, err := env.client.InvoiceRequest.Get(context.Background(), applied.ID)
	require.NoError(t, err)
	require.Equal(t, "issued", ir.Status)
	require.NotNil(t, ir.InvoiceNumber)
	require.Equal(t, "25117000000123456789", *ir.InvoiceNumber)
	require.True(t, ir.EmailSent)
}

// 开关关闭时申请被 403 拦截。
func TestInvoiceApply_DisabledReturns403(t *testing.T) {
	env := newInvoiceTestEnv(t, false)
	user := env.seedUser(t, "disabled@example.com")
	order := env.seedCompletedOrder(t, user, 100, "d1")

	status, _ := env.postJSON(t, "/api/v1/invoices", user.ID, map[string]any{
		"recipient_email": "recipient@example.com",
		"title_type":      "personal",
		"title_name":      "个人",
		"order_ids":       []int64{order.ID},
	})
	require.Equal(t, http.StatusForbidden, status)
}

// 开票上传非 PDF 文件应被拒绝。
func TestInvoiceIssue_RejectsNonPDF(t *testing.T) {
	env := newInvoiceTestEnv(t, true)
	user := env.seedUser(t, "nonpdf@example.com")
	order := env.seedCompletedOrder(t, user, 100, "p1")
	_, applyEnv := env.postJSON(t, "/api/v1/invoices", user.ID, map[string]any{
		"recipient_email": "recipient@example.com",
		"title_type":      "personal",
		"title_name":      "个人",
		"order_ids":       []int64{order.ID},
	})
	var applied struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(applyEnv.Data, &applied))

	body, ct := buildInvoiceMultipart(t, "invoice.txt", []byte("not a pdf"), map[string]string{
		"invoice_number": "123",
	})
	status, _ := env.do(t, http.MethodPost, fmt.Sprintf("/api/v1/admin/invoices/%d/issue", applied.ID), 999, body, ct)
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, 0, env.mailer.calls)
}

// 开票缺少文件应被拒绝。
func TestInvoiceIssue_RequiresFile(t *testing.T) {
	env := newInvoiceTestEnv(t, true)
	user := env.seedUser(t, "nofile@example.com")
	order := env.seedCompletedOrder(t, user, 100, "nf1")
	_, applyEnv := env.postJSON(t, "/api/v1/invoices", user.ID, map[string]any{
		"recipient_email": "recipient@example.com",
		"title_type":      "personal",
		"title_name":      "个人",
		"order_ids":       []int64{order.ID},
	})
	var applied struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(applyEnv.Data, &applied))

	body, ct := buildInvoiceMultipart(t, "", nil, map[string]string{"invoice_number": "123"})
	status, _ := env.do(t, http.MethodPost, fmt.Sprintf("/api/v1/admin/invoices/%d/issue", applied.ID), 999, body, ct)
	require.Equal(t, http.StatusBadRequest, status)
}

// 上传 PDF 识别接口对乱码返回空字段（不报错）。
func TestInvoiceParsePDF_GarbageReturnsEmpty(t *testing.T) {
	env := newInvoiceTestEnv(t, true)
	user := env.seedUser(t, "parse@example.com")
	order := env.seedCompletedOrder(t, user, 100, "pp1")
	_, applyEnv := env.postJSON(t, "/api/v1/invoices", user.ID, map[string]any{
		"recipient_email": "recipient@example.com",
		"title_type":      "personal",
		"title_name":      "个人",
		"order_ids":       []int64{order.ID},
	})
	var applied struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(applyEnv.Data, &applied))

	body, ct := buildInvoiceMultipart(t, "invoice.pdf", []byte("garbage not pdf"), nil)
	status, parseEnv := env.do(t, http.MethodPost, fmt.Sprintf("/api/v1/admin/invoices/%d/parse-pdf", applied.ID), 999, body, ct)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, parseEnv.Code)
	var parsed struct {
		InvoiceNumber string `json:"invoice_number"`
	}
	require.NoError(t, json.Unmarshal(parseEnv.Data, &parsed))
	require.Empty(t, parsed.InvoiceNumber)
}

// 用户取消待开票申请释放订单，订单重新出现在可开票列表。
func TestInvoiceCancel_ReleasesOrderViaHTTP(t *testing.T) {
	env := newInvoiceTestEnv(t, true)
	user := env.seedUser(t, "httpcancel@example.com")
	order := env.seedCompletedOrder(t, user, 100, "hc1")

	_, applyEnv := env.postJSON(t, "/api/v1/invoices", user.ID, map[string]any{
		"recipient_email": "recipient@example.com",
		"title_type":      "personal",
		"title_name":      "个人",
		"order_ids":       []int64{order.ID},
	})
	var applied struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(applyEnv.Data, &applied))

	// 申请后可开票列表为空
	_, listEnv := env.do(t, http.MethodGet, "/api/v1/invoices/invoiceable-orders", user.ID, nil, "")
	var page struct {
		Total int `json:"total"`
	}
	require.NoError(t, json.Unmarshal(listEnv.Data, &page))
	require.Equal(t, 0, page.Total)

	// 取消
	status, _ := env.do(t, http.MethodPost, fmt.Sprintf("/api/v1/invoices/%d/cancel", applied.ID), user.ID, nil, "")
	require.Equal(t, http.StatusOK, status)

	// 取消后订单重新可开票
	_, listEnv2 := env.do(t, http.MethodGet, "/api/v1/invoices/invoiceable-orders", user.ID, nil, "")
	require.NoError(t, json.Unmarshal(listEnv2.Data, &page))
	require.Equal(t, 1, page.Total)
}

// 越权：用户不能访问他人申请单。
func TestInvoiceGet_RejectsOtherUser(t *testing.T) {
	env := newInvoiceTestEnv(t, true)
	owner := env.seedUser(t, "owner@example.com")
	other := env.seedUser(t, "other@example.com")
	order := env.seedCompletedOrder(t, owner, 100, "o1")

	_, applyEnv := env.postJSON(t, "/api/v1/invoices", owner.ID, map[string]any{
		"recipient_email": "recipient@example.com",
		"title_type":      "personal",
		"title_name":      "个人",
		"order_ids":       []int64{order.ID},
	})
	var applied struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(applyEnv.Data, &applied))

	status, _ := env.do(t, http.MethodGet, fmt.Sprintf("/api/v1/invoices/%d", applied.ID), other.ID, nil, "")
	require.Equal(t, http.StatusForbidden, status)
}
