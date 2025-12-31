package tools

import (
	"context"
	"fmt"

	"github.com/heru-oktafian/scafold/config"
)

// UpdateProductStockInRedisAsync updates the product stock in the temporary cache with branch and user context asynchronously
func UpdateProductStockInRedisAsync(cacheKey, productID string, newStock int) {
	go func() {
		ctx := context.Background()
		// Ping Redis to check connection
		if _, err := config.RDB.Ping(ctx).Result(); err != nil {
			fmt.Printf("Redis ping failed: %v\n", err)
			return
		}

		// Ambil data cache produk
		products, err := GetTemporaryProductCache(cacheKey)
		if err != nil {
			fmt.Printf("Failed to get product cache for cacheKey %s: %v\n", cacheKey, err)
			return
		}
		if products == nil {
			fmt.Printf("No product cache found for cacheKey %s\n", cacheKey)
			return
		}

		// Cari dan update stock produk
		found := false
		for i := range products {
			if products[i].ProductId == productID {
				products[i].Stock = newStock
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Product %s not found in cache for cacheKey %s\n", productID, cacheKey)
			return
		}

		// Simpan kembali ke Redis
		if err := SetTemporaryProductCache(cacheKey, products); err != nil {
			fmt.Printf("Failed to update product cache for cacheKey %s: %v\n", cacheKey, err)
			return
		}

		fmt.Printf("Successfully updated stock for product %s in cache key: tmp:products:sale:%s\n", productID, cacheKey)
	}()
}
