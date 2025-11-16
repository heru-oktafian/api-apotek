package controllers

import (
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

// Get Mobile Opnames menampilkan semua opname yang disajikan untuk pengguna mobile
func GetAllMobileOpnames(c *framework.Ctx) error {
	branchID, _ := middlewares.GetBranchID(c.Request)
	var rawOpnames []models.OpnameQueryResult // Gunakan struct untuk menampung hasil query mentah

	// Query dasar (opname_date tanpa TO_CHAR di SQL)
	query := config.DB.Table("opnames pur").
		Select("pur.id, pur.description, pur.opname_date, 'Rp. ' || TO_CHAR(pur.total_opname, 'FM999G999G999') AS total_opname").
		Where("pur.branch_id = ?", branchID).
		Order("pur.created_at DESC")

	if err := query.Scan(&rawOpnames).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Pengambilan opnames gagal", "Gagal mengambil data Opname")
	}

	// Inisialisasi slice untuk hasil akhir yang diformat
	var formattedOpnames []models.AllOpnameMobiles

	// Iterasi hasil query mentah dan format opname_date
	for _, op := range rawOpnames {
		formattedOpnames = append(formattedOpnames, models.AllOpnameMobiles{
			ID:          op.ID,
			Description: op.Description,
			OpnameDate:  utils.FormatIndonesianDate(op.OpnameDate), // <--- Gunakan helper di sini!
			TotalOpname: op.TotalOpname,
		})
	}

	return responses.JSONResponse(c, http.StatusOK, "Data opname berhasil diambil", formattedOpnames)
}

// Get Mobile Opnames Actives menampilkan semua opname yang disajikan untuk pengguna mobile dengan status aktif
func GetAllActiveMobileOpnames(c *framework.Ctx) error {
	branchID, _ := middlewares.GetBranchID(c.Request)
	var rawOpnames []models.OpnameQueryResult // Gunakan struct untuk menampung hasil query mentah

	// Query dasar (opname_date tanpa TO_CHAR di SQL)
	query := config.DB.Table("opnames pur").
		Select("pur.id, pur.description, pur.opname_date, 'Rp. ' || TO_CHAR(pur.total_opname, 'FM999G999G999') AS total_opname").
		Where("pur.branch_id = ? AND pur.opname_status = 'active' ", branchID).
		Order("pur.created_at DESC")

	if err := query.Scan(&rawOpnames).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Pengambilan opnames gagal", "Gagal mengambil data Opname")
	}

	// Inisialisasi slice untuk hasil akhir yang diformat
	var formattedOpnames []models.AllOpnameMobiles

	// Iterasi hasil query mentah dan format opname_date
	for _, op := range rawOpnames {
		formattedOpnames = append(formattedOpnames, models.AllOpnameMobiles{
			ID:          op.ID,
			Description: op.Description,
			OpnameDate:  utils.FormatIndonesianDate(op.OpnameDate), // <--- Gunakan helper di sini!
			TotalOpname: op.TotalOpname,
		})
	}

	return responses.JSONResponse(c, http.StatusOK, "Data opname berhasil diambil", formattedOpnames)
}

// GetMobileOpnameItemDetails adalah fungsi menammpilkan semua item berdasarkan product_name tanpa pagination
func GetMobileOpnameItemDetails(c *framework.Ctx) error {
	// Get branch id
	branchID, _ := middlewares.GetBranchID(c.Request)

	// Parsing body JSON ke struct
	var OpnameItems []models.AllOpnameItemDetails

	// Query dasar
	query := config.DB.Table("opname_items pit").
		Select("pit.id, pit.opname_id, pit.product_id, pro.name AS product_name, pit.price, (pit.qty - pit.qty_exist) AS qty_adjustment, (pit.sub_total - pit.sub_total_exist) AS sub_adjustment, pit.expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN opnames opn ON opn.id = pit.opname_id").
		Where("opn.branch_id = ?", branchID).
		Order("pro.name ASC")

	// Eksekusi query
	if err := query.Scan(&OpnameItems).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Pengambilan item gagal", "Gagal mengambil data item Opname")
	}

	return responses.JSONResponse(c, http.StatusOK, "Item berhasil ditampilkan", OpnameItems)
}

