package controllers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/api-retail/tools"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/helpers"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
	"gorm.io/gorm"
)

// CreateFirstStock Function
func CreateFirstStock(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB

	// Ambil informasi dari token
	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)
	generatedID := helpers.GenerateID("FST")

	// Ambil input dari body
	var input models.FirstStockInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, input.FirstStockDate)
	if err != nil {
		return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", err)
	}

	// Map ke struct model
	first_stock := models.FirstStocks{
		ID:              generatedID,
		Description:     input.Description,
		BranchID:        branchID,
		UserID:          userID,
		FirstStockDate:  parsedDate,
		TotalFirstStock: 0,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}

	// Simpan first_stock
	if err := db.Create(&first_stock).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create FirstStock", err)
	}

	// Buat laporan
	if err := SyncFirstStockReport(db, first_stock); err != nil {
		return responses.InternalServerError(c, "Failed to sync FirstStock report", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "FirstStock created successfully", first_stock)
}

// UpdateFirstStock Function
func UpdateFirstStock(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	id := c.Param("id")

	// Cari data first_stock lama
	var first_stock models.FirstStocks
	if err := db.First(&first_stock, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "FirstStock not found")
	}

	// Gunakan struct input
	var input models.FirstStockInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Cek dan update FirstStockDate
	if input.FirstStockDate != "" {
		layout := "2006-01-02"
		parsedDate, err := time.Parse(layout, input.FirstStockDate)
		if err != nil {
			return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", err)
		}
		first_stock.FirstStockDate = parsedDate
	}

	// Cek dan update Payment
	if input.Payment != "" {
		first_stock.Payment = models.PaymentStatus(input.Payment)
	}

	first_stock.UpdatedAt = nowWIB

	// Hitung ulang total dari first_stock items
	var items []models.FirstStockItems
	if err := db.Where("first_stock_id = ?", id).Find(&items).Error; err != nil {
		return responses.InternalServerError(c, "Failed to retrieve FirstStock items", err)
	}

	if len(items) == 0 {
		first_stock.TotalFirstStock = 0
	} else {
		total := 0
		for _, item := range items {
			total += item.SubTotal
		}
		first_stock.TotalFirstStock = total
	}

	// Cek dan update Description
	if input.Description != "" {
		first_stock.Description = input.Description
	}

	// Simpan perubahan
	if err := db.Save(&first_stock).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update FirstStock", err)
	}

	// Sync report
	if err := SyncFirstStockReport(db, first_stock); err != nil {
		return responses.InternalServerError(c, "Failed to sync FirstStock report", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "FirstStock updated successfully", first_stock)
}

// DeleteFirstStock Function
func DeleteFirstStock(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Ambil first_stock
	var first_stock models.FirstStocks
	if err := db.First(&first_stock, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "FirstStock not found")
	}

	// Ambil item-item dan rollback stok
	var items []models.FirstStockItems
	if err := db.Where("first_stock_id = ?", id).Find(&items).Error; err != nil {
		return responses.InternalServerError(c, "Failed to retrieve FirstStock items", err)
	}

	for _, item := range items {
		// Kurangi stok ke produk
		if err := tools.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
			return responses.InternalServerError(c, "Failed to reduce product stock", err)
		}
	}

	// Hapus semua item dari pembelian
	if err := db.Where("first_stock_id = ?", id).Delete(&models.FirstStockItems{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete FirstStock items", err)
	}

	// Hapus laporan transaksi terkait
	if err := db.Where("id = ? AND transaction_type = ?", first_stock.ID, models.FirstStock).Delete(&models.TransactionReports{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete TransactionReports", err)
	}

	// Hapus first_stock
	if err := db.Delete(&first_stock).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete FirstStock", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "FirstStock deleted successfully", first_stock)
}

// CreateFirstStockItem Function
func CreateFirstStockItem(c *framework.Ctx) error {
	db := config.DB
	var item models.FirstStockItems

	if err := c.BodyParser(&item); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Cek apakah item dengan first_stock_id dan product_id sudah ada
	var existing models.FirstStockItems
	err := db.Where("first_stock_id = ? AND product_id = ?", item.FirstStockId, item.ProductId).First(&existing).Error
	if err == nil {
		// Sudah ada: update qty dan sub_total
		existing.Qty += item.Qty
		existing.SubTotal = existing.Qty * existing.Price // asumsi pakai harga awal

		if err := db.Save(&existing).Error; err != nil {
			return responses.InternalServerError(c, "Failed to update FirstStock item", err)
		}

		// Tambah stok
		if err := tools.AddProductStock(db, item.ProductId, item.Qty); err != nil {
			return responses.InternalServerError(c, "Failed to add product stock", err)
		}

		// Update harga produk jika harga baru lebih tinggi dari yang tersimpan di tabel products
		if err := tools.UpdateProductPriceIfHigher(db, item.ProductId, item.Price); err != nil {
			return responses.InternalServerError(c, "Failed to update product price", err)
		}

		// Recalculate total pembelian
		if err := RecalculateTotalFirstStock(db, item.FirstStockId); err != nil {
			return responses.InternalServerError(c, "Failed to recalculate total FirstStock", err)
		}

		return responses.JSONResponse(c, http.StatusOK, "Item updated successfully", existing)

	} else if err != gorm.ErrRecordNotFound {
		// Error selain record not found
		return responses.InternalServerError(c, "Failed to retrieve FirstStock item", err)
	}

	// Data belum ada, buat item baru
	if item.ID == "" {
		item.ID = helpers.GenerateID("FSI")
	}
	item.SubTotal = item.Qty * item.Price

	if err := db.Create(&item).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create FirstStock item", err)
	}

	if err := tools.AddProductStock(db, item.ProductId, item.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to add product stock", err)
	}

	if err := tools.UpdateProductPriceIfHigher(db, item.ProductId, item.Price); err != nil {
		return responses.InternalServerError(c, "Failed to update product price", err)
	}

	if err := RecalculateTotalFirstStock(db, item.FirstStockId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total FirstStock", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item added successfully", item)
}

