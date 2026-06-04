package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SubscriptionPlan holds the schema definition for the SubscriptionPlan entity.
//
// 删除策略：硬删除
// SubscriptionPlan 使用硬删除而非软删除，原因如下：
//   - 套餐为管理员维护的商品配置，删除即表示下架移除
//   - 通过 for_sale 字段控制是否在售，删除仅用于彻底移除
//   - 已购买的订阅记录保存在 UserSubscription 中，不受套餐删除影响
type SubscriptionPlan struct {
	ent.Schema
}

func (SubscriptionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscription_plans"},
	}
}

func (SubscriptionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("group_id"),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("description").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),
		field.Float("price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Float("original_price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Optional().
			Nillable(),
		field.Int("validity_days").
			Default(30),
		field.String("validity_unit").
			MaxLen(10).
			Default("day"),
		field.String("features").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),
		field.String("product_name").
			MaxLen(100).
			Default(""),
		field.Bool("for_sale").
			Default(true),
		field.Int("sort_order").
			Default(0),
		// 主推标记：admin 在某档套餐打上"⭐主推"后，前端订阅卡片角上展示徽章，
		// 强化用户视觉聚焦在主推档（通常是 Pro 那一档），提升转化。
		// 不影响排序，sort_order 仍然是显示顺序的唯一来源。
		field.Bool("is_popular").
			Default(false),
		field.String("badge_text").
			MaxLen(48).
			Default(""),
		// 套餐角标配色：预设尊贵色板的 key（gold/obsidian/purple/emerald/sapphire/rose）。
		// 仅在 badge_text 非空时生效；默认 gold（鎏金）对齐历史角标的琥珀色。
		field.String("badge_color").
			MaxLen(20).
			Default("gold"),
		// 档位名：订阅页档位对比表的列头显示用（如 Lite/Pro）。留空时前端从套餐名自动推导。
		field.String("plan_label").
			MaxLen(48).
			Default(""),
		// 档位样式：升级阶梯预设 key（basic/standard/advanced/flagship/luxury/supreme），
		// 决定对比表里该档的专属配色 + 图标（越高越豪华）。默认 basic（简约）。
		field.String("tier_style").
			MaxLen(20).
			Default("basic"),
		// 套餐类型：shared = 共享池订阅（用户共享 group 内账号）
		// exclusive = 独享池订阅（购买后从 group 池子里独占分配一个账号）
		field.String("kind").
			MaxLen(20).
			Default("shared"),
		// 套餐级限额/倍率覆盖（migration 138）：NULL = 使用 group 默认值
		// 用于同 group 下不同档位的差异化（Lite/Pro/Max）。购买时会把当前值
		// 快照到 user_subscriptions/exclusive_subscriptions，后续 plan 改不影响已购用户
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
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SubscriptionPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("for_sale"),
	}
}