// GetMobileOpnameItemsGlimpse adalah fungsi menammpilkan 5 item berdasarkan product_name tanpa pagination
func GetMobileOpnameItemsGlimpse(c *framework.Ctx) error {
	// Get branch id
	branchID, _ := middlewares.GetBranchID(c.Request)

	// Parsing body JSON ke struct
	var OpnameItems []models.AllOpnameItemDetails

	// Query dasar
	query := config.DB.Table("opname_items pit").
		Select("pit.id, pit.opname_id, pit.product_id, pro.name AS product_name, pit.price, (pit.qty - pit.qty_exist) AS qty_adjustment, (pit.sub_total - pit.sub_total_exist) AS sub_adjustment, pit.expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN opnames opn ON opn.id = pit.opname_id").
		Where("opn.branch_id = ?", branchID).
		Limit(5)

	// Eksekusi query
	if err := query.Scan(&OpnameItems).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Pengambilan item gagal", "Gagal mengambil data item Opname")
	}

	return responses.JSONResponse(c, http.StatusOK, "Item berhasil ditampilkan", OpnameItems)
}

// Get All Opnames tampilkan semua opname
func GetAllOpnames(c *framework.Ctx) error {
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

	var opnames []models.AllOpnames
	var total int64

	// Query dasar
	query := config.DB.Table("opnames pur").
		Select("pur.id, pur.description, TO_CHAR(pur.opname_date, 'DD-MM-YYYY') AS opname_date, pur.total_opname").
		Where("pur.branch_id = ?", branch_id).
		Order("pur.created_at DESC")

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(pur.description) LIKE ?", "%"+search+"%")
	}

	// Hitung total opname yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Pengambilan opname gagal", "Gagal menghitung stok awal")
	}

	// Ambil data dengan pagination
	if err := query.Offset(offset).Limit(limit).Scan(&opnames).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Pengambilan opname gagal", "Gagal mengambil data stok awal")
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Kembalikan hasil response tanpa nested "data"
	return JSONResponseFlat(c, http.StatusOK, "Data Opname berhasil diambil", map[string]interface{}{
		"per_page":     limit,
		"current_page": page,
		"search":       search,
		"total":        total,
		"total_pages":  totalPages,
		"data":         opnames,
	})
}

func JSONResponseFlat(c *framework.Ctx, status int, message string, payload map[string]interface{}) error {
	payload["status"] = "success"
	payload["message"] = message
	return c.JSON(status, payload)
}

// CreateOpname Function
func CreateOpname(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB

	// Ambil informasi dari token
	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)
	generatedID := helpers.GenerateID("OPN")

	// Ambil input dari body
	var input models.OpnameInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Input tidak valid", err)
	}

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, input.OpnameDate)
	if err != nil {
		return responses.BadRequest(c, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
	}

	// Map ke struct model
	opname := models.Opnames{
		ID:          generatedID,
		Description: input.Description,
		BranchID:    branchID,
		UserID:      userID,
		OpnameDate:  parsedDate,
		TotalOpname: 0,
		CreatedAt:   nowWIB,
		UpdatedAt:   nowWIB,
	}

	// Simpan opname
	if err := db.Create(&opname).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal menyimpan data opname", err.Error())
	}

	// Buat laporan
	if err := reports.SyncOpnameReport(db, opname); err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat laporan opname", err.Error())
	}

	_ = tools.AutoCleanupOpnames(db)

	return responses.JSONResponse(c, http.StatusOK, "Opname berhasil dibuat", opname)
}

