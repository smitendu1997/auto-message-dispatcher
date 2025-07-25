package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/smitendu1997/auto-message-dispatcher/logger"
	"github.com/spf13/viper"
)

// AppConfig defines application configuration
type AppConfig struct {
	Environment     string
	MessagingApiUrl string
	MessagingApiKey string
	Database        DBConfig
	Redis           RedisConfig
}

// DBConfig holds database configuration
type DBConfig struct {
	Type     string
	Host     string
	Port     string
	Username string
	Password string
	Database string
	DSN      string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Type     string
	Host     string
	Port     string
	UserName string
	Password string
	Db       string
	URL      string
}

// LoadConfig initializes the configuration for the service.
func LoadConfig() *AppConfig {
	// If env is empty, use environment variable or default

	env := viper.GetString("APP_ENV")
	if env == "" {
		env = "production"
	}

	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file2")
	flag.Parse()

	viper.AutomaticEnv()

	// if this call is a local call, then get the data from the local config file and set the values in viper
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		configPath = os.Getenv("CONFIG_PATH")
		if configPath != "" {
			viper.AddConfigPath(configPath)
		}
		_, b, _, _ := runtime.Caller(0)
		basePath := filepath.Dir(b)

		logger.Info("BasePath", basePath)
		viper.AddConfigPath(filepath.Join(basePath)) // Look for .env in project root/config
		viper.AddConfigPath(".")
		viper.SetConfigName(".env")
		viper.SetConfigType("env")
	}

	// Set defaults
	viper.SetDefault("DATABASE_TYPE", "mysql")
	viper.SetDefault("REDIS_TYPE", "redis")

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, use environment variables only
		logger.Error("Config file not found, using environment variables only: %v", err)
	}

	// Create config instance
	config := &AppConfig{
		Environment: env,
		Database: DBConfig{
			Type:     viper.GetString("DATABASE_TYPE"),
			Host:     viper.GetString("MYSQL_DB_HOST"),
			Port:     viper.GetString("MYSQL_DB_PORT"),
			Username: viper.GetString("MYSQL_DB_USERNAME"),
			Password: viper.GetString("MYSQL_DB_PASSWORD"),
			Database: viper.GetString("MYSQL_DB_SCHEMA"),
		},
		Redis: RedisConfig{
			Type:     viper.GetString("REDIS_TYPE"),
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			UserName: viper.GetString("REDIS_USERNAME"),
			Password: viper.GetString("REDIS_PASSWORD"),
			Db:       viper.GetString("REDIS_DB"),
		},
		MessagingApiUrl: viper.GetString("MESSAGING_API_BASE_URL"),
		MessagingApiKey: viper.GetString("MESSAGING_API_KEY"),
	}

	// Build connection strings
	if config.Database.Type == "mysql" {
		config.Database.DSN = fmt.Sprintf("user=%s,password=%s,host=%s,port=%s,dbname=%s",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
		)
	}

	if config.Redis.Type == "redis" {
		config.Redis.URL = fmt.Sprintf("addresses=%s:%s,username=%s,password=%s,database=%s",
			config.Redis.Host,
			config.Redis.Port,
			config.Redis.UserName,
			config.Redis.Password,
			config.Redis.Db,
		)
	}

	return config
}
