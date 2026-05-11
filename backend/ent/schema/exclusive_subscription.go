package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ExclusiveSubscription 独享名额：一个用户独占一个上游账号一段时间。
//
// 与 UserSubscription（共享池订阅）的关系：
//   - UserSubscription 表示「用户在某个共享池内有用量额度」
//   - ExclusiveSubscription 表示「用户从某个独享池中获得了一个具体账号的独占使用权」
//
// 两者并存，调度器优先命中 ExclusiveSubscription，否则回退到共享池。
type ExclusiveSubscription struct {
	ent.Schema
}

func (ExclusiveSubscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "exclusive_subscriptions"},
	}
}

func (ExclusiveSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (ExclusiveSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("group_id"),
		field.Int64("plan_id"),
		field.Int64("account_id"),

		// 状态：active / expired / refunded / cancelled
		field.String("status").
			MaxLen(20).
			Default(domain.ExclusiveSeatStatusActive),

		field.Time("starts_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("assigned_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("last_renewal_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		// 累计用量统计（USD），由调度器在记录 usage_log 时同步累加
		field.Float("usage_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),

		// 日/周/月窗口用量（migration 140）：与 daily/weekly/monthly_limit_usd 配套，
		// 用于"请求前"按窗口判断是否超限，超限直接拒绝。窗口起点字段由 lazy reset 维护。
		field.Time("daily_window_start").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("weekly_window_start").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("monthly_window_start").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Float("daily_usage_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Float("weekly_usage_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Float("monthly_usage_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),

		field.String("notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 管理员赠送时记录管理员 ID，正常支付分配为 nil
		field.Int64("assigned_by").
			Optional().
			Nillable(),

		// 创建该 seat 的支付订单 ID（管理员赠送/续费分配新号时为 nil）
		// 用于退款时精确定位需要释放的 seat，避免在多份同 plan 名额时误释放
		field.Int64("source_order_id").
			Optional().
			Nillable(),

		// 限额/倍率快照（migration 138）：购买时从 plan 拷贝；NULL = 回落到 group
		field.Float("daily_limit_usd").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("weekly_limit_usd").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("monthly_limit_usd").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("rate_multiplier").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}),
	}
}

// 通过 user_id / group_id / plan_id / account_id 字段关联，不强约束 edge，
// 避免和现有 schema 的 inverse 边互相打架；如有需要可后续补 edge.From。

func (ExclusiveSubscription) Indexes() []ent.Index {
	return []ent.Index{
		// 同账号同时只能被一份 active seat 占用（线上由部分唯一索引保证，
		// schema 这里用复合普通索引让查询计划合理；唯一约束放在迁移文件）
		index.Fields("account_id"),
		index.Fields("user_id", "status", "group_id"),
		index.Fields("status", "expires_at"),
		index.Fields("plan_id"),
		index.Fields("group_id"),
		index.Fields("source_order_id"),
		index.Fields("deleted_at"),
	}
}
