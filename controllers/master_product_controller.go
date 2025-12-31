package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/heru-oktafian/api-apotek/models"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/helpers"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
)

// Redis client instance
var redisClient *redis.Client = func() *redis.Client {
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

// CreateProduct buat Product
func CreateProduct(c *framework.Ctx) error {
	// get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Creating new Product using helpers
	return helpers.CreateResource(c, config.DB, &models.Product{}, branch_id, "PRD")
}

// UpdateProduct update Product
func UpdateProduct(c *framework.Ctx) error {
	id := c.Param("id")
	// Updating Product using helpers
	return helpers.UpdateResource(c, config.DB, &models.Product{}, id)
}

// DeleteProduct hapus Product
func DeleteProduct(c *framework.Ctx) error {
	id := c.Param("id")
	// Deleting Product using helpers
	return helpers.DeleteResource(c, config.DB, &models.Product{}, id)
}

// GetProduct tampilkan Product berdasarkan id
func GetProduct(c *framework.Ctx) error {
	id := c.Param("id")
	var AllProduct []models.ProductDetail
	if err := config.DB.
		Table("products pro").
		Select("pro.id,pro.sku,pro.name,pro.description, pro.ingredient, pro.dosage, pro.side_affection, pro.unit_id AS unit_id,pro.stock,pro.purchase_price,pro.expired_date,pro.sales_price, pro.alternate_price, pro.product_category_id,pc.name AS product_category_name,un.name AS unit_name,pro.branch_id").
		Joins("LEFT JOIN product_categories pc ON pc.id = pro.product_category_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pro.id = ?", id).
		Scan(&AllProduct).Error; err != nil {
		return responses.JSONResponse(c, http.StatusNotFound, "Data tidak ditemukan", err)
	}

	// print(AllProduct)
	return responses.JSONResponse(c, http.StatusOK, "Data ditemukan", AllProduct)
}

// GetAllProduct tampilkan semua Product
func GetAllProduct(c *framework.Ctx) error {
	// Ambil ID cabang
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10                  // Tetapkan batas data per halaman ke 10
	offset := (page - 1) * limit // Hitung offset berdasarkan halaman dan batas

	var AllProduct []models.ProductDetail
	var total int64

	// Query dasar
	query := config.DB.Table("products pro").
		Select("pro.id,pro.sku,pro.name,pro.description, pro.ingredient, pro.dosage, pro.side_affection, pro.unit_id, un.name AS unit_name,pro.stock,pro.purchase_price,pro.sales_price,pro.alternate_price,pro.expired_date, pro.product_category_id, pc.name AS product_category_name").
		Joins("LEFT JOIN product_categories pc ON pc.id = pro.product_category_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pro.branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search = strings.TrimSpace(search); search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(pro.name) LIKE ? OR LOWER(pro.description) LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Tambahkan sorting ascending berdasarkan pro.name
	query = query.Order("pro.name ASC")

	// Hitung total produk yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get Products failed", "Failed to count Products")
	}

	// Ambil data dengan pagination
	if err := query.Offset(offset).Limit(limit).Scan(&AllProduct).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get Products failed", "Failed to fetch Products with details")
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(c, http.StatusOK, "Products retrieved successfully", search, int(total), page, int(totalPages), int(limit), AllProduct)

}

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

// CmbProdSale mengembalikan daftar produk untuk combo box transaksi penjualan
func CmbProdSale(c *framework.Ctx) error {
	branch_id, _ := middlewares.GetBranchID(c.Request)
	user_id, _ := middlewares.GetUserID(c.Request)
	search := strings.TrimSpace(c.Query("search"))

	// Buat cacheKey berdasarkan branch_id dan user_id
	cacheKey := fmt.Sprintf("%v:%v", branch_id, user_id)

	// Cek apakah ada data di Redis terlebih dahulu
	cachedProducts, err := GetTemporaryProductCache(cacheKey)
	if err != nil {
		fmt.Printf("Failed to get product cache for cacheKey %s: %v\n", cacheKey, err)
		// Lanjutkan ke query database jika gagal ambil cache
	}
	if cachedProducts != nil {
		// Jika ada data di cache, gunakan data tersebut
		return responses.JSONResponse(c, http.StatusOK, "Combo Products retrieved from cache successfully", cachedProducts)
	}

	// Jika tidak ada di cache, lakukan query ke database
	var cmbProducts []models.ProdSaleCombo

	query := config.DB.Table("products").
		Select("products.id as product_id, products.name as product_name, sales_price AS price, products.stock, products.unit_id, units.name AS unit_name").
		Joins("LEFT JOIN units ON units.id = products.unit_id").
		Where("products.branch_id = ?", branch_id)

	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(products.name) LIKE ? OR LOWER(products.description) LIKE ? OR LOWER(products.id) LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	query = query.Order("products.name ASC")

	if err := query.Scan(&cmbProducts).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get Combo Products failed", err)
	}

	// Simpan list produk ke Redis dengan cacheKey
	if err := SetTemporaryProductCache(cacheKey, cmbProducts); err != nil {
		fmt.Printf("Failed to save product cache for cacheKey %s: %v\n", cacheKey, err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Combo Products retrieved successfully", cmbProducts)
}

// CmbProdPurchase mengembalikan daftar produk untuk combo box transaksi pembelian
func CmbProdPurchase(c *framework.Ctx) error {
	// get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Ambil search key dari query parameter
	search := strings.TrimSpace(c.Query("search")) // Default to empty string if not provided

	// deklarasi variabel untuk combo box produk transaksi pembelian
	var cmbProducts []models.ProdPurchaseCombo

	// ambil data produk untuk combo box transaksi pembelian
	query := config.DB.Table("products").
		Select("products.id as product_id, products.name as product_name, purchase_price AS price, products.unit_id, units.name AS unit_name").
		Joins("LEFT JOIN units ON units.id = products.unit_id").
		Where("products.branch_id = ?", branch_id)

	// Jika search tidak kosong, tambahkan kondisi LIKE
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(products.name) LIKE ? OR LOWER(products.description) LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Tambahkan sorting ascending berdasarkan products.name
	query = query.Order("products.name ASC")

	if err := query.Scan(&cmbProducts).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get Combo Products failed", err)
	}
	return responses.JSONResponse(c, http.StatusOK, "Combo Products retrieved successfully", cmbProducts)
}
