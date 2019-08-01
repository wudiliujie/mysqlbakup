package mysqlObjects

import (
	"database/sql"
	"fmt"
	"github.com/wudiliujie/common/db"
	"github.com/wudiliujie/common/log"
)

type MySqlViewList struct {
	_lst []*MySqlView
}

func NewMySqlViewList(_db *sql.DB) *MySqlViewList {
	viewList := new(MySqlViewList)
	viewList._lst = make([]*MySqlView, 0)
	dbname := db.ExecuteScalarStr(_db, "SELECT DATABASE();", 0)
	_sqlShowViewList := fmt.Sprintf("SHOW FULL TABLES FROM `%v` WHERE Table_type = 'VIEW';", dbname)
	dt, err := db.QueryDataTable(_db, _sqlShowViewList)
	if err != nil {
		log.Error("NewMySqlViewList:%v", err)
		return viewList
	}
	for _, row := range dt.DataRows {
		viewList._lst = append(viewList._lst, NewMySqlView(_db, row.GetStringIdx(0)))
	}
	return viewList
}
