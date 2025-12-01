package dao

import (
	"diabetes-agent-backend/config"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	dbConfig := config.Cfg.DB.MySQL

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect database: %v", err))
	}
}
