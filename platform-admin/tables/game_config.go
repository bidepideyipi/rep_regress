package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetGameconfigTable(ctx *context.Context) table.Table {

	gameConfig := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("game_id", db.Varchar))

	info := gameConfig.GetInfo().HideFilterArea()

	info.AddField("游戏Id", "game_id", db.Varchar)
	info.AddField("游戏名称（英文）", "game_name", db.Varchar)
	info.AddField("游戏类型（Slot_3x3、Slot_5x3等）", "game_type", db.Varchar)
	info.AddField("卷轴数量", "reel_count", db.Int)
	info.AddField("符号数量", "symbol_count", db.Int)
	info.AddField("最小下注金额", "min_bet", db.Decimal)
	info.AddField("最大下注金额", "max_bet", db.Decimal)
	info.AddField("最小下注线数", "min_lines", db.Int)
	info.AddField("最大下注线数", "max_lines", db.Int)
	info.AddField("目标Rtp值", "rtp_target", db.Decimal)
	info.AddField("波动性等级（Low/Medium/High）", "volatility_level", db.Varchar)
	info.AddField("特殊功能配置（保留Json格式）", "special_features", db.Json)
	info.AddField("状态（0-禁用，1-启用）", "status", db.Tinyint)
	info.AddField("配置版本号", "version", db.Int)
	info.AddField("备注", "remark", db.Text)
	info.AddField("创建时间", "create_time", db.Datetime)
	info.AddField("更新时间", "update_time", db.Datetime)

	info.SetTable("game_config").SetTitle("Gameconfig").SetDescription("Gameconfig")

	formList := gameConfig.GetForm()
	formList.AddField("游戏Id", "game_id", db.Varchar, form.Text)
	formList.AddField("游戏名称（英文）", "game_name", db.Varchar, form.Text)
	formList.AddField("游戏类型（Slot_3x3、Slot_5x3等）", "game_type", db.Varchar, form.Text)
	formList.AddField("卷轴数量", "reel_count", db.Int, form.Number)
	formList.AddField("符号数量", "symbol_count", db.Int, form.Number)
	formList.AddField("最小下注金额", "min_bet", db.Decimal, form.Text)
	formList.AddField("最大下注金额", "max_bet", db.Decimal, form.Text)
	formList.AddField("最小下注线数", "min_lines", db.Int, form.Number)
	formList.AddField("最大下注线数", "max_lines", db.Int, form.Number)
	formList.AddField("目标Rtp值", "rtp_target", db.Decimal, form.Text)
	formList.AddField("波动性等级（Low/Medium/High）", "volatility_level", db.Varchar, form.Text)
	formList.AddField("特殊功能配置（保留Json格式）", "special_features", db.Json, form.Text)
	formList.AddField("状态（0-禁用，1-启用）", "status", db.Tinyint, form.Number)
	formList.AddField("配置版本号", "version", db.Int, form.Number)
	formList.AddField("备注", "remark", db.Text, form.RichText)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)
	formList.AddField("更新时间", "update_time", db.Datetime, form.Datetime)

	formList.SetTable("game_config").SetTitle("Gameconfig").SetDescription("Gameconfig")

	return gameConfig
}
