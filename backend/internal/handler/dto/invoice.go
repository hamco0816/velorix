package dto

import (
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

// InvoiceItem 发票申请单响应。带 User* 的字段仅管理端填充。
type InvoiceItem struct {
	ID             int64      `json:"id"`
	RecipientEmail string     `json:"recipient_email"`
	TitleType      string     `json:"title_type"`
	TitleName      string     `json:"title_name"`
	TaxID          string     `json:"tax_id,omitempty"`
	UserRemark     string     `json:"user_remark,omitempty"`
	Amount         float64    `json:"amount"`
	Status         string     `json:"status"`
	InvoiceNumber  string     `json:"invoice_number,omitempty"`
	InvoiceDate    *time.Time `json:"invoice_date,omitempty"`
	InvoiceAmount  *float64   `json:"invoice_amount,omitempty"`
	IssuedAt       *time.Time `json:"issued_at,omitempty"`
	RejectReason   string     `json:"reject_reason,omitempty"`
	EmailSent      bool       `json:"email_sent"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	UserID    int64  `json:"user_id,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
	UserName  string `json:"user_name,omitempty"`
}

// InvoiceOrderItem 申请单关联/可开票订单的精简展示。
type InvoiceOrderItem struct {
	ID          int64      `json:"id"`
	OrderType   string     `json:"order_type"`
	Amount      float64    `json:"amount"`     // 到账金额（含赠送，仅参考）
	PayAmount   float64    `json:"pay_amount"` // 实付金额（开票口径）
	PaymentType string     `json:"payment_type"`
	Status      string     `json:"status"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// InvoiceFromEnt 将 ent 实体映射为响应 DTO；includeUser=true 时附带申请人信息（管理端）。
func InvoiceFromEnt(ir *dbent.InvoiceRequest, includeUser bool) InvoiceItem {
	item := InvoiceItem{
		ID:             int64(ir.ID),
		RecipientEmail: ir.RecipientEmail,
		TitleType:      ir.TitleType,
		TitleName:      ir.TitleName,
		Amount:         ir.Amount,
		Status:         ir.Status,
		InvoiceDate:    ir.InvoiceDate,
		InvoiceAmount:  ir.InvoiceAmount,
		IssuedAt:       ir.IssuedAt,
		EmailSent:      ir.EmailSent,
		CreatedAt:      ir.CreatedAt,
		UpdatedAt:      ir.UpdatedAt,
	}
	if ir.TaxID != nil {
		item.TaxID = *ir.TaxID
	}
	if ir.UserRemark != nil {
		item.UserRemark = *ir.UserRemark
	}
	if ir.InvoiceNumber != nil {
		item.InvoiceNumber = *ir.InvoiceNumber
	}
	if ir.RejectReason != nil {
		item.RejectReason = *ir.RejectReason
	}
	if includeUser {
		item.UserID = ir.UserID
		item.UserEmail = ir.UserEmail
		item.UserName = ir.UserName
	}
	return item
}

// InvoicesFromEnt 批量映射申请单。
func InvoicesFromEnt(items []*dbent.InvoiceRequest, includeUser bool) []InvoiceItem {
	out := make([]InvoiceItem, 0, len(items))
	for _, ir := range items {
		out = append(out, InvoiceFromEnt(ir, includeUser))
	}
	return out
}

// InvoiceOrdersFromEnt 批量映射订单为精简展示项。
func InvoiceOrdersFromEnt(orders []*dbent.PaymentOrder) []InvoiceOrderItem {
	out := make([]InvoiceOrderItem, 0, len(orders))
	for _, o := range orders {
		out = append(out, InvoiceOrderItem{
			ID:          int64(o.ID),
			OrderType:   o.OrderType,
			Amount:      o.Amount,
			PayAmount:   o.PayAmount,
			PaymentType: o.PaymentType,
			Status:      o.Status,
			PaidAt:      o.PaidAt,
			CreatedAt:   o.CreatedAt,
		})
	}
	return out
}
