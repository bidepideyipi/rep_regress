package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetSymbolspecialpropertyTable(ctx *context.Context) table.Table {

	symbolSpecialProperty := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("id", db.Bigint))

	info := symbolSpecialProperty.GetInfo().HideFilterArea()

	info.AddField("自增主键", "id", db.Bigint).
		FieldFilterable()
	info.AddField("关联符号配置Id", "symbol_config_id", db.Bigint)
	info.AddField("属性名称：Substitute、Free_spins、Any_position等", "property_name", db.Varchar)
	info.AddField("属性值（Json字符串或具体值）", "property_value", db.Varchar)
	info.AddField("创建时间", "create_time", db.Datetime)

	info.SetTable("symbol_special_property").SetTitle("Symbolspecialproperty").SetDescription("Symbolspecialproperty")

	formList := symbolSpecialProperty.GetForm()
	formList.AddField("自增主键", "id", db.Bigint, form.Default)
	formList.AddField("关联符号配置Id", "symbol_config_id", db.Bigint, form.Number)
	formList.AddField("属性名称：Substitute、Free_spins、Any_position等", "property_name", db.Varchar, form.Text)
	formList.AddField("属性值（Json字符串或具体值）", "property_value", db.Varchar, form.Text)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)

	formList.SetTable("symbol_special_property").SetTitle("Symbolspecialproperty").SetDescription("Symbolspecialproperty")

	return symbolSpecialProperty
}
