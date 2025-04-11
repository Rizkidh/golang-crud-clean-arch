package config

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

func ConnectRedis() *redis.Client {
	// Gunakan satu variabel ENV REDIS_ADDR
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	fmt.Printf("✅ Connected to Redis at %s!\n", addr)
	return client
}