// Update FirstStockItem
func UpdateFirstStockItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var existingItem models.FirstStockItems
	if err := db.First(&existingItem, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Item not found")
	}

	var updatedItem models.FirstStockItems
	if err := c.BodyParser(&updatedItem); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Rollback stok lama
	if err := tools.ReduceProductStock(db, existingItem.ProductId, existingItem.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to rollback product stock", err)
	}

	// Tambah stok baru
	if err := tools.AddProductStock(db, updatedItem.ProductId, updatedItem.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to add product stock", err)
	}

	// Update item
	existingItem.ProductId = updatedItem.ProductId
	existingItem.Qty = updatedItem.Qty
	existingItem.Price = updatedItem.Price
	existingItem.SubTotal = updatedItem.Price * updatedItem.Qty

	if err := db.Save(&existingItem).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update FirstStock item", err)
	}

	// Update harga produk jika harga item lebih tinggi
	if err := tools.UpdateProductPriceIfHigher(db, updatedItem.ProductId, updatedItem.Price); err != nil {
		return responses.InternalServerError(c, "Failed to update product price", err)
	}

	// Recalculate total & sync
	if err := RecalculateTotalFirstStock(db, existingItem.FirstStockId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total FirstStock", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item updated successfully", existingItem)
}

// Delete FirstStockItem
func DeleteFirstStockItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var item models.FirstStockItems
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Item not found")
	}

	// Subtract stok
	if err := tools.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to rollback product stock", err)
	}

	// Hapus item
	if err := db.Delete(&item).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete FirstStock item", err)
	}

	// Recalculate total
	if err := RecalculateTotalFirstStock(db, item.FirstStockId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total FirstStock", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item deleted successfully", item)
}

// Get All FirstStocks tampilkan semua first_stock
func GetAllFirstStocks(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10 // Tetapkan limit ke 10 data per halaman
	offset := (page - 1) * limit

	var FirstStocks []models.AllFirstStocks
	var total int64

	// Query dasar
	query := config.DB.Table("first_stocks pur").
		Select("pur.id, pur.description, pur.first_stock_date, pur.total_first_stock, pur.payment").
		Where("pur.branch_id = ?", branch_id).
		Order("pur.created_at DESC")

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(pur.description) LIKE ?", "%"+search+"%")
	}

	// Hitung total first_stock yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Get FirstStock failed", err)
	}

	// Ambil data dengan pagination
	if err := query.Offset(offset).Limit(limit).Scan(&FirstStocks).Error; err != nil {
		return responses.InternalServerError(c, "Get first_stocks failed", err)
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(c, http.StatusOK, "FirstStocks retrieved successfully", search, int(total), page, int(totalPages), int(limit), FirstStocks)
}

