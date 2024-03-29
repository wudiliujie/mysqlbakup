package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/wudiliujie/common/log"
	"github.com/wudiliujie/mysqlbakup/backup"
)

func main() {
	context, err := sql.Open("mysql", "root:root@tcp(192.168.22.212:3306)/x_game__s1?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("初始化数据库", err)
	}
	col := backup.NewMySqlBackup(context)

	col.ExportToFile("c:/1.sql")
	newdb, err := sql.Open("mysql", "root:root@tcp(192.168.22.212:3306)/test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("初始化数据库", err)
	}

	new := backup.NewMySqlBackup(newdb)
	new.ImportFromFile("c:/1.sql")
	log.Debug("bbb:%v", col)

}
