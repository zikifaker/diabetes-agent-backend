package dao

import (
	"context"
	"diabetes-agent-server/config"
	"fmt"
	"time"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB           *gorm.DB
	MilvusClient *milvusclient.Client
	RedisClient  *redis.Client
)

func init() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Cfg.DB.MySQL.Username,
		config.Cfg.DB.MySQL.Password,
		config.Cfg.DB.MySQL.Host,
		config.Cfg.DB.MySQL.Port,
		config.Cfg.DB.MySQL.DBName,
	)
	DB, err = gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect MySQL: %v", err))
	}
}

func init() {
	var err error
	MilvusClient, err = milvusclient.New(context.Background(), &milvusclient.ClientConfig{
		Address: config.Cfg.Milvus.Endpoint,
		APIKey:  config.Cfg.Milvus.APIKey,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create Milvus client: %v", err))
	}
}

func init() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Cfg.Redis.Host, config.Cfg.Redis.Port),
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}
}