// GetAllFirstStockItems tampilkan semua item berdasarkan first_stock_id tanpa pagination
func GetAllFirstStockItems(c *framework.Ctx) error {
	// Get FirstStock id dari param
	first_stockID := c.Param("id")

	search := strings.TrimSpace(c.Query("search"))

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search)
	}

	var FirstStockItems []models.AllFirstStockItems

	// Query dasar
	query := config.DB.Table("first_stock_items pit").
		Select("pit.id, pit.first_stock_id, pit.product_id, pro.name AS product_name, pit.price, pit.qty, un.name AS unit_name, pit.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pit.first_stock_id = ?", first_stockID).
		Order("pro.name ASC")

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(pro.name) LIKE ?", "%"+search+"%")
	}

	// Eksekusi query
	if err := query.Scan(&FirstStockItems).Error; err != nil {
		return responses.InternalServerError(c, "Get items failed", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Items retrieved successfully", FirstStockItems)
}

// GetFirstStockWithItems menampilkan satu first_stock beserta semua item-nya
func GetFirstStockWithItems(c *framework.Ctx) error {
	db := config.DB

	// Ambil ID pembelian dari parameter URL
	first_stockID := c.Param("id")

	// Struct untuk data utama first_stock
	var first_stock models.AllFirstStocks

	// Ambil data first_stock
	err := db.Table("first_stocks pur").
		Select("pur.id, pur.description, pur.first_stock_date, pur.total_first_stock, pur.payment").
		Where("pur.id = ?", first_stockID).
		Scan(&first_stock).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to get first_stock", err)
	}

	// Ambil item pembelian terkait
	var items []models.AllFirstStockItems
	err = db.Table("first_stock_items pit").
		Select("pit.id, pit.first_stock_id, pit.product_id, pro.name AS product_name, pit.price, pit.qty, un.name AS unit_name, pit.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pit.first_stock_id = ?", first_stockID).
		Order("pro.name ASC").
		Scan(&items).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to get FirstStock items", err)
	}

	// Format tanggal pembelian ke dd-mm-yyyy
	formattedDate := first_stock.FirstStockDate.Format("02-01-2006")

	return JSONFirstStockWithItemsResponse(c, http.StatusOK, "FirstStock retrieved successfully", first_stockID, first_stock.Description, formattedDate, first_stock.TotalFirstStock, formattedDate, items)
}

// --- Structs Permintaan untuk First Stock ---
type FirstStockTransactionRequest struct {
	FirstStock      FirstStockInput       `json:"first_stock" validate:"required"`
	FirstStockItems []FirstStockItemInput `json:"first_stock_items" validate:"required,min=1,dive"`
}

type FirstStockInput struct {
	Description    string `json:"description"`
	FirstStockDate string `json:"first_stock_date"` // String untuk parsing dari request
}

type FirstStockItemInput struct {
	ProductId   string `json:"product_id" validate:"required"`
	UnitId      string `json:"unit_id" validate:"required"`
	Qty         int    `json:"qty" validate:"required,min=1"`
	ExpiredDate string `json:"expired_date" validate:"required"` // String untuk parsing dari request
}

// --- Structs Respons untuk First Stock ---
type FirstStockOutput struct {
	ID              string `json:"id"`
	Description     string `json:"description"`
	FirstStockDate  string `json:"first_stock_date"` // Format YYYY-MM-DD
	BranchID        string `json:"branch_id"`
	TotalFirstStock int    `json:"total_first_stock"` // Ini adalah nilai stok yang ditambahkan
	Payment         string `json:"payment"`           // Akan diisi default "unpaid" atau "no_cost"
	UserID          string `json:"user_id"`
	CreatedAt       string `json:"created_at"` // Format YYYY-MM-DD
	UpdatedAt       string `json:"updated_at"` // Format YYYY-MM-DD
}

type FirstStockItemResponse struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"` // Nama produk untuk respons
	UnitID      string `json:"unit_id"`
	UnitName    string `json:"unit_name"`    // Nama unit untuk respons
	Price       int    `json:"price"`        // Harga beli per unit dasar
	Qty         int    `json:"qty"`          // Qty dalam unit input
	SubTotal    int    `json:"sub_total"`    // SubTotal berdasarkan Price * Qty (nilai stok)
	ExpiredDate string `json:"expired_date"` // Format tanggal kedaluwarsa
}

