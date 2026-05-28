package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetUsersessionTable(ctx *context.Context) table.Table {

	userSession := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("session_id", db.Varchar))

	info := userSession.GetInfo().HideFilterArea()

	info.AddField("会话Id", "session_id", db.Varchar)
	info.AddField("用户Id", "user_id", db.Varchar)
	info.AddField("集成商Id", "integrator_id", db.Varchar)
	info.AddField("访问令牌", "access_token", db.Varchar)
	info.AddField("加密游戏地址", "game_url", db.Varchar)
	info.AddField("过期时间", "expire_time", db.Datetime)
	info.AddField("当前游戏Id", "current_game_id", db.Varchar)
	info.AddField("会话总下注金额", "total_bet_amount", db.Decimal)
	info.AddField("会话总赔付金额", "total_win_amount", db.Decimal)
	info.AddField("会话持续时长（秒）", "session_duration", db.Int)
	info.AddField("会话状态（0-结束，1-活跃，2-过期）", "status", db.Tinyint)
	info.AddField("客户端Ip地址", "client_ip", db.Varchar)
	info.AddField("用户代理信息", "user_agent", db.Varchar)
	info.AddField("创建时间", "create_time", db.Datetime)
	info.AddField("更新时间", "update_time", db.Datetime)

	info.SetTable("user_session").SetTitle("Usersession").SetDescription("Usersession")

	formList := userSession.GetForm()
	formList.AddField("会话Id", "session_id", db.Varchar, form.Text)
	formList.AddField("用户Id", "user_id", db.Varchar, form.Text)
	formList.AddField("集成商Id", "integrator_id", db.Varchar, form.Text)
	formList.AddField("访问令牌", "access_token", db.Varchar, form.Text)
	formList.AddField("加密游戏地址", "game_url", db.Varchar, form.Text)
	formList.AddField("过期时间", "expire_time", db.Datetime, form.Datetime)
	formList.AddField("当前游戏Id", "current_game_id", db.Varchar, form.Text)
	formList.AddField("会话总下注金额", "total_bet_amount", db.Decimal, form.Text)
	formList.AddField("会话总赔付金额", "total_win_amount", db.Decimal, form.Text)
	formList.AddField("会话持续时长（秒）", "session_duration", db.Int, form.Number)
	formList.AddField("会话状态（0-结束，1-活跃，2-过期）", "status", db.Tinyint, form.Number)
	formList.AddField("客户端Ip地址", "client_ip", db.Varchar, form.Text)
	formList.AddField("用户代理信息", "user_agent", db.Varchar, form.Text)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)
	formList.AddField("更新时间", "update_time", db.Datetime, form.Datetime)

	formList.SetTable("user_session").SetTitle("Usersession").SetDescription("Usersession")

	return userSession
}
