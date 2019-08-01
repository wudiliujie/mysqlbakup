package mysqlObjects

import (
	"database/sql"
	"fmt"
	"github.com/wudiliujie/common/db"
	"github.com/wudiliujie/common/log"
)

type MySqlProcedureList struct {
	_lst []*MySqlProcedure
}

func NewMySqlProcedureList(_db *sql.DB) *MySqlProcedureList {
	procedureList := new(MySqlProcedureList)
	procedureList._lst = make([]*MySqlProcedure, 0)
	dbname := db.ExecuteScalarStr(_db, "SELECT DATABASE();", 0)
	_sqlShowProcedures := fmt.Sprintf("SHOW PROCEDURE STATUS WHERE UPPER(TRIM(Db))= UPPER(TRIM('%v'));", dbname)
	dt, err := db.QueryDataTable(_db, _sqlShowProcedures)
	if err != nil {
		log.Error("NewMySqlProcedureList%v", err)
		dt = db.CreateTable()
	}
	for _, v := range dt.DataRows {
		p := NewMySqlProcedure(_db, v.GetString("_name"), v.GetString("Definer"))
		if p != nil {
			procedureList._lst = append(procedureList._lst, p)
		}

	}
	return procedureList
}
