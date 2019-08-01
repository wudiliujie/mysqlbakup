package mysqlObjects

import (
	"common/db"
	"database/sql"
	"fmt"
)

type MySqlTrigger struct {
	_name                           string
	_createTriggerSQL               string
	_createTriggerSQLWithoutDefiner string
}

func NewMySqlTrigger(_db *sql.DB, triggerName string, definer string) *MySqlTrigger {
	trigger := new(MySqlTrigger)
	trigger._name = triggerName
	sql := fmt.Sprintf("SHOW CREATE TRIGGER `%v`;", trigger)
	trigger._createTriggerSQL = db.ExecuteScalarStr(_db, sql, 2)
	trigger._createTriggerSQL = RepairLine(trigger._createTriggerSQL)
	trigger._createTriggerSQLWithoutDefiner = RepairDefiner(trigger._createTriggerSQL, definer)
	return trigger
}