// UpdateOpnameByID Function
func UpdateOpnameByID(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	id := c.Param("id")

	// Cari data opname lama
	var opname models.Opnames
	if err := db.First(&opname, "id = ?", id).Error; err != nil {
		// return c.Status(404).JSON(fiber.Map{"error": "Opname tidak ditemukan"})
		return responses.NotFound(c, "Opname tidak ditemukan")
	}

	// Gunakan struct input
	var input models.OpnameInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Input tidak valid", err)
	}

	// Cek dan update OpnameDate
	if input.OpnameDate != "" {
		layout := "2006-01-02"
		parsedDate, err := time.Parse(layout, input.OpnameDate)
		if err != nil {
			return responses.BadRequest(c, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
		}
		opname.OpnameDate = parsedDate
	}

	// Cek dan update Payment
	if input.Payment != "" {
		opname.Payment = models.PaymentStatus(input.Payment)
	}

	opname.UpdatedAt = nowWIB

	// Hitung ulang total dari opname items
	var items []models.OpnameItems
	if err := db.Where("opname_id = ?", id).Find(&items).Error; err != nil {
		return responses.InternalServerError(c, "Gagal mengambil item opname", err)
	}

	if len(items) == 0 {
		opname.TotalOpname = 0
	} else {
		total := 0
		for _, item := range items {
			total += item.SubTotal
		}
		opname.TotalOpname = total
	}

	// Cek dan update Description
	if input.Description != "" {
		opname.Description = input.Description
	}

	// Simpan perubahan
	if err := db.Save(&opname).Error; err != nil {
		return responses.InternalServerError(c, "Gagal memperbarui opname", err)
	}

	// Sync report
	if err := reports.SyncOpnameReport(db, opname); err != nil {
		return responses.InternalServerError(c, "Gagal menyinkronkan laporan opname", err)
	}

	_ = tools.AutoCleanupOpnames(db)

	return responses.JSONResponse(c, http.StatusOK, "Opname berhasil diperbarui", opname)
}

