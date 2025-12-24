package reports

import (
	"time"

	"github.com/heru-oktafian/api-apotek/models"
	"gorm.io/gorm"
)

// Hapus data DailyProfitReport berdasarkan tipe transaksi
func DeleteDailyProfitReport(db *gorm.DB, id string, transType string) error {
	// Deklarasi variabel untuk menampung data transaksi
	var sale models.Sales
	var duplicate_receipe models.DuplicateReceipts
	var vsale_date time.Time
	vtotal_sale := 0
	vprofit_estimate := 0

	// Ambil data transaksi berdasarkan tipe
	switch transType {
	case "sale":
		if err := db.Where("id = ?", id).First(&sale).Error; err != nil {
			return err
		}
		vtotal_sale = sale.TotalSale
		vsale_date = sale.SaleDate
		vprofit_estimate = sale.ProfitEstimate
	case "duplicate_receipt":
		if err := db.Where("id = ?", id).First(&duplicate_receipe).Error; err != nil {
			return err
		}
		vtotal_sale = duplicate_receipe.TotalDuplicateReceipt
		vprofit_estimate = duplicate_receipe.ProfitEstimate
		vsale_date = duplicate_receipe.DuplicateReceiptDate
	}

	// Ambil data laporan profit harian berdasarkan tanggal transaksi
	var report models.DailyProfitReport
	if err := db.Where("report_date = ?", vsale_date).First(&report).Error; err != nil {
		return err
	}
	report.TotalSales -= vtotal_sale
	report.ProfitEstimate -= vprofit_estimate
	return db.Save(&report).Error
}

// Insert atau update laporan transaksi penjualan berdasarkan DailyProfit
func SyncDailyProfitReport(db *gorm.DB, branchID, userID string, reportDate time.Time, totalSales, profitEstimate int, totalBefore, profitBefore int) error {
	var report models.DailyProfitReport
	vsales := totalSales - totalBefore
	vprofit := profitEstimate - profitBefore

	err := db.Where("branch_id = ? AND user_id = ? AND report_date = ?", branchID, userID, reportDate).First(&report).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	report.UserID = userID
	report.ReportDate = reportDate
	report.TotalSales = vsales
	report.ProfitEstimate = vprofit

	if err == gorm.ErrRecordNotFound {
		return db.Create(&report).Error
	}
	return db.Save(&report).Error
}
