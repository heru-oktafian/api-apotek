package reports

import (
	"github.com/heru-oktafian/api-retail/models"
	"gorm.io/gorm"
)

// Hapus data DailyProfitReport
func DeleteDailyProfitReport(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&models.DailyProfitReport{}).Error
}

// Insert atau update laporan transaksi penjualan berdasarkan DailyProfit
func SyncDailyProfitReport(db *gorm.DB, sale models.Sales) error {
	var report models.DailyProfitReport

	err := db.Where("id = ? AND branch_id = ? AND user_id = ?", sale.ID, sale.BranchID, sale.UserID).
		First(&report).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	report.ID = sale.ID
	report.ReportDate = sale.SaleDate
	report.BranchID = sale.BranchID
	report.UserID = sale.UserID
	report.TotalSales = sale.TotalSale
	report.ProfitEstimate = sale.ProfitEstimate

	if err == gorm.ErrRecordNotFound {
		return db.Create(&report).Error
	} else {
		return db.Save(&report).Error
	}

}
