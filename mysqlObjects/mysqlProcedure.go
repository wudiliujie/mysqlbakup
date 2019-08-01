package mysqlObjects

import (
	"database/sql"
	"fmt"
	"github.com/wudiliujie/common/db"
	"strings"
)

type MySqlProcedure struct {
	_name                             string
	_createProcedureSQL               string
	_createProcedureSQLWithoutDefiner string
}

func NewMySqlProcedure(_db *sql.DB, procedureName string, definer string) *MySqlProcedure {
	procedure := new(MySqlProcedure)
	procedure._name = procedureName

	sql := fmt.Sprintf("SHOW CREATE PROCEDURE `%v`;", procedureName)
	procedure._createProcedureSQL = RepairLine(db.ExecuteScalarStr(_db, sql, 2))

	sa := strings.Split(definer, "@")
	definer = fmt.Sprintf(" DEFINER=`%v`@`%v`", sa[0], sa[1])
	procedure._createProcedureSQLWithoutDefiner = strings.Replace(procedure._createProcedureSQL, definer, "", -1)
	return procedure
}
