package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetReelconfigTable(ctx *context.Context) table.Table {

	reelConfig := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("id", db.Bigint))

	info := reelConfig.GetInfo().HideFilterArea()

	info.AddField("自增主键", "id", db.Bigint).
		FieldFilterable()
	info.AddField("关联游戏Id", "game_id", db.Varchar)
	info.AddField("卷轴索引（1,2,3,4,5）", "reel_index", db.Int)
	info.AddField("卷轴名称（如Reel1, Reel2）", "reel_name", db.Varchar)
	info.AddField("创建时间", "create_time", db.Datetime)

	info.SetTable("reel_config").SetTitle("Reelconfig").SetDescription("Reelconfig")

	formList := reelConfig.GetForm()
	formList.AddField("自增主键", "id", db.Bigint, form.Default)
	formList.AddField("关联游戏Id", "game_id", db.Varchar, form.Text)
	formList.AddField("卷轴索引（1,2,3,4,5）", "reel_index", db.Int, form.Number)
	formList.AddField("卷轴名称（如Reel1, Reel2）", "reel_name", db.Varchar, form.Text)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)

	formList.SetTable("reel_config").SetTitle("Reelconfig").SetDescription("Reelconfig")

	return reelConfig
}
