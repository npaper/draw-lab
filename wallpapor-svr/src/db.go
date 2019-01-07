package main

import (
	"fmt"
	"model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx" //初始化一个mysql驱动，必须
)

var dbConfig map[string]interface{}

// InitDB ...
func InitDB(config map[string]interface{}) {
	dbConfig = config
	model.CheckWallpaperTable(GetConn())
}

// GetConn ...
func GetConn() *sqlx.DB {
	db, err := sqlx.Open("mysql", dbConfig["user"].(string)+":"+dbConfig["password"].(string)+"@tcp("+dbConfig["url"].(string)+")/"+dbConfig["database"].(string)+"?charset=utf8mb4")
	if err != nil {
		fmt.Println("open mysql failed,", err)
		return nil
	}
	return db
}
