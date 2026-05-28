package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetSymbolconfigTable(ctx *context.Context) table.Table {

	symbolConfig := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("id", db.Bigint))

	info := symbolConfig.GetInfo().HideFilterArea()

	info.AddField("自增主键", "id", db.Bigint).
		FieldFilterable()
	info.AddField("关联游戏Id", "game_id", db.Varchar)
	info.AddField("符号Id（英文，如Cherry）", "symbol_id", db.Varchar)
	info.AddField("符号名称（英文，如Cherry）", "symbol_name", db.Varchar)
	info.AddField("符号类型：Normal、Wild、Scatter", "symbol_type", db.Varchar)
	info.AddField("符号描述", "description", db.Varchar)
	info.AddField("是否启用（0-否，1-是）", "is_active", db.Tinyint)
	info.AddField("排序顺序", "sort_order", db.Int)
	info.AddField("创建时间", "create_time", db.Datetime)
	info.AddField("更新时间", "update_time", db.Datetime)

	info.SetTable("symbol_config").SetTitle("Symbolconfig").SetDescription("Symbolconfig")

	formList := symbolConfig.GetForm()
	formList.AddField("自增主键", "id", db.Bigint, form.Default)
	formList.AddField("关联游戏Id", "game_id", db.Varchar, form.Text)
	formList.AddField("符号Id（英文，如Cherry）", "symbol_id", db.Varchar, form.Text)
	formList.AddField("符号名称（英文，如Cherry）", "symbol_name", db.Varchar, form.Text)
	formList.AddField("符号类型：Normal、Wild、Scatter", "symbol_type", db.Varchar, form.Text)
	formList.AddField("符号描述", "description", db.Varchar, form.Text)
	formList.AddField("是否启用（0-否，1-是）", "is_active", db.Tinyint, form.Number)
	formList.AddField("排序顺序", "sort_order", db.Int, form.Number)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)
	formList.AddField("更新时间", "update_time", db.Datetime, form.Datetime)

	formList.SetTable("symbol_config").SetTitle("Symbolconfig").SetDescription("Symbolconfig")

	return symbolConfig
}
