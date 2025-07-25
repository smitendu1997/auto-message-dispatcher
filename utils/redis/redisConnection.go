// redis/redis.go

package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	redisGo "github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type RedisConfig struct {
	Addresses  []string // Multiple addresses for cluster
	Username   string
	Password   string
	Database   int
	IsCluster  bool
	MaxRetries int
	PoolSize   int
	TimeoutMs  int
	TLS        bool
}

type RedisClient struct {
	client redisGo.UniversalClient
	config *RedisConfig
}

func SetRedisConnection(connectionString string) (*RedisClient, error) {
	if connectionString == "" {
		// Use Viper to get values from .env file
		host := viper.GetString("REDIS_HOST")
		port := viper.GetString("REDIS_PORT")
		username := viper.GetString("REDIS_USERNAME")
		password := viper.GetString("REDIS_PASSWORD")
		db := viper.GetString("REDIS_DB")

		// Construct the connection string
		connectionString = fmt.Sprintf("addresses=%s:%s,username=%s,password=%s,database=%s",
			host, port, username, password, db)
	}

	cfg, err := parseRedisConnectionConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", connectionString, err)
	}

	redisClient, err := NewRedisClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error making connection %s: %w", connectionString, err)
	}

	return redisClient, nil
}

// Example connection configuration string format:
// addresses=localhost:6379,username=myuser,password=mypass,database=0,cluster=false,max_retries=3,pool_size=10,timeout_ms=5000,tls=false
func parseRedisConnectionConfig(s string) (*RedisConfig, error) {
	cfg := &RedisConfig{}
	expressions := strings.Split(s, ",")

	for _, e := range expressions {
		if err := parseRedisExpr(e, cfg); err != nil {
			return nil, fmt.Errorf("error evaluating %s: %w", e, err)
		}
	}

	return cfg, nil
}

func parseRedisExpr(expr string, cfg *RedisConfig) error {
	parts := strings.SplitN(expr, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid expression format: %s", expr)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "addresses":
		cfg.Addresses = strings.Split(value, ";")
	case "username":
		cfg.Username = value
	case "password":
		cfg.Password = value
	case "database":
		cfg.Database = cast.ToInt(value)
	case "cluster":
		cfg.IsCluster = cast.ToBool(value)
	case "max_retries":
		cfg.MaxRetries = cast.ToInt(value)
	case "pool_size":
		cfg.PoolSize = cast.ToInt(value)
	case "timeout_ms":
		cfg.TimeoutMs = cast.ToInt(value)
	case "tls":
		cfg.TLS = cast.ToBool(value)
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}

func NewRedisClient(cfg *RedisConfig) (*RedisClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config is required")
	}

	if len(cfg.Addresses) == 0 {
		return nil, fmt.Errorf("redis address is required")
	}

	// Set defaults
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 10
	}
	if cfg.TimeoutMs == 0 {
		cfg.TimeoutMs = 5000
	}

	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond

	universalOptions := &redisGo.UniversalOptions{
		Addrs:        cfg.Addresses,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.Database,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		TLSConfig:    nil, // Configure if cfg.TLS is true
	}

	client := redisGo.NewUniversalClient(universalOptions)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{
		client: client,
		config: cfg,
	}, nil
}

func (r *RedisClient) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func (r *RedisClient) Client() redisGo.UniversalClient {
	return r.client
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Helper methods for common operations
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}
