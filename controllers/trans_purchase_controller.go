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
	"github.com/heru-oktafian/api-retail/reports"
	"github.com/heru-oktafian/api-retail/tools"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/helpers"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
	"gorm.io/gorm"
)

// IsPurchaseEditable Function is using to check if the purchase is editable
func IsPurchaseEditable(db *gorm.DB, purchaseID string) (bool, error) {
	var createdAt time.Time
	err := db.Table("purchases").
		Select("created_at").
		Where("id = ?", purchaseID).
		Scan(&createdAt).Error
	if err != nil {
		return false, err
	}

	if time.Since(createdAt) >= time.Hour {
		return false, nil
	}

	return true, nil
}

// CreatePurchase Function is using to create new purchase
func CreatePurchase(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB

	// Ambil informasi dari token
	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)
	generatedID := helpers.GenerateID("PUR")

	// Ambil input dari body
	var input models.PurchaseInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", nil)
	}

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, input.PurchaseDate)
	if err != nil {
		return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", nil)
	}

	// Map ke struct model
	purchase := models.Purchases{
		ID:            generatedID,
		SupplierId:    input.SupplierId,
		BranchID:      branchID,
		UserID:        userID,
		PurchaseDate:  parsedDate,
		TotalPurchase: 0,
		CreatedAt:     nowWIB,
		UpdatedAt:     nowWIB,
	}

	// Simpan purchase
	if err := db.Create(&purchase).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create purchase", err)
	}

	// Buat laporan
	if err := reports.SyncPurchaseReport(db, purchase); err != nil {
		return responses.InternalServerError(c, "Failed to sync purchase report", err)
	}

	_ = reports.AutoCleanupPurchases(db)

	return responses.JSONResponse(c, http.StatusOK, "Purchase created successfully", purchase)
}

// UpdatePurchase Function is using to update purchase
func UpdatePurchase(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	id := c.Param("id")

	// Cari data purchase lama
	var purchase models.Purchases
	if err := db.First(&purchase, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Purchase not found")
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	editable, err := IsPurchaseEditable(db, purchase.ID)
	if err != nil {
		return responses.InternalServerError(c, "Failed to retrieve purchase timestamp", err)
	}
	if !editable {
		return responses.Forbidden(c, "Data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")
	}

	// Gunakan struct input
	var input models.PurchaseInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Cek dan update SupplierID
	if input.SupplierId != "" {
		purchase.SupplierId = input.SupplierId
	}

	// Cek dan update PurchaseDate
	if input.PurchaseDate != "" {
		layout := "2006-01-02"
		parsedDate, err := time.Parse(layout, input.PurchaseDate)
		if err != nil {
			return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", err)
		}
		purchase.PurchaseDate = parsedDate
	}

	// Cek dan update Payment
	if input.Payment != "" {
		purchase.Payment = models.PaymentStatus(input.Payment)
	}

	purchase.UpdatedAt = nowWIB

	// Hitung ulang total dari purchase items
	var items []models.PurchaseItems
	if err := db.Where("purchase_id = ?", id).Find(&items).Error; err != nil {
		return responses.InternalServerError(c, "Failed to retrieve purchase items", err)
	}

	if len(items) == 0 {
		purchase.TotalPurchase = 0
	} else {
		total := 0
		for _, item := range items {
			total += item.SubTotal
		}
		purchase.TotalPurchase = total
	}

	// Simpan perubahan
	if err := db.Save(&purchase).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update purchase", err)
	}

	// Sync report
	if err := reports.SyncPurchaseReport(db, purchase); err != nil {
		return responses.InternalServerError(c, "Failed to sync purchase report", err)
	}

	_ = reports.AutoCleanupPurchases(db)

	return responses.JSONResponse(c, http.StatusOK, "Purchase updated successfully", purchase)
}

