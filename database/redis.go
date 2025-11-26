package database

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func RedisClient() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	redisHost := os.Getenv("REDISHOST")
	redisPass := os.Getenv("REDISPASS")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPass,
		DB:       0,
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		panic("Could not connect the database redis")
	}
	Redis = redisClient
}