type FirstStockTransactionResponse struct {
	FirstStock      FirstStockOutput         `json:"first_stock"`
	FirstStockItems []FirstStockItemResponse `json:"first_stock_items"`
}

// `CreateFirstStockTransaction` Controller (Full Fungsional)

// CreateFirstStockTransaction controller
func CreateFirstStockTransaction(c *framework.Ctx) error {
	nowWIB := time.Now().In(utils.Location)

	subscriptionType, _ := middlewares.GetClaimsToken(c.Request, "subscription_type")
	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)

	db := config.DB
	var req FirstStockTransactionRequest
	err := c.BodyParser(&req)
	if err != nil {
		return responses.BadRequest(c, "Invalid request body", err)
	}

	// Set Payment secara default karena ini 'first_stock' (tidak ada pembiayaan)
	// Anda bisa pilih "nocost" atau jika punya models.NoCost, gunakan itu.
	var paymentStatus models.PaymentStatus = "nocost" // Default ke nocost

	// Inisialisasi header FirstStock dengan data dari token dan default payment
	firstStockHeader := models.FirstStocks{
		UserID:   userID,
		BranchID: branchID,
		Payment:  paymentStatus,
	}

	// --- VALIDASI INPUT ---
	// Validasi input header dan item
	if err = utils.ValidateStruct(req.FirstStock); err != nil {
		return responses.BadRequest(c, "Validation failed for first stock header input", err)
	}
	for _, item := range req.FirstStockItems {
		if err = utils.ValidateStruct(item); err != nil {
			return responses.BadRequest(c, "Validation failed for one or more first stock items", err)
		}
	}
	// --- AKHIR VALIDASI INPUT ---

	// Parse FirstStockDate
	var parsedFirstStockDate time.Time
	if req.FirstStock.FirstStockDate == "" {
		parsedFirstStockDate = nowWIB
	} else {
		parsedFirstStockDate, err = time.Parse("2006-01-02", req.FirstStock.FirstStockDate)
		if err != nil {
			return responses.BadRequest(c, "Invalid first_stock_date format. Please use `YYYY-MM-DD`.", err)
		}
	}

	// Mengisi detail FirstStocks dari request dan data token/default
	firstStockHeader.ID = helpers.GenerateID("FST") // Generate ID untuk First Stock
	firstStockHeader.Description = req.FirstStock.Description
	firstStockHeader.FirstStockDate = parsedFirstStockDate
	firstStockHeader.CreatedAt = nowWIB
	firstStockHeader.UpdatedAt = nowWIB

	// --- Proses Penyimpanan Data (Dalam Transaksi Database) ---
	tx := db.Begin()
	if tx.Error != nil {
		return responses.InternalServerError(c, "Failed to begin database transaction", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var calculatedTotalFirstStock int
	var firstStockItemsToCreate []models.FirstStockItems
	var firstStockItemsForResponse []FirstStockItemResponse // Slice untuk data respons

	// var stockTracksToCreate []models.StockTracks

	for _, reqItem := range req.FirstStockItems {
		parsedExpiredDate, err := time.Parse("2006-01-02", reqItem.ExpiredDate)
		if err != nil {
			tx.Rollback()
			return responses.BadRequest(c, fmt.Sprintf("Invalid expired_date format for product %s. Please use `YYYY-MM-DD`.", reqItem.ProductId), err)
		}

		var product models.Product
		err = tx.Where("id = ? AND branch_id = ?", reqItem.ProductId, firstStockHeader.BranchID).First(&product).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Product with ID %s not found in branch %s", reqItem.ProductId, firstStockHeader.BranchID))
			}
			return responses.InternalServerError(c, "Failed to retrieve product details", err)
		}

		// Mendapatkan detail unit (sesuai unit_id yang diinput)
		var unit models.Unit
		err = tx.Where("id = ?", reqItem.UnitId).First(&unit).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Unit with ID %s not found", reqItem.UnitId))
			}
			return responses.InternalServerError(c, "Failed to retrieve unit details", err)
		}

		// --- Logika Konversi Satuan ---
		var conversionValue int = 1
		if reqItem.UnitId != product.UnitId { // Jika unit input berbeda dengan unit dasar produk
			var unitConversion models.UnitConversion
			err = tx.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
				reqItem.ProductId,
				reqItem.UnitId, // Unit yang diinput
				product.UnitId, // Unit dasar produk
				firstStockHeader.BranchID,
			).First(&unitConversion).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// Jika tidak ada konversi yang didefinisikan, asumsikan 1:1 atau unit dasar.
					// Anda bisa menambahkan error di sini jika konversi mutlak diperlukan.
					// Saat ini dibiarkan conversionValue = 1
				} else {
					tx.Rollback()
					return responses.InternalServerError(c, "Failed to retrieve unit conversion details", err)
				}
			} else {
				conversionValue = unitConversion.ValueConv
			}
		}
		actualQtyToAdd := reqItem.Qty * conversionValue // Kuantitas aktual dalam satuan dasar
		// --- Akhir Logika Konversi Satuan ---

		// Harga untuk First Stock Items diambil dari PurchasePrice produk
		// Ini merepresentasikan "nilai" dari stok yang masuk, bukan biaya.
		itemPrice := product.PurchasePrice         // Harga beli per unit dasar produk
		itemSubTotal := itemPrice * actualQtyToAdd // SubTotal berdasarkan harga beli dan kuantitas aktual

		firstStockItemDB := models.FirstStockItems{
			ID:           helpers.GenerateID("FSI"), // ID untuk First Stock Item
			FirstStockId: firstStockHeader.ID,
			ProductId:    reqItem.ProductId,
			Price:        itemPrice,    // Price dari PurchasePrice produk
			Qty:          reqItem.Qty,  // Qty yang diinput (dalam unit yang diinput)
			SubTotal:     itemSubTotal, // SubTotal yang dihitung
			ExpiredDate:  parsedExpiredDate,
		}
		firstStockItemsToCreate = append(firstStockItemsToCreate, firstStockItemDB)

		// --- Siapkan data untuk respons ---
		firstStockItemResp := FirstStockItemResponse{
			ID:          firstStockItemDB.ID,
			ProductID:   firstStockItemDB.ProductId,
			ProductName: product.Name,
			UnitID:      reqItem.UnitId, // Unit yang diinput
			UnitName:    unit.Name,
			Price:       firstStockItemDB.Price,
			Qty:         firstStockItemDB.Qty,
			SubTotal:    firstStockItemDB.SubTotal,
			ExpiredDate: parsedExpiredDate.Format("02 January 2006"), // Format tanggal
		}
		firstStockItemsForResponse = append(firstStockItemsForResponse, firstStockItemResp)
		// --- Akhir persiapan data respons ---

		// --- Tambah stok dan cek/update expired_date ---
		updates := map[string]interface{}{
			"stock": product.Stock + actualQtyToAdd, // Tambahkan stok aktual (dalam satuan dasar)
		}

		// Jika ExpiredDate stok baru lebih awal dari yang sudah ada di master produk, update.
		if parsedExpiredDate.Before(product.ExpiredDate) {
			updates["expired_date"] = parsedExpiredDate
		}

		err = tx.Model(&models.Product{}).Where("id = ?", product.ID).Updates(updates).Error
		if err != nil {
			tx.Rollback()
			return responses.InternalServerError(c, fmt.Sprintf("Failed to update product details (stock/expired_date) for product %s", product.Name), err)
		}

		calculatedTotalFirstStock += itemSubTotal // Ini adalah nilai total stok yang dimasukkan
	}

	firstStockHeader.TotalFirstStock = calculatedTotalFirstStock

	// Simpan data FirstStocks
	err = tx.Create(&firstStockHeader).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create first stock entry", err)
	}

	// Simpan FirstStockItems dalam batch
	err = tx.CreateInBatches(&firstStockItemsToCreate, len(firstStockItemsToCreate)).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create first stock items", err)
	}

	// PENTING: TransactionReports dan DailyProfitReport TIDAK relevan untuk First Stock
	// Karena ini bukan transaksi finansial atau penjualan/pembelian berbiaya,
	// bagian untuk membuat TransactionReports atau mengupdate DailyProfitReport dihapus.

	// Cek `subscription_type` jika type nya adalah `quota`
	// Asumsi: First Stock TIDAK mengurangi kuota transaksi.
	// Jika first stock harus mengurangi kuota (misal, setiap entri dianggap transaksi),
	// Anda bisa menambahkan logika pengurangan kuota di sini.
	if subscriptionType == "quota" {
		var branch models.Branch
		err = tx.Where("id = ?", firstStockHeader.BranchID).First(&branch).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Branch with ID %s not found", firstStockHeader.BranchID))
			}
			return responses.InternalServerError(c, "Failed to retrieve branch details for quota check", err)
		}
		// Logika pengurangan kuota dibiarkan kosong di sini, karena first stock tidak mengurangi kuota.
	}

	err = tx.Commit().Error
	if err != nil {
		return responses.InternalServerError(c, "Failed to commit database transaction for first stock", err)
	}

	// --- Mengkonstruksi Objek Respon ---
	response := FirstStockTransactionResponse{
		FirstStock: FirstStockOutput{
			ID:              firstStockHeader.ID,
			Description:     firstStockHeader.Description,
			FirstStockDate:  firstStockHeader.FirstStockDate.Format("2006-01-02"), // Format YYYY-MM-DD
			BranchID:        firstStockHeader.BranchID,
			TotalFirstStock: firstStockHeader.TotalFirstStock,
			Payment:         string(firstStockHeader.Payment),
			UserID:          firstStockHeader.UserID,
			CreatedAt:       firstStockHeader.CreatedAt.Format("2006-01-02"), // Format YYYY-MM-DD
			UpdatedAt:       firstStockHeader.UpdatedAt.Format("2006-01-02"), // Format YYYY-MM-DD
		},
		FirstStockItems: firstStockItemsForResponse,
	}
	// --- Akhir Mengkonstruksi Objek Respon ---

	return responses.JSONResponse(c, http.StatusCreated, "First stock transaction created successfully", response)
}

