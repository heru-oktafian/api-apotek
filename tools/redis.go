package tools

import (
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

// Redis client instance
var RedisClient *redis.Client = func() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	var host, port string
	if addr != "" {
		// Parse REDIS_ADDR as host:port
		parts := strings.Split(addr, ":")
		if len(parts) == 2 {
			host = parts[0]
			port = parts[1]
		} else {
			host = "localhost"
			port = "6379"
		}
	} else {
		host = os.Getenv("REDIS_HOST")
		if host == "" {
			host = "localhost"
		}
		port = os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}
	}
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil {
			db = parsed
		}
	}
	return redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       db,
	})
}()
