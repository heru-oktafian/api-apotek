package reports

import (
	"errors"
	"log"
	"time"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/utils"
	"gorm.io/gorm"
)

// Insert atau update laporan transaksi berdasarkan Purchase
func SyncPurchaseReport(db *gorm.DB, purchase models.Purchases) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Siapkan data report dari purchase
	report := models.TransactionReports{
		ID:              purchase.ID,
		TransactionType: models.Purchase,
		UserID:          purchase.UserID,
		BranchID:        purchase.BranchID,
		Total:           purchase.TotalPurchase,
		CreatedAt:       purchase.CreatedAt,
		UpdatedAt:       purchase.UpdatedAt,
		Payment:         purchase.Payment,
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

// AutoCleanupPurchases will delete any purchases older than 2 hours without purchase items
func AutoCleanupPurchases(db *gorm.DB) error {
	var purchases []models.Purchases

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Ambil semua purchase yang tidak punya purchase_items dan lebih dari 2 jam
	err := db.
		Where("created_at < ?", nowWIB.Add(-2*time.Hour)).
		Find(&purchases).Error
	if err != nil {
		return err
	}

	for _, purchase := range purchases {
		var itemCount int64
		db.Model(&models.PurchaseItems{}).
			Where("purchase_id = ?", purchase.ID).
			Count(&itemCount)

		if itemCount == 0 {
			log.Printf("🧹 Auto-cleaning orphan purchase: %s\n", purchase.ID)

			// Hapus transaction_report
			db.Where("id = ?", purchase.ID).Delete(&models.TransactionReports{})

			// Hapus purchase
			db.Where("id = ?", purchase.ID).Delete(&models.Purchases{})
		}
	}

	return nil
}
