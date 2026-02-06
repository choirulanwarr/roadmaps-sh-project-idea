package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
	cacheTTL    = 12 * time.Hour // Cache expiration: 12 hours
)

// InitRedis initializes Redis connection
func InitRedis() error {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Test connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient = nil // Reset to nil if connection fails
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✅ Redis connected successfully")
	return nil
}

// GetFromCache retrieves data from Redis cache
func GetFromCache(key string, dest interface{}) bool {
	// Check if Redis client is initialized
	if redisClient == nil {
		log.Println("⚠️  Redis client not initialized, skipping cache")
		return false
	}

	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("Cache miss for key: %s", key)
		} else {
			log.Printf("⚠️  Cache get error for key %s: %v", key, err)
		}
		return false
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		log.Printf("⚠️  Cache unmarshal error for key %s: %v", key, err)
		return false
	}

	log.Printf("✅ Cache hit for key: %s", key)
	return true
}

// SetCache stores data in Redis with expiration
func SetCache(key string, value interface{}) error {
	// Check if Redis client is initialized
	if redisClient == nil {
		log.Println("⚠️  Redis client not initialized, skipping cache set")
		return nil // Don't return error, just skip caching
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	err = redisClient.Set(ctx, key, data, cacheTTL).Err()
	if err != nil {
		log.Printf("⚠️  Failed to set cache for key %s: %v", key, err)
		return fmt.Errorf("failed to set cache: %w", err)
	}

	log.Printf("💾 Cached data for key: %s", key)
	return nil
}

// DeleteCache removes a key from cache
func DeleteCache(key string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return redisClient.Del(ctx, key).Err()
}
