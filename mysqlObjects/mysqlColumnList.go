package mysqlObjects

import (
	"common/db"
	"common/log"
	"database/sql"
	"fmt"
)

type MySqlColumnList struct {
	_tableName          string
	_lst                []*MySqlColumn
	_sqlShowFullColumns string
}

func (c *MySqlColumnList) Count() int {
	return len(c._lst)
}
func (c *MySqlColumnList) GetColumns() []*MySqlColumn {
	return c._lst
}

func NewMySqlColumnList(_db *sql.DB, tableName string) *MySqlColumnList {
	_list := new(MySqlColumnList)
	_list._tableName = tableName
	dtDataType, err := db.QueryDataTable(_db, "select * from "+tableName+" where 1=2; ")
	if err != nil {
		log.Error("dtDataType :%v", err)
		dtDataType = db.CreateTable()
		return nil
	}
	_sqlShowFullColumns := fmt.Sprintf("SHOW FULL COLUMNS FROM `%v`;", tableName)
	dtColInfo, err := db.QueryDataTable(_db, _sqlShowFullColumns)
	if err != nil {
		log.Error("dtDataType :%v", err)

		return nil
	}
	for _, v := range dtDataType.Columns {
		isNull, _ := v.Nullable()
		col := NewMySqlColumn(v.Name(),
			v.ScanType(),
			v.DatabaseTypeName(),
			dtColInfo.DataRows[0].GetString("Collation"),
			isNull,
			dtColInfo.DataRows[0].GetString("Key"),
			dtColInfo.DataRows[0].GetString("Default"),
			dtColInfo.DataRows[0].GetString("Extra"),
			dtColInfo.DataRows[0].GetString("Privileges"),
			dtColInfo.DataRows[0].GetString("Comment"),
		)
		_list._lst = append(_list._lst, col)
	}
	return _list
}
