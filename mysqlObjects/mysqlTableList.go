package mysqlObjects

import (
	"common/db"
	"common/log"
	"database/sql"
)

type MySqlTableList struct {
	_lst []*MySqlTable
}

func (t *MySqlTableList) Contains(name string) bool {
	for _, v := range t._lst {
		if v._name == name {
			return true
		}
	}
	return false
}
func (t *MySqlTableList) GetTable(name string) *MySqlTable {
	for _, v := range t._lst {
		if v._name == name {
			return v
		}
	}
	return nil
}

func NewMysqlTableList(_db *sql.DB) *MySqlTableList {
	tableList := new(MySqlTableList)
	_sqlShowFullTables := "SHOW FULL TABLES WHERE Table_type = 'BASE TABLE';"
	dtTableList, err := db.QueryDataTable(_db, _sqlShowFullTables)
	if err != nil {
		log.Error("NewMysqlTableList:%v", err)
		dtTableList = db.CreateTable()
	}
	for _, v := range dtTableList.DataRows {
		table := NewMysqlTable(_db, v.GetStringIdx(0))
		if table != nil {
			tableList._lst = append(tableList._lst, table)
		}
	}
	return tableList
}
