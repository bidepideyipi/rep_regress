package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetIntegratorconfigTable(ctx *context.Context) table.Table {

	integratorConfig := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("integrator_id", db.Varchar))

	info := integratorConfig.GetInfo().HideFilterArea()

	info.AddField("集成商Id", "integrator_id", db.Varchar)
	info.AddField("集成商名称", "integrator_name", db.Varchar)
	info.AddField("公司名称", "company_name", db.Varchar)
	info.AddField("联系人", "contact_person", db.Varchar)
	info.AddField("联系邮箱", "contact_email", db.Varchar)
	info.AddField("联系电话", "contact_phone", db.Varchar)
	info.AddField("是否平台自营（0-否，1-是）", "is_platform_self", db.Tinyint)
	info.AddField("Api密钥", "api_key", db.Varchar)
	info.AddField("Api密钥（加密存储）", "api_secret", db.Varchar)
	info.AddField("游戏访问权限列表", "game_access_permitted", db.Json)
	info.AddField("财务权限（0-否，1-是）", "financial_permission", db.Tinyint)
	info.AddField("状态（0-禁用，1-启用）", "status", db.Tinyint)
	info.AddField("备注", "remark", db.Text)
	info.AddField("创建时间", "create_time", db.Datetime)
	info.AddField("更新时间", "update_time", db.Datetime)

	info.SetTable("integrator_config").SetTitle("Integratorconfig").SetDescription("Integratorconfig")

	formList := integratorConfig.GetForm()
	formList.AddField("集成商Id", "integrator_id", db.Varchar, form.Text)
	formList.AddField("集成商名称", "integrator_name", db.Varchar, form.Text)
	formList.AddField("公司名称", "company_name", db.Varchar, form.Text)
	formList.AddField("联系人", "contact_person", db.Varchar, form.Text)
	formList.AddField("联系邮箱", "contact_email", db.Varchar, form.Text)
	formList.AddField("联系电话", "contact_phone", db.Varchar, form.Text)
	formList.AddField("是否平台自营（0-否，1-是）", "is_platform_self", db.Tinyint, form.Number)
	formList.AddField("Api密钥", "api_key", db.Varchar, form.Text)
	formList.AddField("Api密钥（加密存储）", "api_secret", db.Varchar, form.Text)
	formList.AddField("游戏访问权限列表", "game_access_permitted", db.Json, form.Text)
	formList.AddField("财务权限（0-否，1-是）", "financial_permission", db.Tinyint, form.Number)
	formList.AddField("状态（0-禁用，1-启用）", "status", db.Tinyint, form.Number)
	formList.AddField("备注", "remark", db.Text, form.RichText)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)
	formList.AddField("更新时间", "update_time", db.Datetime, form.Datetime)

	formList.SetTable("integrator_config").SetTitle("Integratorconfig").SetDescription("Integratorconfig")

	return integratorConfig
}