// DeleteOpnameByID Function
func DeleteOpnameByID(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Ambil opname
	var opname models.Opnames
	if err := db.First(&opname, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Opname tidak ditemukan")
	}

	// Ambil item-item dan rollback stok
	var items []models.OpnameItems
	if err := db.Where("opname_id = ?", id).Find(&items).Error; err != nil {
		return responses.InternalServerError(c, "Gagal mengambil item opname", err)
	}

	for _, item := range items {
		// Kosongkan stok ke produk
		if err := tools.ZeroProductStock(db, item.ProductId, item.Qty); err != nil {
			return responses.InternalServerError(c, "Gagal mengosongkan stok produk", err)
		}
	}

	// Hapus semua item dari pembelian
	if err := db.Where("opname_id = ?", id).Delete(&models.OpnameItems{}).Error; err != nil {
		return responses.InternalServerError(c, "Gagal menghapus item opname", err)
	}

	// Hapus laporan transaksi terkait
	if err := db.Where("id = ? AND transaction_type = ?", opname.ID, models.Opname).Delete(&models.TransactionReports{}).Error; err != nil {
		return responses.InternalServerError(c, "Gagal menghapus laporan transaksi", err)
	}

	// Hapus opname
	if err := db.Delete(&opname).Error; err != nil {
		return responses.InternalServerError(c, "Gagal menghapus opname", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Opname berhasil dihapus", opname)
}

// GetOpnameWithItems menampilkan satu opname beserta semua item-nya
func GetOpnameWithItems(c *framework.Ctx) error {
	db := config.DB

	// Ambil ID pembelian dari parameter URL
	opnameID := c.Param("id")

	// Struct untuk data utama opname
	var opname models.AllOpnames

	// Ambil data opname
	err := db.Table("opnames pur").
		Select("pur.id, pur.description, TO_CHAR(pur.opname_date, 'DD-MM-YYYY') AS opname_date, pur.total_opname, pur.payment").
		Where("pur.id = ?", opnameID).
		Scan(&opname).Error

	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mendapatkan opname", err.Error())
	}

	// Ambil item pembelian terkait
	var items []models.AllOpnameItemMobiles
	err = db.Table("opname_items pit").
		Select("pit.id, pit.opname_id, pit.product_id, pro.name AS product_name, pit.price, pit.qty, pit.sub_total, TO_CHAR(pit.expired_date, 'DD-MM-YYYY') AS expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Where("pit.opname_id = ?", opnameID).
		Order("pro.name ASC").
		Scan(&items).Error

	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mendapatkan item Opname", err.Error())
	}

	return responses.JSONResponse(c, http.StatusOK, "Opname berhasil diambil", map[string]interface{}{
		"id":           opnameID,
		"description":  opname.Description,
		"opname_date":  opname.OpnameDate,
		"total_opname": opname.TotalOpname,
		"items":        items,
	})
}

// CreateOpnameItem Function
func CreateOpnameItem(c *framework.Ctx) error {
	db := config.DB
	var input tools.CreateOpnameItemInput

	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Masukan tidak valid: "+err.Error(), err)
	}

	// Ambil data produk untuk mendapatkan price, stock, dan purchase_price
	var product models.Product
	if err := db.Where("id = ?", input.ProductId).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return responses.NotFound(c, "Produk tidak ditemukan")
		}
		return responses.InternalServerError(c, "Gagal mengambil data produk: "+err.Error(), err)
	}

	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, input.ExpiredDate)
	if err != nil {
		return responses.BadRequest(c, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
	}

	if product.Stock > 0 {
		if product.ExpiredDate.After(parsedDate) {
			product.ExpiredDate = parsedDate
		}
	} else {
		product.ExpiredDate = parsedDate
	}
	if err := db.Model(&product).Update("expired_date", product.ExpiredDate).Error; err != nil {
		return responses.InternalServerError(c, "Gagal memperbarui tanggal kedaluwarsa produk: "+err.Error(), err)
	}

	var opnameItem models.OpnameItems
	opnameItem.OpnameId = input.OpnameId
	opnameItem.ProductId = input.ProductId
	opnameItem.Qty = input.Qty
	opnameItem.ExpiredDate = parsedDate
	opnameItem.Price = product.PurchasePrice
	opnameItem.QtyExist = product.Stock
	opnameItem.SubTotalExist = product.Stock * product.PurchasePrice
	opnameItem.SubTotal = opnameItem.Qty * product.PurchasePrice

	var existingItem models.OpnameItems
	err = db.Where("opname_id = ? AND product_id = ?", opnameItem.OpnameId, opnameItem.ProductId).First(&existingItem).Error

	if err == nil {
		existingItem.Qty = opnameItem.Qty
		existingItem.SubTotal = opnameItem.SubTotal
		existingItem.ExpiredDate = opnameItem.ExpiredDate

		if err := db.Save(&existingItem).Error; err != nil {
			// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memperbarui item opname: " + err.Error()})
			return responses.InternalServerError(c, "Gagal memperbarui item opname: "+err.Error(), err)
		}

		if err := tools.OpnameProductStock(db, opnameItem.ProductId, opnameItem.Qty); err != nil {
			return responses.InternalServerError(c, "Gagal menyesuaikan stok produk saat pembaruan: "+err.Error(), err)
		}

		if err := tools.RecalculateTotalOpname(db, opnameItem.OpnameId); err != nil {
			return responses.InternalServerError(c, "Gagal menghitung ulang total opname: "+err.Error(), err)
		}

		return responses.JSONResponse(c, http.StatusOK, "Item opname berhasil diperbarui", existingItem)

	} else if err != gorm.ErrRecordNotFound {
		return responses.InternalServerError(c, "Terjadi kesalahan di database saat pengecekan: "+err.Error(), err)
	}

	if opnameItem.ID == "" {
		opnameItem.ID = helpers.GenerateID("OPI")
	}

	if err := db.Create(&opnameItem).Error; err != nil {
		return responses.InternalServerError(c, "Gagal menambahkan item opname: "+err.Error(), err)
	}

	if err := tools.OpnameProductStock(db, opnameItem.ProductId, opnameItem.Qty); err != nil {
		return responses.InternalServerError(c, "Gagal menyesuaikan stok produk saat pembuatan: "+err.Error(), err)
	}

	if err := tools.RecalculateTotalOpname(db, opnameItem.OpnameId); err != nil {
		return responses.InternalServerError(c, "Gagal menghitung ulang total opname: "+err.Error(), err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item opname berhasil disimpan", opnameItem)
}

// GetAllOpnameItems tampilkan semua item berdasarkan product_name tanpa pagination
func GetAllOpnameItems(c *framework.Ctx) error {

	// Get Opname id dari param
	opnameID := c.Param("id")

	// Parsing body JSON ke struct
	var OpnameItems []models.AllOpnameItemMobiles

	// Query dasar
	query := config.DB.Table("opname_items pit").
		Select("pit.id, pit.opname_id, pit.product_id, pro.name AS product_name, TO_CHAR(pit.price, 'FM999G999G999') AS price, pit.qty, pit.qty_exist, TO_CHAR(pit.sub_total, 'FM999G999G999') AS sub_total, TO_CHAR(pit.sub_total_exist, 'FM999G999G999') AS sub_total_exist, TO_CHAR(pit.expired_date, 'DD-MM-YYYY') AS expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Where("pit.opname_id = ?", opnameID).
		Order("pro.name ASC")

	// Eksekusi query
	if err := query.Scan(&OpnameItems).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Pengambilan item gagal", "Gagal mengambil data item Opname")
	}

	return responses.JSONResponse(c, http.StatusOK, "Item berhasil ditampilkan", OpnameItems)
}

// Update OpnameItem
func UpdateOpnameItemByID(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var existingItem models.OpnameItems
	if err := db.First(&existingItem, "id = ?", id).Error; err != nil {
		return responses.JSONResponse(c, http.StatusNotFound, "Item tidak ditemukan", nil)
	}

	var updatedItem tools.CreateOpnameItemUpdate
	if err := c.BodyParser(&updatedItem); err != nil {
		return responses.JSONResponse(c, http.StatusBadRequest, "Masukan tidak valid", nil)
	}

	// Kosongkan stok lama
	if err := tools.ZeroProductStock(db, existingItem.ProductId, existingItem.Qty); err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengosongkan stok lama: "+err.Error(), err)
	}

	// Tambah stok baru
	if err := tools.AddProductStock(db, updatedItem.ProductId, updatedItem.Qty); err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal menambah stok baru: "+err.Error(), err)
	}

	// Update item
	existingItem.ProductId = updatedItem.ProductId
	existingItem.Qty = updatedItem.Qty
	existingItem.Price = updatedItem.Price
	existingItem.SubTotal = updatedItem.Price * updatedItem.Qty

	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, updatedItem.ExpiredDate)
	if err != nil {
		return responses.JSONResponse(c, http.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
	}
	existingItem.ExpiredDate = parsedDate

	if err := db.Save(&existingItem).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal menyimpan item: "+err.Error(), err)
	}

	// Update harga produk jika harga item lebih tinggi
	if err := tools.UpdateProductPriceIfHigher(db, updatedItem.ProductId, updatedItem.Price); err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui harga produk: "+err.Error(), err)
	}

	// Recalculate total & sync
	if err := tools.RecalculateTotalOpname(db, existingItem.OpnameId); err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui total opname: "+err.Error(), err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item berhasil diperbarui", existingItem)
}

