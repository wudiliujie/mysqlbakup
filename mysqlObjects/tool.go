package mysqlObjects

import (
	"bytes"
	"fmt"
	"strings"
)

//修复换行
func RepairLine(str string) string {
	str = strings.Replace(str, "\r\n", "^~~~~~~~~~~~~~~^", -1)
	str = strings.Replace(str, "\r", "^~~~~~~~~~~~~~~^", -1)
	str = strings.Replace(str, "\n", "^~~~~~~~~~~~~~~^", -1)
	str = strings.Replace(str, "^~~~~~~~~~~~~~~^", "\r\rn", -1)
	return str
}
func RepairDefiner(sql string, definer string) string {
	sa := strings.Split(definer, "@")
	definer = fmt.Sprintf(" DEFINER=`%v`@`%v`", sa[0], sa[1])
	return strings.Replace(sql, definer, "", -1)
}

func EraseDefiner(input string) string {
	buffer := bytes.Buffer{}
	definer := " DEFINER="
	dIndex := strings.IndexAny(input, definer)
	buffer.WriteString(definer)

	pointAliasReached := false
	point3rdQuoteReached := false

	for i := dIndex + len(definer); i < len(input); i++ {
		if !pointAliasReached {
			if input[i] == '@' {
				pointAliasReached = true
			}

			buffer.WriteByte(input[i])
			continue
		}

		if !point3rdQuoteReached {
			if input[i] == '`' {
				point3rdQuoteReached = true
			}

			buffer.WriteByte(input[i])
			continue
		}

		if input[i] != '`' {
			buffer.WriteByte(input[i])
			continue
		} else {
			buffer.WriteByte(input[i])
			break
		}
	}
	return strings.Replace(input, buffer.String(), "", -1)
}