// DeletePurchase Function
func DeletePurchase(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Ambil purchase
	var purchase models.Purchases
	if err := db.First(&purchase, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Purchase not found")
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	editable, err := IsPurchaseEditable(db, purchase.ID)
	if err != nil {
		return responses.InternalServerError(c, "Failed to retrieve purchase timestamp", err)
	}
	if !editable {
		return responses.Forbidden(c, "Data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")
	}

	// Ambil item-item dan rollback stok
	var items []models.PurchaseItems
	if err := db.Where("purchase_id = ?", id).Find(&items).Error; err != nil {
		return responses.InternalServerError(c, "Failed to retrieve purchase items", err)
	}

	for _, item := range items {
		// Rollback stok ke produk
		if err := tools.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
			return responses.InternalServerError(c, fmt.Sprintf("Failed to rollback stock for product ID %s", item.ProductId), err)
		}
	}

	// Hapus semua item dari pembelian
	if err := db.Where("purchase_id = ?", id).Delete(&models.PurchaseItems{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete purchase items", err)
	}

	// Hapus laporan transaksi terkait
	if err := db.Where("id = ? AND transaction_type = ?", purchase.ID, models.Purchase).Delete(&models.TransactionReports{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete transaction report", err)
	}

	// Hapus purchase
	if err := db.Delete(&purchase).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete purchase", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Purchase deleted successfully", purchase)
}

// CreatePurchaseItem Function is using to create new purchase item
func CreatePurchaseItem(c *framework.Ctx) error {
	db := config.DB
	var item models.PurchaseItems

	if err := c.BodyParser(&item); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	editable, errr := IsPurchaseEditable(db, item.PurchaseId)
	if errr != nil {
		return responses.InternalServerError(c, "Failed to retrieve purchase timestamp", errr)
	}
	if !editable {
		return responses.Forbidden(c, "Data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")
	}

	// Cek apakah item dengan purchase_id dan product_id sudah ada
	var existing models.PurchaseItems
	err := db.Where("purchase_id = ? AND product_id = ?", item.PurchaseId, item.ProductId).First(&existing).Error
	if err == nil {
		// Sudah ada: update qty dan sub_total
		existing.Qty += item.Qty
		existing.SubTotal = existing.Qty * existing.Price // asumsi pakai harga awal

		if err := db.Save(&existing).Error; err != nil {
			return responses.InternalServerError(c, "Failed to update existing item", err)
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
		if err := tools.RecalculateTotalPurchase(db, item.PurchaseId); err != nil {
			return responses.InternalServerError(c, "Failed to recalculate total purchase", err)
		}

		return responses.JSONResponse(c, http.StatusOK, "Item updated successfully", existing)

	} else if err != gorm.ErrRecordNotFound {
		// Error selain record not found
		return responses.InternalServerError(c, "Failed to check existing item", err)
	}

	// Data belum ada, buat item baru
	if item.ID == "" {
		item.ID = helpers.GenerateID("PIT")
	}
	item.SubTotal = item.Qty * item.Price

	if err := db.Create(&item).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create item", err)
	}
	// Tambah stok
	if err := tools.AddProductStock(db, item.ProductId, item.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to add product stock", err)
	}

	if err := tools.UpdateProductPriceIfHigher(db, item.ProductId, item.Price); err != nil {
		return responses.InternalServerError(c, "Failed to update product price", err)
	}

	if err := tools.RecalculateTotalPurchase(db, item.PurchaseId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total purchase", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item added successfully", item)
}

// Update PurchaseItem is using to update purchase
func UpdatePurchaseItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var existingItem models.PurchaseItems
	if err := db.First(&existingItem, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Item not found")
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	editable, errr := IsPurchaseEditable(db, existingItem.PurchaseId)
	if errr != nil {
		// return c.Status(500).JSON(fiber.Map{"error": "Failed to retrieve purchase timestamp"})
		return responses.InternalServerError(c, "Failed to retrieve purchase timestamp", errr)
	}
	if !editable {
		return responses.Forbidden(c, "Data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")
	}

	var updatedItem models.PurchaseItems
	if err := c.BodyParser(&updatedItem); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Rollback stok lama
	if err := tools.ReduceProductStock(db, existingItem.ProductId, existingItem.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to rollback old stock", err)
	}

	// Tambah stok baru
	if err := tools.AddProductStock(db, updatedItem.ProductId, updatedItem.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to add new stock", err)
	}

	// Update item
	existingItem.ProductId = updatedItem.ProductId
	existingItem.Qty = updatedItem.Qty
	existingItem.Price = updatedItem.Price
	existingItem.SubTotal = updatedItem.Price * updatedItem.Qty

	if err := db.Save(&existingItem).Error; err != nil {
		// return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		return responses.InternalServerError(c, "Failed to update item", err)
	}

	// Update harga produk jika harga item lebih tinggi
	if err := tools.UpdateProductPriceIfHigher(db, updatedItem.ProductId, updatedItem.Price); err != nil {
		return responses.InternalServerError(c, "Failed to update product price", err)
	}

	// Recalculate total & sync
	if err := tools.RecalculateTotalPurchase(db, existingItem.PurchaseId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total purchase", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item updated successfully", existingItem)
}

// Delete PurchaseItem is using to delete purchase
func DeletePurchaseItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var item models.PurchaseItems
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Item not found")
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	editable, errr := IsPurchaseEditable(db, item.PurchaseId)
	if errr != nil {
		return responses.InternalServerError(c, "Failed to retrieve purchase timestamp", errr)
	}
	if !editable {
		return responses.Forbidden(c, "Data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")
	}

	// Subtract stok
	if err := tools.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to reduce product stock", err)
	}

	// Hapus item
	if err := db.Delete(&item).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete item", err)
	}

	// Recalculate total
	if err := tools.RecalculateTotalPurchase(db, item.PurchaseId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total purchase", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item deleted successfully", item)
}

