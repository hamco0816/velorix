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

// DesktopRelease 桌面客户端版本发布记录。
//
// 存储布局：安装包(setup_file)与 blockmap 平铺在更新目录根，文件名含版本号不冲突、永久保留；
// latest.yml 内容存 latest_yml 字段，激活某版本时写到更新目录根的 latest.yml + release.json，
// 供 electron-updater 拉取。同一 channel 同一时刻只有一条 status=active 的记录对外提供更新。
//
// 删除策略：硬删除（同时删磁盘文件）。
type DesktopRelease struct {
	ent.Schema
}

func (DesktopRelease) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "desktop_releases"},
	}
}

func (DesktopRelease) Fields() []ent.Field {
	return []ent.Field{
		field.String("version").
			MaxLen(50).
			NotEmpty().
			Comment("版本号，如 0.2.0"),
		field.String("channel").
			MaxLen(20).
			Default("stable").
			Comment("发布通道: stable, beta"),
		field.Bool("mandatory").
			Default(false).
			Comment("是否强制更新（写入 release.json 供客户端弹不可跳过遮罩）"),
		field.String("notes").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default("").
			Comment("更新说明（展示给用户）"),
		field.String("setup_file").
			MaxLen(255).
			NotEmpty().
			Comment("安装包文件名（相对更新目录根，如 Velorix-0.2.0-setup.exe）"),
		field.String("blockmap_file").
			MaxLen(255).
			Default("").
			Comment("blockmap 文件名（增量更新用，可空）"),
		field.String("latest_yml").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default("").
			Comment("electron-updater 的 latest.yml 全文内容"),
		field.Int64("file_size").
			Default(0).
			Comment("安装包字节大小"),
		field.String("status").
			MaxLen(20).
			Default("active").
			Comment("状态: active(当前对外), archived(历史), rolledback(已回滚)"),
		field.Int64("created_by").
			Optional().
			Nillable().
			Comment("发布人用户ID（管理员）"),
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

func (DesktopRelease) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("version", "channel").Unique(),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
