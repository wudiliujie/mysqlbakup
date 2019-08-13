package mysqlObjects

import (
	"database/sql"
	"fmt"
	"github.com/wudiliujie/common/db"
	"github.com/wudiliujie/common/log"
)

type MySqlFunctionList struct {
	_lst []*MySqlFunction
}

func NewMySqlFunctionList(_db *sql.DB) *MySqlFunctionList {
	functionList := new(MySqlFunctionList)
	functionList._lst = make([]*MySqlFunction, 0)
	dbname := db.ExecuteScalarStr(_db, "SELECT DATABASE();", 0)
	_sqlShowProcedures := fmt.Sprintf("SHOW FUNCTION STATUS WHERE UPPER(TRIM(Db))= UPPER(TRIM('%v'));", dbname)
	dt, err := db.QueryDataTable(_db, _sqlShowProcedures)
	if err != nil {
		log.Error("NewMySqlFunctionList%v", err)
		dt = db.CreateTable()
	}
	for _, v := range dt.DataRows {
		f := NewMySqlFunction(_db, v.GetStringIdx(1), v.GetStringIdx(3))
		if f != nil {
			functionList._lst = append(functionList._lst, f)
		}
	}
	return functionList
}
