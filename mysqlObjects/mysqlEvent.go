package mysqlObjects

import (
	"database/sql"
	"fmt"
	"github.com/wudiliujie/common/db"
)

type MySqlEvent struct {
	_name                         string
	_createEventSql               string
	_createEventSqlWithoutDefiner string
}

func NewMySqlEvent(_db *sql.DB, eventName string, definer string) *MySqlEvent {
	event := new(MySqlEvent)
	sql := fmt.Sprintf("SHOW CREATE EVENT `%v`;", eventName)
	event._createEventSql = RepairLine(db.ExecuteScalarStr(_db, sql, 3))
	event._createEventSqlWithoutDefiner = RepairDefiner(event._createEventSql, definer)
	return event
}
