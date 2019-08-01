package mysqlObjects

import (
	"common/db"
	"common/log"
	"database/sql"
	"fmt"
)

type MySqlEventList struct {
	_lst []*MySqlEvent
}

func NewMysqlEventList(_db *sql.DB) *MySqlEventList {
	eventList := new(MySqlEventList)
	eventList._lst = make([]*MySqlEvent, 0)
	dbname := db.ExecuteScalarStr(_db, "SELECT DATABASE();", 0)
	_sqlShowEvents := fmt.Sprintf("SHOW EVENTS WHERE UPPER(TRIM(Db))=UPPER(TRIM('%v'));", dbname)
	dt, err := db.QueryDataTable(_db, _sqlShowEvents)
	if err != nil {
		log.Error("NewMysqlEventList:%v", err)
		return eventList
	}
	for _, row := range dt.DataRows {
		eventList._lst = append(eventList._lst, NewMySqlEvent(_db, row.GetString("_name"), row.GetString("Definer")))
	}
	return eventList
}
