package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InvoiceRequest holds the schema definition for the InvoiceRequest entity.
//
// 发票申请单：用户对一笔或多笔「已完成」付费订单合并申请开票。
//
// 删除策略：硬删除
// 与 PaymentOrder 一致，通过 status 字段追踪生命周期（待开票/已开票/已驳回/已取消），
// 实际业务中不删除，保留作为开票审计记录。
//
// 金额口径：amount 取所选订单的 pay_amount（用户实付现金）之和，
// 不含赠送倍率、不含折扣，只算客户真实支付的金额。
type InvoiceRequest struct {
	ent.Schema
}

func (InvoiceRequest) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "invoice_requests"},
	}
}

func (InvoiceRequest) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (InvoiceRequest) Fields() []ent.Field {
	return []ent.Field{
		// 申请人信息（冗余存储，避免关联查询，与 PaymentOrder 一致）
		field.Int64("user_id"),
		field.String("user_email").
			MaxLen(255),
		field.String("user_name").
			MaxLen(100).
			Default(""),

		// 接收发票的邮箱
		field.String("recipient_email").
			MaxLen(255),

		// 抬头信息（普票）：title_type 取 personal(个人) / company(企业)
		field.String("title_type").
			MaxLen(20).
			Default("personal"),
		field.String("title_name").
			MaxLen(255),
		// 纳税人识别号（企业必填，个人留空）
		field.String("tax_id").
			MaxLen(64).
			Optional().
			Nillable(),
		// 用户备注（开票内容说明等）
		field.String("user_remark").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 申请开票总金额 = 所选订单 pay_amount 之和（申请时快照）
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),

		// 状态：pending 待开票 / issued 已开票 / rejected 已驳回 / cancelled 用户取消
		field.String("status").
			MaxLen(20).
			Default("pending"),

		// 开票后留存的发票元数据（PDF 文件本身不入库）
		field.String("invoice_number").
			MaxLen(64).
			Optional().
			Nillable(),
		field.Time("invoice_date").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Float("invoice_amount").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),

		// 开票操作记录
		field.Time("issued_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int64("issued_by").
			Optional().
			Nillable(),

		// 驳回原因
		field.String("reject_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 发票邮件发送结果
		field.Bool("email_sent").
			Default(false),
		field.String("email_error").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
	}
}

func (InvoiceRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("invoice_requests").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (InvoiceRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
