package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetGameconfigversionTable(ctx *context.Context) table.Table {

	gameConfigVersion := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("id", db.Bigint))

	info := gameConfigVersion.GetInfo().HideFilterArea()

	info.AddField("自增主键", "id", db.Bigint).
		FieldFilterable()
	info.AddField("游戏Id", "game_id", db.Varchar)
	info.AddField("版本号", "version", db.Int)
	info.AddField("配置快照Json", "config_snapshot", db.Json)
	info.AddField("变更原因", "change_reason", db.Varchar)
	info.AddField("操作人员", "operator", db.Varchar)
	info.AddField("创建时间", "create_time", db.Datetime)

	info.SetTable("game_config_version").SetTitle("Gameconfigversion").SetDescription("Gameconfigversion")

	formList := gameConfigVersion.GetForm()
	formList.AddField("自增主键", "id", db.Bigint, form.Default)
	formList.AddField("游戏Id", "game_id", db.Varchar, form.Text)
	formList.AddField("版本号", "version", db.Int, form.Number)
	formList.AddField("配置快照Json", "config_snapshot", db.Json, form.Text)
	formList.AddField("变更原因", "change_reason", db.Varchar, form.Text)
	formList.AddField("操作人员", "operator", db.Varchar, form.Text)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)

	formList.SetTable("game_config_version").SetTitle("Gameconfigversion").SetDescription("Gameconfigversion")

	return gameConfigVersion
}