// Get All Purchases tampilkan semua purchase
func GetAllPurchases(c *framework.Ctx) error {
	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	branchID, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10                  // Tetapkan limit ke 10 data per halaman
	offset := (page - 1) * limit // Hitung offset berdasarkan halaman dan limit

	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	var purchases []models.AllPurchases
	var total int64

	// Mulai bangun query
	query := config.DB.Table("purchases pur").
		Select("pur.id, pur.supplier_id, sup.name AS supplier_name, pur.purchase_date, pur.total_purchase, pur.payment").
		Joins("LEFT JOIN suppliers sup ON sup.id = pur.supplier_id").
		Where("pur.branch_id = ? AND pur.total_purchase > 0", branchID)

	// Filter bulan
	startDate, err := time.Parse("2006-01", month)
	if err != nil {
		return responses.BadRequest(c, "Invalid month format", err)
	}
	endDate := startDate.AddDate(0, 1, 0)
	query = query.Where("pur.purchase_date >= ? AND pur.purchase_date < ?", startDate, endDate)

	// Filter search jika ada
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(sup.name) LIKE ?", "%"+search+"%")
	}

	// Hitung total
	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Get purchase failed", err)
	}

	// Ambil data paginasi
	if err := query.Order("pur.created_at DESC").Offset(offset).Limit(limit).Scan(&purchases).Error; err != nil {
		return responses.InternalServerError(c, "Get purchases failed", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedPurchasesData []models.PurchaseDetailResponse
	for _, purchase := range purchases {
		formattedPurchasesData = append(formattedPurchasesData, models.PurchaseDetailResponse{
			ID:            purchase.ID,
			SupplierId:    purchase.SupplierId,
			SupplierName:  purchase.SupplierName,
			PurchaseDate:  utils.FormatIndonesianDate(purchase.PurchaseDate), // Format tanggal di sini
			TotalPurchase: purchase.TotalPurchase,
			Payment:       string(purchase.Payment),
		})
	}

	return responses.JSONResponseGetAll(c, http.StatusOK, "Purchases retrieved successfully", search, int(total), page, totalPages, limit, formattedPurchasesData)
}

// GetAllPurchaseItems tampilkan semua item berdasarkan purchase_id tanpa pagination
func GetAllPurchaseItems(c *framework.Ctx) error {
	// Get purchase id dari param
	purchaseID := c.Param("id")

	// Parsing body JSON ke struct
	var PurchaseItems []models.AllPurchaseItems

	// Query dasar
	query := config.DB.Table("purchase_items pit").
		Select("pit.id, pit.purchase_id, pit.product_id, pro.name AS product_name, pit.price, pit.qty, pro.unit_id, un.name AS unit_name, pit.sub_total, pit.expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pit.purchase_id = ?", purchaseID).
		Order("pro.name ASC")

	// Eksekusi query
	if err := query.Scan(&PurchaseItems).Error; err != nil {
		return responses.InternalServerError(c, "Failed to get purchase items", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Items retrieved successfully", PurchaseItems)
}

// GetPurchaseWithItems menampilkan satu purchase beserta semua item-nya
func GetPurchaseWithItems(c *framework.Ctx) error {
	db := config.DB

	// Ambil ID pembelian dari parameter URL
	purchaseID := c.Param("id")

	// Struct untuk data utama purchase
	var purchase models.AllPurchases

	// Ambil data purchase dengan LEFT JOIN ke suppliers
	err := db.Table("purchases pur").
		Select("pur.id, pur.supplier_id, sup.name AS supplier_name, pur.purchase_date, pur.total_purchase, pur.payment").
		Joins("LEFT JOIN suppliers sup ON sup.id = pur.supplier_id").
		Where("pur.id = ?", purchaseID).
		Scan(&purchase).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to get purchase", err)
	}

	// Ambil item pembelian terkait
	var items []models.AllPurchaseItems
	err = db.Table("purchase_items pit").
		Select("pit.id, pit.purchase_id, pit.product_id, pro.name AS product_name, pit.unit_id AS unit_id, un.name AS unit_name, pit.price, pit.qty, pit.sub_total, pit.expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN units un ON un.id = pit.unit_id").
		Where("pit.purchase_id = ?", purchaseID).
		Order("pro.name ASC").
		Scan(&items).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to get purchase items", err)
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formatedPurchaseItems []models.FormatedPurchaseItems
	for _, purItems := range items {
		formatedPurchaseItems = append(formatedPurchaseItems, models.FormatedPurchaseItems{
			ID:          purItems.ID,
			ProductId:   purItems.ProductId,
			ProductName: purItems.ProductName,
			UnitId:      purItems.UnitId,
			UnitName:    purItems.UnitName,
			Price:       purItems.Price,
			Qty:         purItems.Qty,
			SubTotal:    purItems.SubTotal,
			ExpiredDate: utils.FormatIndonesianDate(purItems.ExpiredDate), // Format tanggal di sini
		})
	}

	// Format tanggal secara manual untuk respons ini
	// Menggunakan helper FormatIndonesianDate yang sudah kita buat
	formattedPurchaseDate := utils.FormatIndonesianDate(purchase.PurchaseDate)

	// Buat objek respons menggunakan struct PurchaseItemResponse yang baru
	// dan isi field-fieldnya
	responseDetail := models.PurchaseDetailWithItemsResponse{
		ID:            purchase.ID,
		SupplierId:    purchase.SupplierId,
		SupplierName:  purchase.SupplierName,
		PurchaseDate:  formattedPurchaseDate, // Gunakan tanggal yang sudah diformat
		TotalPurchase: purchase.TotalPurchase,
		Payment:       string(purchase.Payment),
		Items:         formatedPurchaseItems,
	}

	// Panggil JSONResponse yang sudah ada, meneruskan PurchaseItemResponse sebagai 'data'
	return responses.JSONResponse(c, http.StatusOK, "Purchase retrieved successfully", responseDetail)
}

// CreatePurchaseTransaction controller
func CreatePurchaseTransaction(c *framework.Ctx) error {
	nowWIB := time.Now().In(utils.Location)

	subscriptionType, _ := middlewares.GetClaimsToken(c.Request, "subscription_type")
	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)

	db := config.DB
	var req models.PurchaseTransactionRequest
	err := c.BodyParser(&req)
	if err != nil {
		return responses.BadRequest(c, "Invalid request body", err)
	}

	if req.Purchase.Payment == "" {
		req.Purchase.Payment = "paid_by_cash"
	}

	req.Purchase.UserID = userID
	req.Purchase.BranchID = branchID

	// --- VALIDASI INPUT ---
	err = utils.ValidateStruct(req.Purchase)
	if err != nil {
		return responses.BadRequest(c, "Validation failed for purchase input", err)
	}

	for _, item := range req.PurchaseItems {
		err = utils.ValidateStruct(item)
		if err != nil {
			return responses.BadRequest(c, "Validation failed for one or more purchase items", err)
		}
	}
	// --- AKHIR VALIDASI INPUT ---

	var purchaseDate time.Time
	if req.Purchase.PurchaseDate == "" {
		purchaseDate = nowWIB
	} else {
		purchaseDate, err = time.Parse("2006-01-02", req.Purchase.PurchaseDate)
		if err != nil {
			return responses.BadRequest(c, "Invalid purchase_date format. Please use `YYYY-MM-DD`.", err)
		}
	}

	purchase := models.Purchases{
		SupplierId:   req.Purchase.SupplierId,
		PurchaseDate: purchaseDate,
		BranchID:     req.Purchase.BranchID,
		Payment:      req.Purchase.Payment,
		UserID:       req.Purchase.UserID,
	}

	// --- Proses Penyimpanan Data ---
	tx := db.Begin()
	if tx.Error != nil {
		return responses.InternalServerError(c, "Failed to begin database transaction", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	purchaseID := helpers.GenerateID("PUR")
	purchase.ID = purchaseID
	purchase.CreatedAt = nowWIB
	purchase.UpdatedAt = nowWIB

	var calculatedTotalPurchase int
	var purchaseItemsToCreate []models.PurchaseItems
	var purchaseItemsForResponse []models.PurchaseItemResponse // <--- Slice baru untuk data respons

	// Mendapatkan nama supplier (di luar loop item untuk efisiensi)
	var supplier models.Supplier
	err = tx.Where("id = ?", req.Purchase.SupplierId).First(&supplier).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return responses.NotFound(c, fmt.Sprintf("Supplier with ID %s not found", req.Purchase.SupplierId))
		}
		return responses.InternalServerError(c, "Failed to retrieve supplier details", err)
	}

	// var stockTracksToCreate []models.StockTracks

	for i := range req.PurchaseItems {
		parsedExpiredDate, err := time.Parse("2006-01-02", req.PurchaseItems[i].ExpiredDate)
		if err != nil {
			tx.Rollback()
			return responses.BadRequest(c, fmt.Sprintf("Invalid expired_date format for product %s. Please use `YYYY-MM-DD`.", req.PurchaseItems[i].ProductId), err)
		}

		var product models.Product
		err = tx.Where("id = ?", req.PurchaseItems[i].ProductId).First(&product).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Product with ID %s not found", req.PurchaseItems[i].ProductId))
			}
			return responses.InternalServerError(c, "Failed to retrieve product details", err)
		}

		// Mendapatkan nama unit (sesuai unit_id yang diinput)
		var unit models.Unit
		err = tx.Where("id = ?", req.PurchaseItems[i].UnitId).First(&unit).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Unit with ID %s not found", req.PurchaseItems[i].UnitId))
			}
			return responses.InternalServerError(c, "Failed to retrieve unit details", err)
		}

		// --- Logika Konversi Satuan ---
		var conversionValue int = 1
		if req.PurchaseItems[i].UnitId != product.UnitId {
			var unitConversion models.UnitConversion
			err = tx.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
				req.PurchaseItems[i].ProductId,
				req.PurchaseItems[i].UnitId,
				product.UnitId,
				purchase.BranchID,
			).First(&unitConversion).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					conversionValue = 1
				} else {
					tx.Rollback()
					return responses.InternalServerError(c, "Failed to retrieve unit conversion details", err)
				}
			} else {
				conversionValue = unitConversion.ValueConv
			}
		}
		actualQtyToAdd := req.PurchaseItems[i].Qty * conversionValue
		// --- Akhir Logika Konversi Satuan ---

		// --- Perhitungan Price dan SubTotal Otomatis ---
		itemPrice := req.PurchaseItems[i].Price * conversionValue
		itemSubTotal := itemPrice * req.PurchaseItems[i].Qty

		purchaseItemDB := models.PurchaseItems{
			ID:          helpers.GenerateID("PIT"),
			PurchaseId:  purchaseID,
			ProductId:   req.PurchaseItems[i].ProductId,
			UnitId:      req.PurchaseItems[i].UnitId,
			Price:       itemPrice,
			Qty:         req.PurchaseItems[i].Qty,
			SubTotal:    itemSubTotal,
			ExpiredDate: parsedExpiredDate,
		}
		purchaseItemsToCreate = append(purchaseItemsToCreate, purchaseItemDB)

		// --- Siapkan data untuk respons ---
		purchaseItemResp := models.PurchaseItemResponse{
			ID:          purchaseItemDB.ID,
			ProductID:   purchaseItemDB.ProductId,
			ProductName: product.Name, // <--- Ambil dari product yang sudah diambil
			UnitID:      purchaseItemDB.UnitId,
			UnitName:    unit.Name, // <--- Ambil dari unit yang sudah diambil
			Price:       purchaseItemDB.Price,
			Qty:         purchaseItemDB.Qty,
			SubTotal:    purchaseItemDB.SubTotal,
			ExpiredDate: parsedExpiredDate.Format("02 January 2006"), // <--- Format tanggal
		}
		purchaseItemsForResponse = append(purchaseItemsForResponse, purchaseItemResp)
		// --- Akhir persiapan data respons ---

		// --- Tambah stok dan cek/update expired_date ---
		updates := map[string]interface{}{
			"stock": product.Stock + actualQtyToAdd,
		}

		if parsedExpiredDate.Before(product.ExpiredDate) {
			updates["expired_date"] = parsedExpiredDate
		}

		err = tx.Model(&models.Product{}).Where("id = ?", product.ID).Updates(updates).Error
		if err != nil {
			tx.Rollback()
			return responses.InternalServerError(c, fmt.Sprintf("Failed to update product details (stock/expired_date) for product %s", product.Name), err)
		}
		// --- Akhir tambah stok dan cek/update expired_date ---
		calculatedTotalPurchase += itemSubTotal
	}

	purchase.TotalPurchase = calculatedTotalPurchase

	err = tx.Create(&purchase).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create purchase", err)
	}

	err = tx.CreateInBatches(&purchaseItemsToCreate, len(purchaseItemsToCreate)).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create purchase items", err)
	}

	transactionReportID := helpers.GenerateID("TRX")
	transactionReport := models.TransactionReports{
		ID:              transactionReportID,
		TransactionType: models.Purchase,
		UserID:          purchase.UserID,
		BranchID:        purchase.BranchID,
		Total:           purchase.TotalPurchase,
		Payment:         purchase.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	err = tx.Create(&transactionReport).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create transaction report for purchase", err)
	}

	if subscriptionType == "quota" {
		var branch models.Branch
		err = tx.Where("id = ?", req.Purchase.BranchID).First(&branch).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Branch with ID %s not found", req.Purchase.BranchID))
			}
			return responses.InternalServerError(c, "Failed to retrieve branch details for quota update", err)
		}

		if branch.Quota > 0 {
			branch.Quota -= 1
			err = tx.Save(&branch).Error
			if err != nil {
				tx.Rollback()
				return responses.InternalServerError(c, fmt.Sprintf("Failed to update quota for branch %s", branch.BranchName), err)
			}
		} else {
			tx.Rollback()
			return responses.BadRequest(c, fmt.Sprintf("No quota available for branch %s", branch.BranchName), errors.New("quota exceeded"))
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return responses.InternalServerError(c, "Failed to commit database transaction", err)
	}

	// --- Akhir: Mengkonstruksi Objek Respon ---
	response := models.PurchaseResponse{
		ID:            purchase.ID,
		SupplierID:    purchase.SupplierId,
		SupplierName:  supplier.Name,                          // Ambil nama supplier yang sudah diambil
		PurchaseDate:  purchaseDate.Format("02 January 2006"), // Format tanggal pembelian
		TotalPurchase: purchase.TotalPurchase,
		Payment:       purchase.Payment,
		Items:         purchaseItemsForResponse, // Sertakan slice item yang sudah disiapkan
	}
	// --- Akhir Mengkonstruksi Objek Respon ---

	return responses.JSONResponse(c, http.StatusOK, "Purchase transaction created successfully", response)
}

