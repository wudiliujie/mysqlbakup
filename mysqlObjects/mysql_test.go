package mysqlObjects

import (
	"testing"
)

func TestNewMySqlColumnSlow(t *testing.T) {

	_fractionLength := ""
	for _, v := range []byte("9addf1122aa333400asdf") {
		//取出来数值
		if v >= 48 && v <= 57 {
			_fractionLength += string(v)
		}
	}
	t.Log(_fractionLength)
}
