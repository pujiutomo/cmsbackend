package controller

import (
	"context"
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pujiutomo/cmsbackend/database"
	"github.com/redis/go-redis/v9"
)

// Helper Redis save functions
func SaveToRedis(keyRedis string, data string, expiration ...time.Duration) error {
	var exp time.Duration //default 0 (no expiration)
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	userKeyRedis := keyRedis

	ctx := context.Background()
	err := database.Redis.Set(ctx, userKeyRedis, data, exp).Err()
	if err != nil {
		return fmt.Errorf("failed to save user to redis: %w", err)
	}
	return nil
}

func GetFromRedis(keyRedis string) ([]map[string]interface{}, error) {
	val, err := database.Redis.Get(context.Background(), keyRedis).Result()
	if err != nil {
		// Jika key tidak ditemukan
		if err == redis.Nil {
			return nil, fmt.Errorf("key %s not found in redis", keyRedis)
		}
		return nil, fmt.Errorf("failed to get data from redis for key %s: %w", keyRedis, err)
	}

	// Coba unmarshal sebagai array terlebih dahulu
	var arrayResult []map[string]interface{}
	if err := jsoniter.Unmarshal([]byte(val), &arrayResult); err == nil {
		return arrayResult, nil
	}

	// Jika bukan array, coba unmarshal sebagai single object
	var objectResult map[string]interface{}
	if err := jsoniter.Unmarshal([]byte(val), &objectResult); err == nil {
		// Wrap single object dalam array
		return []map[string]interface{}{objectResult}, nil
	}

	return nil, fmt.Errorf("failed to unmarshal JSON data for key %s: invalid format", keyRedis)
}
