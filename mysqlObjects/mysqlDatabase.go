package mysqlObjects

import (
	"common/db"
	"common/log"
	"database/sql"
	"fmt"
	"github.com/wudiliujie/mysqlbakup/infoObjects"
	"strings"
)

type MySqlDatabase struct {
	_name              string
	_createDatabaseSql string
	_dropDatabaseSql   string
	_defaultCharSet    string
	_listTable         *MySqlTableList
	_listProcedure     *MySqlProcedureList
	_listFunction      *MySqlFunctionList
	_listEvent         *MySqlEventList
	_listView          *MySqlViewList
	_listTrigger       *MySqlTriggerList
}

func (d *MySqlDatabase) GetDatabaseInfo(_db *sql.DB, enumGetTotalRowsMode infoObjects.GetTotalRowsMethod) {
	d._name = db.ExecuteScalarStr(_db, "SELECT DATABASE();", 0)
	d._defaultCharSet = db.ExecuteScalarStr(_db, "SHOW VARIABLES LIKE 'character_set_database';", 1)
	d._createDatabaseSql = strings.Replace(db.ExecuteScalarStr(_db, fmt.Sprintf("SHOW CREATE DATABASE `%v`;", d._name), 1), "CREATE DATABASE", "CREATE DATABASE IF NOT EXISTS", -1) + ";"
	d._dropDatabaseSql = fmt.Sprintf("DROP DATABASE IF EXISTS `%v`;", d._name)
	d._listTable = NewMysqlTableList(_db)
	d._listProcedure = NewMySqlProcedureList(_db)
	d._listFunction = NewMySqlFunctionList(_db)
	d._listTrigger = NewMySqlTriggerList(_db)
	d._listEvent = NewMysqlEventList(_db)
	d._listView = NewMySqlViewList(_db)
	if enumGetTotalRowsMode != infoObjects.GetTotalRowsMethod_Skip {
		d.GetTotalRows(_db, enumGetTotalRowsMode)
	}
}
func (d *MySqlDatabase) GetTotalRows(_db *sql.DB, enumGetTotalRowsMode infoObjects.GetTotalRowsMethod) {
	if enumGetTotalRowsMode == infoObjects.GetTotalRowsMethod_InformationSchema {
		dtTotalRows, err := db.QueryDataTable(_db, fmt.Sprintf("SELECT TABLE_NAME, TABLE_ROWS FROM `information_schema`.`tables` WHERE `table_schema` = '%v';", d._name))
		if err != nil {
			log.Error("GetTotalRows a %v", err)
			return
		}
		for _, row := range dtTotalRows.DataRows {
			tbName := row.GetString("TABLE_NAME")
			totalRowsThisTable := row.GetInt64("TABLE_ROWS")
			if d._listTable.Contains(tbName) {
				d._listTable.GetTable(tbName).SetTotalRows(totalRowsThisTable)
			}
		}
	} else if enumGetTotalRowsMode == infoObjects.GetTotalRowsMethod_SelectCount {
		for _, v := range d._listTable._lst {
			v.GetTotalRowsByCounting(_db)
		}
	}
}
func (d *MySqlDatabase) GetTables() []*MySqlTable {
	return d._listTable._lst
}
func (d *MySqlDatabase) GetFunctions() []*MySqlFunction {
	return d._listFunction._lst
}

func (d *MySqlDatabase) GetTable(name string) *MySqlTable {
	for _, v := range d._listTable._lst {
		if v._name == name {
			return v
		}
	}
	return nil
}
func (d *MySqlDatabase) GetName() string {
	return d._name
}
func (d *MySqlDatabase) GetCreateDatabaseSQL() string {
	return d._createDatabaseSql
}
func NewMySqlDatabase() *MySqlDatabase {
	return new(MySqlDatabase)
}
