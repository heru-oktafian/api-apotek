package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/heru-oktafian/api-apotek/models"
	"github.com/heru-oktafian/scafold/config"
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

// SetTemporaryProductCache menyimpan daftar produk sementara ke Redis dengan cacheKey sebagai pembeda
func SetTemporaryProductCache(cacheKey string, products []models.ProdSaleCombo) error {
	ctx := context.Background()
	// Ping Redis to check connection
	if _, err := config.RDB.Ping(ctx).Result(); err != nil {
		fmt.Printf("Redis ping failed: %v\n", err)
		return err
	}
	key := fmt.Sprintf("tmp:products:sale:%s", cacheKey)
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}
	// Set dengan TTL 30 menit
	err = config.RDB.Set(ctx, key, data, 30*time.Minute).Err()
	if err == nil {
		fmt.Printf("Successfully saved product cache to Redis key: %s\n", key)
	}
	return err
}

// GetTemporaryProductCache mengambil daftar produk sementara dari Redis berdasarkan cacheKey
func GetTemporaryProductCache(cacheKey string) ([]models.ProdSaleCombo, error) {
	ctx := context.Background()
	key := fmt.Sprintf("tmp:products:sale:%s", cacheKey)
	val, err := config.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Tidak ada data cache
	}
	if err != nil {
		return nil, err
	}
	var products []models.ProdSaleCombo
	if err := json.Unmarshal([]byte(val), &products); err != nil {
		return nil, err
	}
	return products, nil
}

// DeleteTemporaryProductCache menghapus cache produk sementara dari Redis berdasarkan cacheKey
func DeleteTemporaryProductCache(cacheKey string) error {
	ctx := context.Background()
	key := fmt.Sprintf("tmp:products:sale:%s", cacheKey)
	return config.RDB.Del(ctx, key).Err()
}

// SetTemporaryPurchaseProductCache menyimpan daftar produk pembelian sementara ke Redis dengan cacheKey sebagai pembeda
func SetTemporaryPurchaseProductCache(cacheKey string, products []models.ProdPurchaseCombo) error {
	ctx := context.Background()
	// Ping Redis to check connection
	if _, err := config.RDB.Ping(ctx).Result(); err != nil {
		fmt.Printf("Redis ping failed: %v\n", err)
		return err
	}
	key := fmt.Sprintf("tmp:products:purchase:%s", cacheKey)
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}
	// Set dengan TTL 30 menit
	err = config.RDB.Set(ctx, key, data, 30*time.Minute).Err()
	if err == nil {
		fmt.Printf("Successfully saved purchase product cache to Redis key: %s\n", key)
	}
	return err
}

// GetTemporaryPurchaseProductCache mengambil daftar produk pembelian sementara dari Redis berdasarkan cacheKey
func GetTemporaryPurchaseProductCache(cacheKey string) ([]models.ProdPurchaseCombo, error) {
	ctx := context.Background()
	key := fmt.Sprintf("tmp:products:purchase:%s", cacheKey)
	val, err := config.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Tidak ada data cache
	}
	if err != nil {
		return nil, err
	}
	var products []models.ProdPurchaseCombo
	if err := json.Unmarshal([]byte(val), &products); err != nil {
		return nil, err
	}
	return products, nil
}

// DeleteTemporaryPurchaseProductCache menghapus cache produk pembelian sementara dari Redis berdasarkan cacheKey
func DeleteTemporaryPurchaseProductCache(cacheKey string) error {
	ctx := context.Background()
	key := fmt.Sprintf("tmp:products:purchase:%s", cacheKey)
	return config.RDB.Del(ctx, key).Err()
}
