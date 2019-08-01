package mysqlObjects

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/wudiliujie/common/char"
	"github.com/wudiliujie/common/db"
	"strings"
)

//换行分割
const LineSpilt = "^~~~~~~^"

type MySqlTable struct {
	_name                                string
	_lst                                 *MySqlColumnList
	_totalRows                           int64
	_createTableSql                      string
	_createTableSqlWithoutAutoIncrement  string
	_insertStatementHeader               string
	_insertStatementHeaderWithoutColumns string
}

func NewMysqlTable(_db *sql.DB, tableName string) *MySqlTable {
	table := new(MySqlTable)
	table._name = tableName
	sql := fmt.Sprintf("SHOW CREATE TABLE `%v`;", tableName)
	table._createTableSql = db.ExecuteScalarStr(_db, sql, 1)
	table._createTableSql = strings.Replace(table._createTableSql, "\r", LineSpilt, -1)
	table._createTableSql = strings.Replace(table._createTableSql, "\n", LineSpilt, -1)
	table._createTableSql = strings.Replace(table._createTableSql, LineSpilt, "\r\n", -1)
	table._createTableSql = strings.Replace(table._createTableSql, "CREATE TABLE ", "CREATE TABLE IF NOT EXISTS ", -1) + ";"
	table._createTableSqlWithoutAutoIncrement = removeAutoIncrement(table._createTableSql)
	table._lst = NewMySqlColumnList(_db, tableName)
	table.getInsertStatementHeaders()
	return table
}
func removeAutoIncrement(sql string) string {
	a := "AUTO_INCREMENT="

	if strings.Contains(sql, a) {
		i := strings.LastIndex(sql, a)
		b := i + len(a)
		d := ""
		count := 0

		data := []byte(sql)
		for {
			if char.IsDigit(data[b+count]) {
				d = d + string(data[b+count])
			} else {
				break
			}
			count++
		}
		sql = strings.Replace(sql, a+d, "", -1)
	}
	return sql
}
func (t *MySqlTable) getInsertStatementHeaders() {
	t._insertStatementHeaderWithoutColumns = fmt.Sprintf("INSERT INTO `%v` VALUES", t._name)
	var buffer bytes.Buffer
	buffer.WriteString("INSERT INTO `")
	buffer.WriteString(t._name)
	buffer.WriteString("` (")
	for i, v := range t._lst.GetColumns() {
		if i > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString("`")
		buffer.WriteString(v._name)
		buffer.WriteString("`")
	}
	buffer.WriteString(") VALUES")
	t._insertStatementHeader = buffer.String()

}
func (t *MySqlTable) SetTotalRows(count int64) {
	t._totalRows = count
}
func (t *MySqlTable) GetTotalRowsByCounting(_db *sql.DB) {
	sql := fmt.Sprintf("SELECT COUNT(*) FROM `%v`;", t._name)
	t._totalRows = db.ExecuteScalarint64(_db, sql, 0)
}
func (t *MySqlTable) GetName() string {
	return t._name
}
func (t *MySqlTable) GetCreateTableSql() string {
	return t._createTableSql
}
func (t *MySqlTable) GetTotalRows() int64 {
	return t._totalRows
}
func (t *MySqlTable) GetColumns(name string) *MySqlColumn {
	for _, v := range t._lst._lst {
		if v.GetName() == name {
			return v
		}
	}
	return nil
}
