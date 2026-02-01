package dao

import (
	"context"
	"diabetes-agent-server/config"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB           *gorm.DB
	MilvusClient *milvusclient.Client
)

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

func init() {
	milvusConfig := milvusclient.ClientConfig{
		Address: config.Cfg.Milvus.Endpoint,
		APIKey:  config.Cfg.Milvus.APIKey,
	}

	var err error
	MilvusClient, err = milvusclient.New(context.Background(), &milvusConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Milvus client: %v", err))
	}
}
