package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetReelsymbolweightTable(ctx *context.Context) table.Table {

	reelSymbolWeight := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("id", db.Bigint))

	info := reelSymbolWeight.GetInfo().HideFilterArea()

	info.AddField("自增主键", "id", db.Bigint).
		FieldFilterable()
	info.AddField("关联卷轴配置Id", "reel_config_id", db.Bigint)
	info.AddField("关联符号配置Id", "symbol_config_id", db.Bigint)
	info.AddField("符号权重（权重越大出现概率越高）", "weight", db.Int)
	info.AddField("创建时间", "create_time", db.Datetime)

	info.SetTable("reel_symbol_weight").SetTitle("Reelsymbolweight").SetDescription("Reelsymbolweight")

	formList := reelSymbolWeight.GetForm()
	formList.AddField("自增主键", "id", db.Bigint, form.Default)
	formList.AddField("关联卷轴配置Id", "reel_config_id", db.Bigint, form.Number)
	formList.AddField("关联符号配置Id", "symbol_config_id", db.Bigint, form.Number)
	formList.AddField("符号权重（权重越大出现概率越高）", "weight", db.Int, form.Number)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)

	formList.SetTable("reel_symbol_weight").SetTitle("Reelsymbolweight").SetDescription("Reelsymbolweight")

	return reelSymbolWeight
}
