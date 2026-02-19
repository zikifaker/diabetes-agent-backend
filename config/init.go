package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var Cfg Config

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		CORS struct {
			AllowedOrigins []string `yaml:"allowed_origins"`
		} `yaml:"cors"`
		LogLevel string `yaml:"log_level"`
	}
	Client struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	}
	DB struct {
		MySQL DBConfig `yaml:"mysql"`
	} `yaml:"db"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	JWT struct {
		SecretKey string `yaml:"secret_key"`
	} `yaml:"jwt"`
	MCP struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	}
	OSS struct {
		Region          string `yaml:"region"`
		BucketName      string `yaml:"bucket_name"`
		AccessKeyID     string `yaml:"access_key_id"`
		AccessKeySecret string `yaml:"access_key_secret"`
		RoleARN         string `yaml:"role_arn"`
		CustomDomain    string `yaml:"custom_domain"`
	} `yaml:"oss"`
	MQ struct {
		NameServer []string `yaml:"name_server"`
	} `yaml:"mq"`
	Model struct {
		BaseURL string `yaml:"base_url"`
		APIKey  string `yaml:"api_key"`
	} `yaml:"model"`
	Milvus struct {
		Endpoint string `yaml:"endpoint"`
		APIKey   string `yaml:"api_key"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"milvus"`
	Email struct {
		Host      string `yaml:"host"`
		Port      string `yaml:"port"`
		Password  string `yaml:"password"`
		FromEmail string `yaml:"from_email"`
	} `yaml:"email"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
}

func init() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %v", err))
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		panic(fmt.Sprintf("Failed to parse config: %v", err))
	}
}
