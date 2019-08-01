package mysqlObjects

import (
	"database/sql"
	"fmt"
	"github.com/wudiliujie/common/db"
	"strings"
)

type MySqlFunction struct {
	_name                            string
	_createFunctionSQL               string
	_createFunctionSqlWithoutDefiner string
}

func (f *MySqlFunction) GetCreateFunctionSQL() string {
	return f._createFunctionSQL
}
func (f *MySqlFunction) GetCreateFunctionSqlWithoutDefiner() string {
	return f._createFunctionSqlWithoutDefiner
}
func (f *MySqlFunction) GetName() string {
	return f._name
}

func NewMySqlFunction(_db *sql.DB, functionName string, definer string) *MySqlFunction {
	function := new(MySqlFunction)
	function._name = functionName
	sql := fmt.Sprintf("SHOW CREATE FUNCTION `%v`;", functionName)
	function._createFunctionSQL = RepairLine(db.ExecuteScalarStr(_db, sql, 2))

	sa := strings.Split(definer, "@")
	definer = fmt.Sprintf(" DEFINER=`%v`@`%v`", sa[0], sa[1])
	function._createFunctionSqlWithoutDefiner = RepairDefiner(function._createFunctionSQL, definer)

	return function
}
