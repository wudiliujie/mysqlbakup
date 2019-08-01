package mysqlObjects

import (
	"common/db"
	"database/sql"
	"fmt"
)

type MySqlView struct {
	_name                        string
	_createViewSQL               string
	_createViewSQLWithoutDefiner string
}

func NewMySqlView(_db *sql.DB, viewName string) *MySqlView {
	view := new(MySqlView)
	view._name = viewName
	sql := fmt.Sprintf("SHOW CREATE VIEW `%v`;", viewName)

	view._createViewSQL = RepairLine(db.ExecuteScalarStr(_db, sql, 1))
	view._createViewSQLWithoutDefiner = EraseDefiner(view._createViewSQL)
	return view
}
