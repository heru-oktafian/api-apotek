package reports

import (
	"errors"
	"time"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/utils"
	"gorm.io/gorm"
)

// Insert atau update laporan transaksi berdasarkan FirstStocks / Pengeluaran
func SyncOpnameReport(db *gorm.DB, opname models.Opnames) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Siapkan data report dari FirstStock
	report := models.TransactionReports{
		ID:              opname.ID,
		TransactionType: models.Ipname,
		UserID:          opname.UserID,
		BranchID:        opname.BranchID,
		Total:           opname.TotalOpname,
		CreatedAt:       opname.CreatedAt,
		UpdatedAt:       opname.UpdatedAt,
		Payment:         opname.Payment,
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
