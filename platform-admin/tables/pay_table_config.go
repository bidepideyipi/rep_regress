package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetPaytableconfigTable(ctx *context.Context) table.Table {

	payTableConfig := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("id", db.Bigint))

	info := payTableConfig.GetInfo().HideFilterArea()

	info.AddField("自增主键", "id", db.Bigint).
		FieldFilterable()
	info.AddField("关联游戏Id", "game_id", db.Varchar)
	info.AddField("赔付线数量", "pay_line_count", db.Int)
	info.AddField("赔付线模式配置", "pay_line_pattern", db.Json)
	info.AddField("创建时间", "create_time", db.Datetime)
	info.AddField("更新时间", "update_time", db.Datetime)

	info.SetTable("pay_table_config").SetTitle("Paytableconfig").SetDescription("Paytableconfig")

	formList := payTableConfig.GetForm()
	formList.AddField("自增主键", "id", db.Bigint, form.Default)
	formList.AddField("关联游戏Id", "game_id", db.Varchar, form.Text)
	formList.AddField("赔付线数量", "pay_line_count", db.Int, form.Number)
	formList.AddField("赔付线模式配置", "pay_line_pattern", db.Json, form.Text)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)
	formList.AddField("更新时间", "update_time", db.Datetime, form.Datetime)

	formList.SetTable("pay_table_config").SetTitle("Paytableconfig").SetDescription("Paytableconfig")

	return payTableConfig
}
