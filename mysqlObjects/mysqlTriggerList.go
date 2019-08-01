package mysqlObjects

import (
	"database/sql"
	"github.com/wudiliujie/common/db"
	"github.com/wudiliujie/common/log"
)

type MySqlTriggerList struct {
	_lst []*MySqlTrigger
}

func NewMySqlTriggerList(_db *sql.DB) *MySqlTriggerList {
	triggerList := new(MySqlTriggerList)
	triggerList._lst = make([]*MySqlTrigger, 0)
	dt, err := db.QueryDataTable(_db, "SHOW TRIGGERS;")
	if err != nil {
		log.Error("NewMySqlTriggerList:%v", err)
		return triggerList
	}
	for _, row := range dt.DataRows {
		triggerList._lst = append(triggerList._lst, NewMySqlTrigger(_db, row.GetString("Trigger"), row.GetString("Definer")))
	}
	return triggerList
}
