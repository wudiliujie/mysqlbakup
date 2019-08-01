package db

import (
	"common/convert"
	"common/log"
	"database/sql"
	"time"
)

type DataRow struct {
	Fields map[string]interface{}
}
type DataRow1 struct {
	table  *DataTable
	Fields []interface{}
}

func CreateDataRow1(size int, container []interface{}) *DataRow1 {
	row := new(DataRow1)
	row.Fields = make([]interface{}, size)
	for i := 0; i < size; i++ {
		row.Fields[i] = &container[i]
	}
	return row
}

type DataTable struct {
	TableName string
	Columns   []*sql.ColumnType
	DataRows  []*DataRow1
}

func (d *DataTable) GetColumnsIdx(fieldName string) int {
	for i, v := range d.Columns {
		if v.Name() == fieldName {
			return i
		}
	}
	return 0
}
func (d *DataTable) AddRows(row *DataRow1) {
	row.table = d
	d.DataRows = append(d.DataRows, row)
}
func (d *DataRow1) GetString(fieldName string) string {
	if d.table == nil {
		log.Error("datarow table is nil")
		return ""
	}
	return d.GetStringIdx(d.table.GetColumnsIdx(fieldName))
}
func (d *DataRow1) GetStringIdx(idx int) string {
	v := d.Fields[idx]
	return convert.ToString(v)
}
func (d *DataRow1) GetInt64(fieldName string) int64 {
	if d.table == nil {
		log.Error("datarow table is nil")
		return 0
	}
	return d.GetInt64Idx(d.table.GetColumnsIdx(fieldName))
}
func (d *DataRow1) GetInt64Idx(idx int) int64 {
	v := d.Fields[idx]
	return convert.ToInt64(v)
}

func CreateTable() *DataTable {
	dt := new(DataTable)
	dt.Columns = make([]*sql.ColumnType, 0)
	dt.DataRows = make([]*DataRow1, 0)
	return dt
}

func CreateDataRow(size int) *DataRow {
	row := &DataRow{}
	row.Fields = make(map[string]interface{}, size)
	return row
}
func (d *DataRow) GetInt64(fieldName string) int64 {
	v, ok := d.Fields[fieldName]
	if ok {
		return convert.ToInt64(v)
	} else {
		log.Debug("GetInt64>>>%v>>%v", fieldName, d.Fields)
	}
	return 0
}
func (d *DataRow) GetInt32(fieldName string) int32 {
	v, ok := d.Fields[fieldName]
	if ok {
		return convert.ToInt32(v)
	} else {
		log.Debug("GetInt32>>>%v>>%v", fieldName, d.Fields)
	}
	return 0
}
func (d *DataRow) GetInt32Default(fieldName string, val int32) int32 {
	v, ok := d.Fields[fieldName]
	if ok {
		return convert.ToInt32(v)
	} else {
		log.Debug("GetInt32Default>>>%v>>%v", fieldName, d.Fields)
	}
	return val
}
func (d *DataRow) GetTime(fieldName string) time.Time {
	v, ok := d.Fields[fieldName]
	if ok {
		return convert.ToTime(v)
	} else {
		log.Debug("GetTime>>>%v>>%v", fieldName, d.Fields)
	}
	return time.Now()
}
func (d *DataRow) GetString(fieldName string) string {
	v, ok := d.Fields[fieldName]
	if ok {
		return convert.ToString(v)
	} else {
		log.Debug("GetString>>>%v>>%v", fieldName, d.Fields)
	}
	return ""
}

func (d *DataRow) GetBytes(fieldName string) []byte {
	v, ok := d.Fields[fieldName]
	if ok {
		return convert.ToBytes(v)
	} else {
		log.Debug("GetBytes>>>%v>>%v", fieldName, d.Fields)
	}
	return []byte{}
}

func Query(context *sql.DB, strsql string, args ...interface{}) ([]*DataRow, error) {
	stmt, err := context.Prepare(strsql)
	if err != nil {
		log.Error("mysql query:%v>>%v", strsql, err)
		return nil, err
	}
	rows, err := stmt.Query(args...)
	//log.Debug("mysql query OpenConnections:%v", Context.Stats().OpenConnections)

	if err != nil {
		log.Error("mysql query:%v>>%v", strsql, err)
		return nil, err
	}
	defer func() {
		stmt.Close()
		rows.Close()
	}()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	size := len(columns)
	pts := make([]interface{}, size)
	container := make([]interface{}, size)

	for i := range pts {
		pts[i] = &container[i]
	}
	ret := make([]*DataRow, 0)
	for rows.Next() {
		err = rows.Scan(pts...)
		if err != nil {
			return nil, err
		}
		var r = CreateDataRow(size)
		for i, column := range columns {
			r.Fields[column] = container[i]
		}
		ret = append(ret, r)
	}
	return ret, nil
}

func QueryRow(context *sql.DB, strsql string, args ...interface{}) (*DataRow, error) {
	rows, err := Query(context, strsql, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) <= 0 {
		return nil, sql.ErrNoRows
	}
	return rows[0], nil
}

func QueryDataTable(context *sql.DB, strsql string, args ...interface{}) (*DataTable, error) {
	stmt, err := context.Prepare(strsql)
	if err != nil {
		log.Error("mysql query:%v>>%v", strsql, err)
		return nil, err
	}
	rows, err := stmt.Query(args...)
	//log.Debug("mysql query OpenConnections:%v", Context.Stats().OpenConnections)

	if err != nil {
		log.Error("mysql query:%v>>%v", strsql, err)
		return nil, err
	}
	defer func() {
		stmt.Close()
		rows.Close()
	}()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	size := len(columns)
	dataTable := CreateTable()
	dataTable.Columns, err = rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	pts := make([]interface{}, size)
	container := make([]interface{}, size)

	for i := range pts {
		pts[i] = &container[i]
	}
	for rows.Next() {
		dataRow := CreateDataRow1(size, container)
		err = rows.Scan(pts...)
		if err != nil {
			return nil, err
		}
		for i := 0; i < size; i++ {
			dataRow.Fields[i] = container[i]
		}
		dataTable.AddRows(dataRow)
	}
	return dataTable, nil
}
func ExecuteScalarStr(_db *sql.DB, sql string, columnIndex int) string {
	dt, err := QueryDataTable(_db, sql)
	if err != nil {
		log.Error("ExecuteScalarStr:%v", err)
		return ""
	}

	for _, v := range dt.DataRows {
		return v.GetStringIdx(columnIndex)
	}
	return ""
}
func ExecuteScalarint64(_db *sql.DB, sql string, columnIndex int) int64 {
	dt, err := QueryDataTable(_db, sql)
	if err != nil {
		log.Error("ExecuteScalarStr:%v", err)
		return 0
	}
	for _, v := range dt.DataRows {
		return v.GetInt64Idx(columnIndex)
	}
	return 0
}
