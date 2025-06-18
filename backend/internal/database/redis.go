package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var RedisSubscriber *redis.Client

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// LoadRedisConfig loads Redis configuration from environment variables
func LoadRedisConfig() *RedisConfig {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	return &RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       0, // Default DB
	}
}

// ConnectRedis establishes Redis connection
func ConnectRedis() error {
	config := LoadRedisConfig()

	redisOptions := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	}

	// Create separate clients for publish and subscribe
	RedisClient = redis.NewClient(redisOptions)
	RedisSubscriber = redis.NewClient(redisOptions)

	// Test both connections
	ctx := context.Background()

	// Test publish client
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis (publish client): %w", err)
	}

	// Test subscribe client
	_, err = RedisSubscriber.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis (subscribe client): %w", err)
	}

	log.Println("Redis connection established successfully (publish + subscribe clients)")
	return nil
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	var err1, err2 error

	if RedisClient != nil {
		err1 = RedisClient.Close()
	}

	if RedisSubscriber != nil {
		err2 = RedisSubscriber.Close()
	}

	if err1 != nil {
		return err1
	}
	return err2
}