// GetFixedPrice menghitung harga produk setelah konversi satuan
// Endpoint ini akan dipanggil dengan query parameters: product_id, init_id, final_id
// Contoh: GET /api/fixed-price?product_id=PRD123&init_id=UNT_BOX&final_id=UNT_PCS
func GetFixedPrice(c *framework.Ctx) error {
	db := config.DB
	var req models.GetFixedPriceRequest

	// Parsing query parameters
	if err := c.QueryParser(&req); err != nil {
		return responses.BadRequest(c, "Invalid query parameters", err)
	}

	// Get BranchID from token (jika harga dan konversi spesifik per cabang)
	// Jika konversi satuan dan harga tidak spesifik cabang, Anda bisa hapus baris ini dan filter branch_id di query
	branchID, _ := middlewares.GetBranchID(c.Request)
	if branchID == "" {
		return responses.Unauthorized(c, "Branch ID not found in token. Unauthorized.")
	}

	// antangin sachet 12
	// tolak angin 12
	// amlodipin strip 10 box
	// grantusif 10 box
	// asamefenamat 10 box

	// --- 1. Dapatkan Product dan PurchasePrice-nya ---
	var product models.Product
	err := db.Where("id = ? AND branch_id = ?", req.ProductID, branchID).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responses.NotFound(c, fmt.Sprintf("Product with ID %s not found in branch %s", req.ProductID, branchID))
		}
		return responses.InternalServerError(c, "Failed to retrieve product details", err)
	}

	// --- 2. Dapatkan UnitConversion Value ---
	var conversionValue int = 1 // Default to 1 if no conversion or no explicit FinalID
	var unitConversion models.UnitConversion

	// Asumsi: jika final_id dari request adalah sama dengan init_id, atau
	// jika init_id sudah merupakan unit dasar, maka value_conv adalah 1.
	// Jika req.InitID sama dengan req.FinalID, maka konversinya 1.
	// Jika UnitId di Product adalah satuan dasar, maka FinalId di UnitConversion harus merujuk ke UnitId di Product.
	// Kita akan mencari konversi dari init_id ke final_id yang diberikan di parameter.

	// Hanya cari konversi jika init_id tidak sama dengan final_id,
	// dan juga init_id tidak sama dengan unit default produk (jika diasumsikan sebagai satuan dasar).
	if req.InitID != req.FinalID { // Konversi hanya diperlukan jika satuan awal dan akhir berbeda
		err = db.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
			req.ProductID,
			req.InitID,
			req.FinalID,
			branchID,
		).First(&unitConversion).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Jika konversi spesifik tidak ditemukan, kita bisa memilih untuk:
				// a) Mengembalikan error (lebih ketat)
				// b) Mengasumsikan value_conv adalah 1 (lebih longgar, mungkin berarti unit_id == final_id secara implisit)
				// Untuk endpoint ini, saya akan mengembalikan error karena permintaan Anda eksplisit.
				return responses.NotFound(c, fmt.Sprintf("Unit conversion from %s to %s for product %s not found in branch %s", req.InitID, req.FinalID, req.ProductID, branchID))
			}
			return responses.InternalServerError(c, "Failed to retrieve unit conversion details", err)
		}
		conversionValue = unitConversion.ValueConv
	}

	// --- 3. Hitung FixPrice ---
	fixPrice := product.PurchasePrice * conversionValue

	// --- 4. Buat Respon ---
	response := models.FixedPriceResponse{
		FixPrice: fixPrice,
	}

	return responses.JSONResponse(c, http.StatusOK, "Fixed price calculated successfully", response)
}

