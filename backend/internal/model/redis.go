package model

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"campus-service/internal/config"
	"campus-service/internal/pkg/logger"
)

// RedisClient 全局Redis客户端
var RedisClient *redis.Client

// InitRedis 初始化Redis连接
func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Cfg.Redis.GetRedisAddr(),
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	logger.Info("Redis connected successfully")
	return nil
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// SetAccessToken 存储access token到Redis
func SetAccessToken(ctx context.Context, userID uint, token string, expiry time.Duration) error {
	key := fmt.Sprintf("token:access:%d", userID)
	return RedisClient.Set(ctx, key, token, expiry).Err()
}

// GetAccessToken 从Redis获取access token
func GetAccessToken(ctx context.Context, userID uint) (string, error) {
	key := fmt.Sprintf("token:access:%d", userID)
	return RedisClient.Get(ctx, key).Result()
}

// DeleteAccessToken 删除access token
func DeleteAccessToken(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("token:access:%d", userID)
	return RedisClient.Del(ctx, key).Err()
}

// SetRefreshToken 存储refresh token到Redis
func SetRefreshToken(ctx context.Context, userID uint, token string, expiry time.Duration) error {
	key := fmt.Sprintf("token:refresh:%d", userID)
	return RedisClient.Set(ctx, key, token, expiry).Err()
}

// GetRefreshToken 从Redis获取refresh token
func GetRefreshToken(ctx context.Context, userID uint) (string, error) {
	key := fmt.Sprintf("token:refresh:%d", userID)
	return RedisClient.Get(ctx, key).Result()
}

// DeleteRefreshToken 删除refresh token
func DeleteRefreshToken(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("token:refresh:%d", userID)
	return RedisClient.Del(ctx, key).Err()
}

// ExistsToken 检查token是否存在
func ExistsToken(ctx context.Context, userID uint, tokenType string) (bool, error) {
	key := fmt.Sprintf("token:%s:%d", tokenType, userID)
	result, err := RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}