// DeleteOpnameItemByID OpnameItem
func DeleteOpnameItemByID(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var item models.OpnameItems
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return responses.JSONResponse(c, http.StatusNotFound, "Item tidak ditemukan", err)
	}

	// Subtract stok
	if err := tools.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengurangi stok produk: "+err.Error(), err)
	}

	// Hapus item
	if err := db.Delete(&item).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal menghapus item: "+err.Error(), err)
	}

	// Recalculate total
	if err := tools.RecalculateTotalOpname(db, item.OpnameId); err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui total opname: "+err.Error(), err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item berhasil dihapus", item)
}

// GetBuyProductsCombobox dengan pencarian berdasarkan body
func GetProductsComboboxByName(c *framework.Ctx) error {
	// Ambil branch ID dari token
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Bersihkan dan ubah ke lowercase
	search := strings.TrimSpace(c.Query("search"))

	// Inisialisasi response
	var prodCombo []models.ComboboxProducts

	// Query dasar
	query := config.DB.Table("products pro").
		Select("pro.id AS pro_id, pro.name AS pro_name, pro.unit_id, pro.stock, unt.name AS unit_name, pro.purchase_price AS price").
		Joins("LEFT JOIN units unt ON unt.id = pro.unit_id").
		Where("pro.branch_id = ?", branch_id).
		Order("pro.name ASC")

	// Tambahkan filter jika ada search
	if search != "" {
		query = query.Where("LOWER(pro.name) LIKE ? OR LOWER(pro.description) LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Eksekusi query
	if err := query.Find(&prodCombo).Error; err != nil {
		return responses.JSONResponse(c, http.StatusNotFound, "Combobox tidak ditemukan", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Data Combobox ditemukan", prodCombo)
}
