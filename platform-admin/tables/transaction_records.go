package main

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func GetTransactionrecordsTable(ctx *context.Context) table.Table {

	transactionRecords := table.NewDefaultTable(ctx, table.DefaultConfigWithDriver("mysql").SetPrimaryKey("id", db.Bigint))

	info := transactionRecords.GetInfo().HideFilterArea()

	info.AddField("自增主键", "id", db.Bigint).
		FieldFilterable()
	info.AddField("交易Id", "transaction_id", db.Varchar)
	info.AddField("用户Id", "user_id", db.Varchar)
	info.AddField("集成商Id", "integrator_id", db.Varchar)
	info.AddField("交易类型（Bet/Win/Refund等）", "transaction_type", db.Varchar)
	info.AddField("变动金额", "amount", db.Decimal)
	info.AddField("交易前余额", "balance_before", db.Decimal)
	info.AddField("交易后余额", "balance_after", db.Decimal)
	info.AddField("关联游戏会话Id", "game_session_id", db.Varchar)
	info.AddField("关联参考Id", "reference_id", db.Varchar)
	info.AddField("交易状态（0-失败，1-成功，2-处理中）", "status", db.Tinyint)
	info.AddField("错误信息", "error_message", db.Varchar)
	info.AddField("创建时间", "create_time", db.Datetime)

	info.SetTable("transaction_records").SetTitle("Transactionrecords").SetDescription("Transactionrecords")

	formList := transactionRecords.GetForm()
	formList.AddField("自增主键", "id", db.Bigint, form.Default)
	formList.AddField("交易Id", "transaction_id", db.Varchar, form.Text)
	formList.AddField("用户Id", "user_id", db.Varchar, form.Text)
	formList.AddField("集成商Id", "integrator_id", db.Varchar, form.Text)
	formList.AddField("交易类型（Bet/Win/Refund等）", "transaction_type", db.Varchar, form.Text)
	formList.AddField("变动金额", "amount", db.Decimal, form.Text)
	formList.AddField("交易前余额", "balance_before", db.Decimal, form.Text)
	formList.AddField("交易后余额", "balance_after", db.Decimal, form.Text)
	formList.AddField("关联游戏会话Id", "game_session_id", db.Varchar, form.Text)
	formList.AddField("关联参考Id", "reference_id", db.Varchar, form.Text)
	formList.AddField("交易状态（0-失败，1-成功，2-处理中）", "status", db.Tinyint, form.Number)
	formList.AddField("错误信息", "error_message", db.Varchar, form.Text)
	formList.AddField("创建时间", "create_time", db.Datetime, form.Datetime)

	formList.SetTable("transaction_records").SetTitle("Transactionrecords").SetDescription("Transactionrecords")

	return transactionRecords
}
