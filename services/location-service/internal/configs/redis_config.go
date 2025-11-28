package configs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/env"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
	db, err := strconv.Atoi(env.Get("REDIS_DB", "0"))
	if err != nil {
		db = 0
	}

	ttl, err := strconv.Atoi(env.Get("REDIS_TIME_TO_LIVE", "3600"))
	if err != nil {
		ttl = 3600
	}

	return &RedisConfig{
		Host:     env.Get("REDIS_HOST", "redis"),
		Port:     env.Get("REDIS_PORT", "6379"),
		Password: env.Get("REDIS_PASSWORD", "redispassword"),
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

	logger.Info("Successfully connected to Redis",
		zap.String("host", config.Host),
		zap.String("port", config.Port))
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
