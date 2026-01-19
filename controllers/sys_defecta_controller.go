package controllers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heru-oktafian/api-apotek/models"
	"github.com/heru-oktafian/api-apotek/tools"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/helpers"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
)

// CreateDefecta handles the creation of a new defecta.
func CreateDefecta(c *framework.Ctx) error {

	// Get the current time in WIB (Western Indonesia Time)
	nowWIB := time.Now().In(utils.Location)

	// Ambil informasi dari token melalui middleware
	branchID, _ := middlewares.GetBranchID(c.Request)
	// userID, _ := middlewares.GetUserID(c.Request)
	generatedID := helpers.GenerateID("DFT")

	var input models.DefectaInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid Input", nil)
	}

	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, input.DefectaDate)
	if err != nil {
		return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", nil)
	}

	// Initialize database connection
	db := config.DB

	// Create new defecta record
	defecta := models.Defectas{
		ID:            generatedID,
		DefectaDate:   parsedDate,
		TotalEstimate: 0, // Akan dikalkulasi nanti
		DefectaStatus: input.DefectaStatus,
		BranchID:      branchID,
		CreatedAt:     nowWIB,
		UpdatedAt:     nowWIB,
	}

	// Simpan defekta ke database
	if err := db.Create(&defecta).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create defecta", nil)
	}

	return responses.JSONResponse(c, http.StatusOK, "Defecta created successfully", defecta)
}

func UpdateDefecta(c *framework.Ctx) error {

	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	id := c.Param("id")

	// Inisialisasi input dan defecta
	var input models.DefectaInput
	var defecta models.Defectas

	// Cek validasi input
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Cek apakah defecta dengan ID tersebut ada
	if err := db.First(&defecta, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Defecta not found")
	}

	// Cek apakah defecta masih bisa diedit
	editable, err := tools.IsEditable(db, "defectas", defecta.ID, 30*time.Minute)
	if err != nil {
		return responses.InternalServerError(c, "Error checking defecta status", nil)
	}
	if !editable {
		return responses.BadRequest(c, "Defecta cannot be edited in its current status", nil)
	}

	// Perbarui field defecta_date jika ada di input
	if input.DefectaDate != "" {
		layout := "2006-01-02"
		parsedDate, err := time.Parse(layout, input.DefectaDate)
		if err != nil {
			return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", nil)
		}
		defecta.DefectaDate = parsedDate
	}

	// Perbarui field defecta_status jika ada di input
	if input.DefectaStatus != "" {
		defecta.DefectaStatus = input.DefectaStatus
	}

	// Perbarui field-field defecta
	defecta.DefectaStatus = input.DefectaStatus
	defecta.UpdatedAt = nowWIB

	// Simpan perubahan ke database
	if err := db.Save(&defecta).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update defecta", nil)
	}

	// Kembalikan respons sukses
	return responses.JSONResponse(c, http.StatusOK, "Defecta updated successfully", defecta)
}

func DeleteDefecta(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Cek apakah defecta dengan ID tersebut ada
	var defecta models.Defectas
	if err := db.First(&defecta, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Defecta not found")
	}

	// Cek apakah defecta masih bisa dihapus
	editable, err := tools.IsEditable(db, "defectas", defecta.ID, 30*time.Minute)
	if err != nil {
		return responses.InternalServerError(c, "Error checking defecta status", nil)
	}
	if !editable {
		return responses.BadRequest(c, "Defecta cannot be deleted in its current status", nil)
	}

	// Hapus detail items yang terkait dengan defecta
	if err := db.Where("defecta_id = ?", defecta.ID).Delete(&models.DefectaItems{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete defecta items", nil)
	}

	// Hapus defecta dari database
	if err := db.Delete(&defecta).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete defecta", nil)
	}

	// Kembalikan respons sukses
	return responses.JSONResponse(c, http.StatusOK, "Defecta deleted successfully", nil)
}

func CreateDefectaItem(c *framework.Ctx) error {

	db := config.DB

	var input models.DefectaInputItem
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid Input", nil)
	}

	generatedID := helpers.GenerateID("DFI")

	// Cek apakah product_id sudah ada dalam defecta_items dengan defecta_id yang sama
	var existingItem models.DefectaItems
	result := db.Where("defecta_id = ? AND product_id = ?", input.DefectaId, input.ProductId).First(&existingItem)

	if result.Error == nil {
		// Item sudah ada, update qty
		existingItem.Qty += input.Qty
		existingItem.SubTotal = existingItem.Price * existingItem.Qty

		if err := db.Save(&existingItem).Error; err != nil {
			return responses.InternalServerError(c, "Failed to update defecta item", nil)
		}

		defectaItem := existingItem
		return responses.JSONResponse(c, http.StatusOK, "Defecta item updated successfully", defectaItem)
	}

	// Item belum ada, buat item baru
	defectaItem := models.DefectaItems{
		ID:        generatedID,
		DefectaId: input.DefectaId,
		ProductId: input.ProductId,
		UnitId:    input.UnitId,
		Price:     input.Price,
		Qty:       input.Qty,
		SubTotal:  input.Price * input.Qty,
	}

	// Simpan defecta item ke database
	if err := db.Create(&defectaItem).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create defecta item", nil)
	}

	return responses.JSONResponse(c, http.StatusOK, "Defecta item created successfully", defectaItem)
}

