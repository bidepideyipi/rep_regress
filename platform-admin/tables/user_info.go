package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetUserinfoTable(ctx *context.Context) table.Table {

	userInfo := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("user_id", db.Varchar))

	info := userInfo.GetInfo().HideFilterArea()

	info.AddField("用户Id", "user_id", db.Varchar)
	info.AddField("所属集成商Id（必须关联集成商）", "integrator_id", db.Varchar)
	info.AddField("用户名称", "user_name", db.Varchar)
	info.AddField("用户昵称", "nickname", db.Varchar)
	info.AddField("用户邮箱", "email", db.Varchar)
	info.AddField("用户电话", "phone", db.Varchar)
	info.AddField("密码哈希值（加密存储）", "password_hash", db.Varchar)
	info.AddField("Rtp容忍阈值（百分比，0表示完全容忍）", "rtp_tolerance_threshold", db.Decimal)
	info.AddField("账户余额（独立行级锁）", "balance", db.Decimal)
	info.AddField("Vip等级", "vip_level", db.Tinyint)
	info.AddField("状态（0-禁用，1-启用，2-冻结）", "status", db.Tinyint)
	info.AddField("最后登录时间", "last_login_time", db.Datetime)
	info.AddField("最后登录Ip", "last_login_ip", db.Varchar)
	info.AddField("创建时间", "create_time", db.Datetime)
	info.AddField("更新时间", "update_time", db.Datetime)

	info.SetTable("user_info").SetTitle("Userinfo").SetDescription("Userinfo")

	formList := userInfo.GetForm()
	formList.AddField("用户Id", "user_id", db.Varchar, form.Text)
	formList.AddField("所属集成商Id（必须关联集成商）", "integrator_id", db.Varchar, form.Text)
	formList.AddField("用户名称", "user_name", db.Varchar, form.Text)
	formList.AddField("用户昵称", "nickname", db.Varchar, form.Text)
	formList.AddField("用户邮箱", "email", db.Varchar, form.Email)
	formList.AddField("用户电话", "phone", db.Varchar, form.Text)
	formList.AddField("密码哈希值（加密存储）", "password_hash", db.Varchar, form.Text)
	formList.AddField("Rtp容忍阈值（百分比，0表示完全容忍）", "rtp_tolerance_threshold", db.Decimal, form.Text)
	formList.AddField("账户余额（独立行级锁）", "balance", db.Decimal, form.Text)
	formList.AddField("Vip等级", "vip_level", db.Tinyint, form.Number)
	formList.AddField("状态（0-禁用，1-启用，2-冻结）", "status", db.Tinyint, form.Number)
	formList.AddField("最后登录时间", "last_login_time", db.Datetime, form.Datetime)
	formList.AddField("最后登录Ip", "last_login_ip", db.Varchar, form.Text)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)
	formList.AddField("更新时间", "update_time", db.Datetime, form.Datetime)

	formList.SetTable("user_info").SetTitle("Userinfo").SetDescription("Userinfo")

	return userInfo
}
