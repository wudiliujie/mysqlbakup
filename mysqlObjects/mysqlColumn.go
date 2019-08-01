package mysqlObjects

import (
	"common/char"
	"common/convert"
	"reflect"
	"strings"
)

type MySqlColumn struct {
	_name              string
	DataType           reflect.Type
	MySqlDataType      string
	Collation          string //排序规则
	AllowNull          bool   //是否允许为空
	Key                string
	DefaultValue       string //默认值
	Extra              string
	Privileges         string //特殊
	Comment            string //注释
	IsPrimaryKey       bool
	TimeFractionLength int
	_isGeneratedColumn bool //是否是生成列
}

func (c *MySqlColumn) GetName() string {
	return c._name
}
func (c *MySqlColumn) GetIsGeneratedColumn() bool {
	return c._isGeneratedColumn
}

func NewMySqlColumn(name string, _type reflect.Type, mySqlDataType string,
	collation string, allowNull bool, key string, defaultValue string,
	extra string, privileges string, comment string) *MySqlColumn {
	column := new(MySqlColumn)
	column._name = name
	column.DataType = _type
	column.MySqlDataType = strings.ToLower(mySqlDataType)
	column.Collation = collation
	column.AllowNull = allowNull
	column.Key = key
	column.DefaultValue = defaultValue
	column.Extra = extra
	column.Privileges = privileges
	column.Comment = comment

	if strings.ToLower(key) == "pri" {
		column.IsPrimaryKey = true
	}
	if _type.Name() == "time.Time" {
		if len(column.MySqlDataType) > 8 {
			_fractionLength := ""
			for _, v := range []byte(column.MySqlDataType) {
				//取出来数值
				if char.IsDigit(v) {
					_fractionLength += string(v)
				}
			}
			if len(_fractionLength) > 0 {
				column.TimeFractionLength = int(convert.ToInt32(_fractionLength))
			}
		}
	}
	if strings.Index(strings.ToLower(extra), "generated") >= 0 {
		column._isGeneratedColumn = true
	} else {
		column._isGeneratedColumn = false
	}
	return column
}