// UpdateDefectaItem menangani pembaruan item defecta yang sudah ada.
func UpdateDefectaItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var input models.DefectaInputItem
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid Input", nil)
	}

	// Cek apakah defecta item dengan ID tersebut ada
	var defectaItem models.DefectaItems
	if err := db.First(&defectaItem, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Defecta item not found")
	}

	// Perbarui field-field defecta item
	defectaItem.ProductId = input.ProductId
	if input.UnitId != "" {
		defectaItem.UnitId = input.UnitId
	}
	if input.Price != 0 {
		defectaItem.Price = input.Price
	}
	defectaItem.Qty = input.Qty
	defectaItem.SubTotal = input.Price * input.Qty

	// Simpan perubahan ke database
	if err := db.Save(&defectaItem).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update defecta item", nil)
	}

	// Kembalikan respons sukses
	return responses.JSONResponse(c, http.StatusOK, "Defecta item updated successfully", defectaItem)
}

// DeleteDefectaItem menangani penghapusan item defecta yang ada.
func DeleteDefectaItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Cek apakah defecta item dengan ID tersebut ada
	var defectaItem models.DefectaItems
	if err := db.First(&defectaItem, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Defecta item not found")
	}

	// Hapus defecta item dari database
	if err := db.Delete(&defectaItem).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete defecta item", nil)
	}

	// Kembalikan respons sukses
	return responses.JSONResponse(c, http.StatusOK, "Defecta item deleted successfully", nil)
}

func GetAllDefectas(c *framework.Ctx) error {
	// Dapatkan waktu sekarang di WIB
	nowWIB := time.Now().In(utils.Location)

	// Ambil informasi dari token melalui middleware
	branchID, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter query dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

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

	// Inisialisasi slice untuk menampung defectas dan variabel total
	var defectas []models.Defectas
	var total int64

	// Bangun query dasar
	query := config.DB.Table("defectas df").
		Select("df.id, df.defecta_date, df.total_estimate, df.defecta_status").
		Where("df.branch_id = ?", branchID)

	// Filter berdasarkan bulan
	startDate, err := time.Parse("2006-01", month)
	if err != nil {
		return responses.BadRequest(c, "Invalid month format. Use YYYY-MM", nil)
	}
	endDate := startDate.AddDate(0, 1, 0)
	query = query.Where("df.defecta_date >= ? AND df.defecta_date < ?", startDate, endDate)

	// Terapkan pencarian jika ada
	if search != "" {
		likeSearch := "%" + search + "%"
		query = query.Where("df.id LIKE ? OR df.defecta_status LIKE ?", likeSearch, likeSearch)
	}

	// Hitung total data untuk pagination
	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Failed to count defectas", nil)
	}

	// Ambil data dengan pagination
	if err := query.Order("df.created_at DESC").Limit(limit).Offset(offset).Find(&defectas).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch defectas", nil)
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Format data defectas sebelum dikirimkan dalam respons
	var formattedDefectas []models.DefectaDetailResponse
	for _, d := range defectas {
		formattedDefectas = append(formattedDefectas, models.DefectaDetailResponse{
			ID:            d.ID,
			DefectaDate:   utils.FormatIndonesianDate(d.DefectaDate),
			TotalEstimate: d.TotalEstimate,
			DefectaStatus: string(d.DefectaStatus),
		})
	}

	// Siapkan data respons dengan pagination
	return responses.JSONResponseGetAll(c, http.StatusOK, "Defectas retrieved successfully", search, int(total), page, totalPages, limit, formattedDefectas)
}

// GetAllDefectaItems menangani pengambilan semua item defecta untuk defecta tertentu.
func GetAllDefectaItems(c *framework.Ctx) error {
	// Ambil parameter defectaID dari URL
	defectaID := c.Param("id")

	// Inisialisasi slice untuk menampung defecta items
	var defectaItems []models.AllDefectaItems

	// Bangun query untuk mengambil defecta items beserta nama produk dan unitnya
	query := config.DB.Table("defecta_items di").
		Select("di.id, di.defecta_id, pro.name as product_name, un.name as unit_name, di.price, di.qty, di.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = di.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("di.defecta_id = ?", defectaID)

	// Eksekusi query
	if err := query.Find(&defectaItems).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch defecta items", nil)
	}

	// Kembalikan respons sukses dengan data defecta items
	return responses.JSONResponse(c, http.StatusOK, "Defecta items retrieved successfully", defectaItems)
}

func GetDefetaWithItems(c *framework.Ctx) error {
	defectaID := c.Param("id")
	db := config.DB

	// Ambil data defecta
	var defecta models.Defectas
	if err := db.First(&defecta, "id = ?", defectaID).Error; err != nil {
		return responses.NotFound(c, "Defecta not found")
	}

	// Ambil data item defecta beserta nama produk dan unitnya
	var defectaItems []models.AllDefectaItems
	if err := db.Table("defecta_items di").
		Select("di.id, di.defecta_id, pro.name as product_name, un.name as unit_name, di.price, di.qty, di.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = di.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("di.defecta_id = ?", defecta.ID).
		Find(&defectaItems).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch defecta items", nil)
	}

	var formatedDefectaItems []models.AllDefectaItems
	for _, item := range defectaItems {
		formatedDefectaItems = append(formatedDefectaItems, models.AllDefectaItems{
			ID:          item.ID,
			DefectaId:   item.DefectaId,
			ProductName: item.ProductName,
			UnitName:    item.UnitName,
			Price:       item.Price,
			Qty:         item.Qty,
			SubTotal:    item.SubTotal,
		})
	}

	formatedDefetaDate := utils.FormatIndonesianDate(defecta.DefectaDate)

	// Siapkan respons dengan detail defecta dan itemnya
	response := models.DefectaDetailWithItemsResponse{
		ID:            defecta.ID,
		DefectaDate:   formatedDefetaDate,
		TotalEstimate: defecta.TotalEstimate,
		DefectaStatus: string(defecta.DefectaStatus),
		Items:         formatedDefectaItems,
	}

	return responses.JSONResponse(c, http.StatusOK, "Defecta details retrieved successfully", response)
}
