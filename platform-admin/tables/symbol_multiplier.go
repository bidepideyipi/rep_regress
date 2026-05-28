package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetSymbolmultiplierTable(ctx *context.Context) table.Table {

	symbolMultiplier := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("id", db.Bigint))

	info := symbolMultiplier.GetInfo().HideFilterArea()

	info.AddField("自增主键", "id", db.Bigint).
		FieldFilterable()
	info.AddField("关联符号配置Id", "symbol_config_id", db.Bigint)
	info.AddField("连击数（如2、3、4、5）", "match_count", db.Int)
	info.AddField("赔付倍数", "multiplier", db.Decimal)
	info.AddField("是否需要匹配下注线（0-否，1-是）", "is_bet_line", db.Tinyint)
	info.AddField("创建时间", "create_time", db.Datetime)

	info.SetTable("symbol_multiplier").SetTitle("Symbolmultiplier").SetDescription("Symbolmultiplier")

	formList := symbolMultiplier.GetForm()
	formList.AddField("自增主键", "id", db.Bigint, form.Default)
	formList.AddField("关联符号配置Id", "symbol_config_id", db.Bigint, form.Number)
	formList.AddField("连击数（如2、3、4、5）", "match_count", db.Int, form.Number)
	formList.AddField("赔付倍数", "multiplier", db.Decimal, form.Text)
	formList.AddField("是否需要匹配下注线（0-否，1-是）", "is_bet_line", db.Tinyint, form.Number)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)

	formList.SetTable("symbol_multiplier").SetTitle("Symbolmultiplier").SetDescription("Symbolmultiplier")

	return symbolMultiplier
}
