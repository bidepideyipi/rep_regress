package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetUserfinancestatsTable(ctx *context.Context) table.Table {

	userFinanceStats := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("id", db.Bigint))

	info := userFinanceStats.GetInfo().HideFilterArea()

	info.AddField("自增主键", "id", db.Bigint).
		FieldFilterable()
	info.AddField("关联用户Id", "user_id", db.Varchar)
	info.AddField("总下注金额", "total_bet", db.Decimal)
	info.AddField("总赔付金额", "total_win", db.Decimal)
	info.AddField("总游戏次数", "total_games", db.Bigint)
	info.AddField("净收益（总赔付-总下注）", "net_result", db.Decimal)
	info.AddField("用户Rtp值（总赔付/总下注*100）", "rtp_value", db.Decimal)
	info.AddField("最大单次赔付", "max_single_win", db.Decimal)
	info.AddField("更新时间", "update_time", db.Datetime)

	info.SetTable("user_finance_stats").SetTitle("Userfinancestats").SetDescription("Userfinancestats")

	formList := userFinanceStats.GetForm()
	formList.AddField("自增主键", "id", db.Bigint, form.Default)
	formList.AddField("关联用户Id", "user_id", db.Varchar, form.Text)
	formList.AddField("总下注金额", "total_bet", db.Decimal, form.Text)
	formList.AddField("总赔付金额", "total_win", db.Decimal, form.Text)
	formList.AddField("总游戏次数", "total_games", db.Bigint, form.Number)
	formList.AddField("净收益（总赔付-总下注）", "net_result", db.Decimal, form.Text)
	formList.AddField("用户Rtp值（总赔付/总下注*100）", "rtp_value", db.Decimal, form.Text)
	formList.AddField("最大单次赔付", "max_single_win", db.Decimal, form.Text)
	formList.AddField("更新时间", "update_time", db.Datetime, form.Datetime)

	formList.SetTable("user_finance_stats").SetTitle("Userfinancestats").SetDescription("Userfinancestats")

	return userFinanceStats
}
