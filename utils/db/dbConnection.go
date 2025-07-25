// db/mysql.go

package db

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type mysqlConnectionConfig struct {
	cfg        *mysql.Config
	Alias      string
	RetryCount int
	RetryDelay time.Duration
	TimeZone   *time.Location
	DebugFlag  bool
	IsOptional bool
}

func (c *mysqlConnectionConfig) SetDefaultValues() {
	c.cfg = mysql.NewConfig()
	c.cfg.Net = "tcp"
	c.cfg.InterpolateParams = true
	c.cfg.ParseTime = true
	c.RetryCount = 5
	c.RetryDelay = time.Duration(100) * time.Millisecond
	c.DebugFlag = cast.ToBool(viper.GetString("SQL_DEBUG"))
	c.TimeZone = time.UTC
}

type MySQLDB struct {
	db     *sql.DB
	config mysqlConnectionConfig
}

func constructMySQLConnectionStringFromEnv() string {
	return fmt.Sprintf("user=%s,password=%s,host=%s,port=%s,dbname=%s",
		viper.GetString("MYSQL_DB_USERNAME"),
		viper.GetString("MYSQL_DB_PASSWORD"),
		viper.GetString("MYSQL_DB_HOST"),
		viper.GetString("MYSQL_DB_PORT"),
		viper.GetString("MYSQL_DB_SCHEMA"))
}

func SetMySqlConnection(connectionString string) (*MySQLDB, error) {
	if connectionString == "" {
		connectionString = constructMySQLConnectionStringFromEnv()
	}
	logger.Info("SetMySqlConnection", "connection_string", connectionString)
	// Parse the connection configuration
	cfg, err := parseMysqlConnectionConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection config: %w", err)
	}

	// Set alias for the connection
	cfg.Alias = "default"
	// Create the database connection
	mysqlDB, err := MakeMySQLConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}

	return mysqlDB, nil
}

func parseMysqlConnectionConfig(s string) (mysqlConnectionConfig, error) {
	var cfg mysqlConnectionConfig
	cfg.SetDefaultValues()

	expressions := strings.Split(s, ",")

	for _, e := range expressions {
		_, err := parse(e, &cfg)
		if err != nil {
			fmt.Printf("error evaluating %s, %+v\n", e, err)
			return cfg, err
		}
	}

	return cfg, nil
}

// Example expr string format:
// "user=myuser,password=mypassword,host=localhost,port=3306,dbname=mydb,optional=false,debug=true,retry_count=5,retry_delay=10s"
func parse(expr string, cfg *mysqlConnectionConfig) (bool, error) {
	parts := strings.SplitN(expr, "=", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid expression format: %s", expr)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "user":
		cfg.cfg.User = value
	case "password":
		cfg.cfg.Passwd = value
	case "host":
		cfg.cfg.Addr = value
	case "port":
		port := cast.ToInt(value)
		cfg.cfg.Addr = net.JoinHostPort(cfg.cfg.Addr, fmt.Sprintf("%d", port))
	case "dbname":
		cfg.cfg.DBName = value
	case "optional":
		cfg.IsOptional = cast.ToBool(value)
	case "debug":
		cfg.DebugFlag = cast.ToBool(value)
	case "retry_count":
		cfg.RetryCount = cast.ToInt(value)
	case "retry_delay":
		if d, err := time.ParseDuration(value); err == nil {
			cfg.RetryDelay = d
		} else {
			return false, fmt.Errorf("invalid retry_delay duration: %s", value)
		}
	default:
		return false, fmt.Errorf("unknown configuration key: %s", key)
	}

	return true, nil
}

// MakeMySQLConnection creates a new MySQL connection
func MakeMySQLConnection(config mysqlConnectionConfig) (*MySQLDB, error) {
	if config.RetryCount == 0 {
		config.RetryCount = 3 // default retry count
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second // default retry delay
	}
	if config.TimeZone == nil {
		config.TimeZone = time.UTC
	}

	mysqlDB := &MySQLDB{
		config: config,
	}

	err := mysqlDB.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	return mysqlDB, nil
}

func (m *MySQLDB) connect() error {
	var err error
	for attempt := 0; attempt <= m.config.RetryCount; attempt++ {
		if attempt > 0 {
			if m.config.DebugFlag {
				log.Printf("[%s] Retrying connection attempt %d/%d after %v",
					m.config.Alias, attempt, m.config.RetryCount, m.config.RetryDelay)
			}
			time.Sleep(m.config.RetryDelay)
		}

		m.config.cfg.Loc = m.config.TimeZone
		dsn := m.config.cfg.FormatDSN()
		if m.config.DebugFlag {
			log.Printf("[%s] Attempting to connect to MySQL with DSN: %s", m.config.Alias, dsn)
		}

		m.db, err = sql.Open("mysql", dsn)
		if err != nil {
			continue
		}

		// Test the connection
		err = m.Ping()
		if err == nil {
			if m.config.DebugFlag {
				log.Printf("[%s] Successfully connected to MySQL", m.config.Alias)
			}
			return nil
		}
	}

	if m.config.IsOptional {
		if m.config.DebugFlag {
			log.Printf("[%s] Failed to connect to optional MySQL database: %v", m.config.Alias, err)
		}
		return nil
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", m.config.RetryCount, err)
}

func (m *MySQLDB) Close() error {
	if m.db != nil {
		if m.config.DebugFlag {
			log.Printf("[%s] Closing MySQL connection", m.config.Alias)
		}
		return m.db.Close()
	}
	return nil
}

func (m *MySQLDB) DB() *sql.DB {
	return m.db
}

func (m *MySQLDB) Ping() error {
	if m.db == nil {
		return fmt.Errorf("database connection not initialized")
	}
	return m.db.Ping()
}
