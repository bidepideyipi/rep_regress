package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetUseroverviewTable(ctx *context.Context) table.Table {

	userOverview := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql"))

	info := userOverview.GetInfo().HideFilterArea()

	info.AddField("用户Id", "user_id", db.Varchar)
	info.AddField("所属集成商Id（必须关联集成商）", "integrator_id", db.Varchar)
	info.AddField("集成商名称", "integrator_name", db.Varchar)
	info.AddField("用户名称", "user_name", db.Varchar)
	info.AddField("用户昵称", "nickname", db.Varchar)
	info.AddField("用户邮箱", "email", db.Varchar)
	info.AddField("Rtp容忍阈值（百分比，0表示完全容忍）", "rtp_tolerance_threshold", db.Decimal)
	info.AddField("账户余额（独立行级锁）", "balance", db.Decimal)
	info.AddField("总下注金额", "total_bet", db.Decimal)
	info.AddField("总赔付金额", "total_win", db.Decimal)
	info.AddField("用户Rtp值（总赔付/总下注*100）", "user_rtp", db.Decimal)
	info.AddField("总游戏次数", "total_games", db.Bigint)
	info.AddField("净收益（总赔付-总下注）", "net_result", db.Decimal)
	info.AddField("最大单次赔付", "max_single_win", db.Decimal)
	info.AddField("Vip等级", "vip_level", db.Tinyint)
	info.AddField("状态（0-禁用，1-启用，2-冻结）", "status", db.Tinyint)
	info.AddField("最后登录时间", "last_login_time", db.Datetime)
	info.AddField("创建时间", "create_time", db.Datetime)

	info.SetTable("user_overview").SetTitle("Useroverview").SetDescription("Useroverview")

	formList := userOverview.GetForm()
	formList.AddField("用户Id", "user_id", db.Varchar, form.Text)
	formList.AddField("所属集成商Id（必须关联集成商）", "integrator_id", db.Varchar, form.Text)
	formList.AddField("集成商名称", "integrator_name", db.Varchar, form.Text)
	formList.AddField("用户名称", "user_name", db.Varchar, form.Text)
	formList.AddField("用户昵称", "nickname", db.Varchar, form.Text)
	formList.AddField("用户邮箱", "email", db.Varchar, form.Email)
	formList.AddField("Rtp容忍阈值（百分比，0表示完全容忍）", "rtp_tolerance_threshold", db.Decimal, form.Text)
	formList.AddField("账户余额（独立行级锁）", "balance", db.Decimal, form.Text)
	formList.AddField("总下注金额", "total_bet", db.Decimal, form.Text)
	formList.AddField("总赔付金额", "total_win", db.Decimal, form.Text)
	formList.AddField("用户Rtp值（总赔付/总下注*100）", "user_rtp", db.Decimal, form.Text)
	formList.AddField("总游戏次数", "total_games", db.Bigint, form.Number)
	formList.AddField("净收益（总赔付-总下注）", "net_result", db.Decimal, form.Text)
	formList.AddField("最大单次赔付", "max_single_win", db.Decimal, form.Text)
	formList.AddField("Vip等级", "vip_level", db.Tinyint, form.Number)
	formList.AddField("状态（0-禁用，1-启用，2-冻结）", "status", db.Tinyint, form.Number)
	formList.AddField("最后登录时间", "last_login_time", db.Datetime, form.Datetime)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)

	formList.SetTable("user_overview").SetTitle("Useroverview").SetDescription("Useroverview")

	return userOverview
}