// GetUnitsByProductIdRequest merepresentasikan body request untuk endpoint GET ini
type GetUnitsByProductIdRequest struct {
	ProductID string `json:"product_id" validate:"required"`
}

// GetProductUnitsWithConvertedPrices mengambil daftar unit yang tersedia untuk dibeli
// beserta harga pembelian yang sudah dikonversi
// Endpoint: GET /api/product-units (dengan body request)
func GetProductUnitsWithConvertedPrices(c *framework.Ctx) error {
	db := config.DB
	var req GetUnitsByProductIdRequest

	if err := c.BodyParser(&req); err != nil {
		// return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 	"message": "Invalid request body",
		// 	"error":   err.Error(),
		// })
		return responses.BadRequest(c, "Invalid request body", err)
	}

	if err := utils.ValidateStruct(req); err != nil {
		return responses.BadRequest(c, "Validation failed for product ID", err)
	}

	branchID, _ := middlewares.GetBranchID(c.Request)
	if branchID == "" {
		return responses.Unauthorized(c, "Branch ID not found in token")
	}

	// 1. Dapatkan detail Product, khususnya Product.PurchasePrice (harga dasar) dan Product.UnitId (unit dasar)
	var product models.Product
	err := db.Where("id = ? AND branch_id = ?", req.ProductID, branchID).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return responses.NotFound(c, fmt.Sprintf("Product with ID %s not found in branch %s", req.ProductID, branchID))
		}
		return responses.InternalServerError(c, "Failed to retrieve product details", err)
	}

	basePurchasePrice := product.PurchasePrice // Harga dasar (per unit dasar produk)
	baseUnitID := product.UnitId               // ID unit dasar produk

	// Map untuk menyimpan harga yang sudah dihitung untuk setiap unit ID
	// Ini akan menghindari duplikasi unit di respons akhir dan memudahkan lookup
	calculatedPrices := make(map[string]models.ProductUnitResponseItem)

	// Tambahkan unit dasar produk itu sendiri sebagai opsi pertama
	// Ini adalah harga acuan tanpa konversi
	calculatedPrices[baseUnitID] = models.ProductUnitResponseItem{
		UnitId:        baseUnitID,
		PurchasePrice: basePurchasePrice,
	}

	// 2. Dapatkan semua UnitConversion yang terkait dengan ProductID ini
	var unitConversions []models.UnitConversion
	err = db.Where("product_id = ? AND branch_id = ?", req.ProductID, branchID).Find(&unitConversions).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return responses.InternalServerError(c, "Failed to retrieve unit conversions", err)
	}

	// Kumpulkan semua ID unit yang perlu dicari namanya (dari produk dasar, init_id, dan final_id konversi)
	unitIDsToFetch := []string{baseUnitID} // Mulai dengan unit dasar produk
	for _, uc := range unitConversions {
		unitIDsToFetch = append(unitIDsToFetch, uc.InitId)
		unitIDsToFetch = append(unitIDsToFetch, uc.FinalId)
	}

	// Dapatkan nama-nama unit yang relevan sekaligus
	unitNames := make(map[string]string)
	var units []models.Unit
	if len(unitIDsToFetch) > 0 {
		err = db.Where("id IN (?) AND branch_id = ?", unitIDsToFetch, branchID).Find(&units).Error
		if err != nil {
			return responses.InternalServerError(c, "Failed to retrieve unit names", err)
		}
		for _, u := range units {
			unitNames[u.ID] = u.Name
		}
	}

	// Set nama untuk unit dasar
	if item, ok := calculatedPrices[baseUnitID]; ok {
		item.UnitName = unitNames[baseUnitID]
		calculatedPrices[baseUnitID] = item
	}

	// 3. Iterasi melalui UnitConversions dan hitung harga untuk setiap unit yang relevan
	// Kita harus mempertimbangkan konversi 'maju' (dari unit besar ke unit kecil)
	// dan konversi 'mundur' (dari unit kecil ke unit besar)
	// tergantung bagaimana ValueConv didefinisikan.

	// Untuk setiap konversi, tentukan harga untuk InitId dan FinalId
	for _, uc := range unitConversions {
		// Kasus 1: Konversi dari unit yang lebih besar ke unit yang lebih kecil (misal: Box ke Pcs, Strip ke Sachet)
		// Jika init_id adalah unit yang lebih besar, dan final_id adalah unit yang lebih kecil,
		// maka harga final_id = harga init_id / value_conv.
		// Asumsi: ValueConv selalu > 0.

		// Untuk menyederhanakan, kita akan selalu mengonversi ke/dari unit dasar (`baseUnitID`)
		// jika memungkinkan, untuk mendapatkan harga yang konsisten.

		// Skenario A: Konversi dari unit dasar ke unit lain (FinalId)
		// Contoh: 1 PCS (base) = 12 SACHET.
		// Product.PurchasePrice (harga per PCS) = 3000.
		// Harga Sachet = 3000 / 12 = 250.
		if uc.InitId == baseUnitID {
			if uc.ValueConv > 0 {
				priceForFinalID := basePurchasePrice / uc.ValueConv
				if existing, ok := calculatedPrices[uc.FinalId]; !ok || existing.PurchasePrice > priceForFinalID {
					// Hanya tambahkan/perbarui jika unit belum ada atau harga yang baru lebih murah
					// (Ini bisa jadi diperlukan jika ada beberapa jalur konversi, ambil yang paling optimal)
					calculatedPrices[uc.FinalId] = models.ProductUnitResponseItem{
						UnitId:        uc.FinalId,
						UnitName:      unitNames[uc.FinalId],
						PurchasePrice: priceForFinalID,
					}
				}
			}
		}

		// Skenario B: Konversi dari unit lain (InitId) ke unit dasar (FinalId)
		// Contoh: 1 BOX = 10 PCS (base).
		// Product.PurchasePrice (harga per PCS) = 250.
		// Harga Box = 250 * 10 = 2500.
		if uc.FinalId == baseUnitID {
			if uc.ValueConv > 0 {
				priceForInitID := basePurchasePrice * uc.ValueConv
				if existing, ok := calculatedPrices[uc.InitId]; !ok || existing.PurchasePrice > priceForInitID {
					// Hanya tambahkan/perbarui jika unit belum ada atau harga yang baru lebih murah
					calculatedPrices[uc.InitId] = models.ProductUnitResponseItem{
						UnitId:        uc.InitId,
						UnitName:      unitNames[uc.InitId],
						PurchasePrice: priceForInitID,
					}
				}
			}
		}
	}

	// 4. Ubah map menjadi slice untuk respons akhir
	var finalResponseItems []models.ProductUnitResponseItem
	for _, item := range calculatedPrices {
		if item.UnitName != "" { // Pastikan unit punya nama yang ditemukan
			finalResponseItems = append(finalResponseItems, item)
		}
	}

	// Opsional: Urutkan hasil jika diinginkan (misal berdasarkan nama unit atau harga)
	// sort.Slice(finalResponseItems, func(i, j int) bool {
	// 	return finalResponseItems[i].PurchasePrice < finalResponseItems[j].PurchasePrice
	// })

	// 5. Kembalikan respons sukses
	return responses.JSONResponse(c, http.StatusOK, fmt.Sprintf("Units retrieved successfully for Product ID %s", req.ProductID), finalResponseItems)
}
