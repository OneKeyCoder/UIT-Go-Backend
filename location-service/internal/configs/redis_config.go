package configs

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"common_pkg/helpers"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	TTL      int
}

var RedisClient *redis.Client

func GetRedisConfig() *RedisConfig {
	db, err := strconv.Atoi(helpers.GetEnv("REDIS_DB", "0"))
	if err != nil {
		log.Printf("Invalid REDIS_DB value, using default: 0")
		db = 0
	}

	ttl, err := strconv.Atoi(helpers.GetEnv("REDIS_TIME_TO_LIVE", "3600"))
	if err != nil {
		log.Printf("Invalid REDIS_TIME_TO_LIVE value, using default: 3600")
		ttl = 3600
	}

	return &RedisConfig{
		Host:     helpers.GetEnv("REDIS_HOST", "localhost"),
		Port:     helpers.GetEnv("REDIS_PORT", "6379"),
		Password: helpers.GetEnv("REDIS_PASSWORD", "redis123"),
		DB:       db,
		TTL:      ttl,
	}
}

func ConnectRedis() (*redis.Client, error) {
	config := GetRedisConfig()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Successfully connected to Redis")
	return client, nil
}

func InitRedis() error {
	client, err := ConnectRedis()
	if err != nil {
		return err
	}
	RedisClient = client
	return nil
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

func GetTTL() time.Duration {
	config := GetRedisConfig()
	return time.Duration(config.TTL) * time.Second
}
