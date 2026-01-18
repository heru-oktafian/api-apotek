package controllers

import (
	"net/http"
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