// Insert atau update laporan transaksi berdasarkan FirstStocks / Pengeluaran
func SyncFirstStockReport(db *gorm.DB, first_stock models.FirstStocks) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Siapkan data report dari FirstStock
	report := models.TransactionReports{
		ID:              first_stock.ID,
		TransactionType: models.FirstStock,
		UserID:          first_stock.UserID,
		BranchID:        first_stock.BranchID,
		Total:           first_stock.TotalFirstStock,
		CreatedAt:       first_stock.CreatedAt,
		UpdatedAt:       first_stock.UpdatedAt,
		Payment:         first_stock.Payment,
	}

	var existing models.TransactionReports
	err := db.Take(&existing, "id = ?", report.ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert
		return db.Create(&report).Error
	}
	if err != nil {
		return err
	}

	// Jika ditemukan, lakukan update pada kolom yang dibutuhkan
	existing.Total = report.Total
	existing.UpdatedAt = nowWIB
	existing.Payment = report.Payment

	return db.Save(&existing).Error
}

func RecalculateTotalFirstStock(db *gorm.DB, first_stockID string) error {
	var total int64

	// Hitung total sub_total dari first_stock_items
	err := db.Model(&models.FirstStockItems{}).
		Where("first_stock_id = ?", first_stockID).
		Select("COALESCE(SUM(sub_total), 0)").
		Scan(&total).Error

	if err != nil {
		return err
	}

	// Update ke first_stocks
	if err := db.Model(&models.FirstStocks{}).
		Where("id = ?", first_stockID).
		Update("total_first_stock", total).Error; err != nil {
		return err
	}

	// Ambil first_stock lengkap buat update report
	var first_stock models.FirstStocks
	if err := db.First(&first_stock, "id = ?", first_stockID).Error; err != nil {
		return err
	}

	// Update transaction_reports juga
	if err := SyncFirstStockReport(db, first_stock); err != nil {
		return err
	}

	return nil
}

// JSONFirstStockWithItemsResponse sends a standard JSON response format / structure
func JSONFirstStockWithItemsResponse(c *framework.Ctx, status int, message string, first_stock_id string, description string, first_stock_date string, total_first_stock int, payment string, items interface{}) error {
	resp := ResponseFirstStockWithItemsResponse{
		Status:          http.StatusText(status),
		Message:         message,
		FirstStockId:    first_stock_id,
		Description:     description,
		FirstStockDate:  first_stock_date,
		TotalFirstStock: total_first_stock,
		Payment:         payment,
		Items:           items,
	}
	return responses.JSONResponse(c, status, message, resp)
}

// Response menampilkan satu first_stock beserta semua item-nya
type ResponseFirstStockWithItemsResponse struct {
	Status          string      `json:"status"`
	Message         string      `json:"message"`
	FirstStockId    string      `json:"first_stock_id"`
	Description     string      `json:"description"`
	FirstStockDate  string      `json:"first_stock_date"`
	TotalFirstStock int         `json:"total_first_stock"`
	Payment         string      `json:"payment"`
	Items           interface{} `json:"items"`
}
