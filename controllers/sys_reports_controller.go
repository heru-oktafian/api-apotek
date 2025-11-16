package controllers

import (
	"net/http"
	"time"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
)

// Report neraca saldo
// GetNeracaSaldo adalah handler untuk mengambil ringkasan neraca saldo berdasarkan cabang dan bulan tertentu.
// Fungsi ini akan mengelompokkan transaksi berdasarkan tipe (debit/kredit) dan tanggal transaksi,
// lalu menghitung total debit, total kredit, serta saldo akhir.
//
// Parameter:
//   - c: *fiber.Ctx, context dari request HTTP.
//
// Proses:
//  1. Mengambil branch_id dari token JWT. Jika tidak ada, akan mengembalikan error 400.
//  2. Mengambil parameter bulan (format "YYYY-MM") dari query string, lalu mengonversinya ke rentang tanggal.
//  3. Mengambil data transaksi dari tabel "transaction_reports" berdasarkan branch_id dan rentang tanggal,
//     serta mengelompokkan berdasarkan tipe transaksi dan tanggal.
//  4. Mengelompokkan hasil transaksi menjadi debit (penjualan, pemasukan) dan kredit (pembelian, pengeluaran, retur penjualan).
//  5. Menghitung total debit, total kredit, dan saldo akhir (total debit - total kredit).
//  6. Mengembalikan hasil dalam format JSON.
//
// Response JSON:
//
//	{
//	  "debit":        [ ... ], // Daftar transaksi debit
//	  "credit":       [ ... ], // Daftar transaksi kredit
//	  "total_debit":  int,     // Total nilai debit
//	  "total_credit": int,     // Total nilai kredit
//	  "total_saldo":  int      // Saldo akhir (debit - kredit)
//	}
//
// Catatan:
//   - Jika parameter bulan tidak valid, akan mengembalikan error 400 dengan pesan yang sesuai.
//   - Jika terjadi error saat mengambil data dari database, akan mengembalikan error 500.
func GetNeracaSaldo(c *framework.Ctx) error {
	db := config.DB
	branchID, _ := middlewares.GetBranchID(c.Request)
	month := c.Query("month") // format: "2025-05"

	if branchID == "" {
		return responses.BadRequest(c, "branch_id is required", nil)
	}

	// Konversi bulan ke rentang tanggal
	var startDate, endDate time.Time
	var err error
	if month != "" {
		startDate, err = time.Parse("2006-01", month)
		if err != nil {
			return responses.BadRequest(c, "Format bulan tidak valid. Gunakan format YYYY-MM.", nil)
		}
		endDate = startDate.AddDate(0, 1, 0) // awal bulan berikutnya
	}

	type Summary struct {
		TransactionType string
		TransactionDate string
		Total           int
	}

	var summaries []Summary

	query := db.Table("transaction_reports").
		Select("transaction_type, DATE(created_at) AS transaction_date, SUM(total) AS total").
		Where("branch_id = ? AND payment != 'paid_by_credit'", branchID).
		Group("transaction_type, DATE(created_at)").
		Order("transaction_date ASC")

	if month != "" {
		query = query.Where("created_at >= ? AND created_at < ?", startDate, endDate)
	}

	if err := query.Scan(&summaries).Error; err != nil {
		return responses.InternalServerError(c, "Gagal mengambil data transaksi", err)
	}

	// Kategorikan dan hitung
	var debit []framework.Map
	var credit []framework.Map
	var totalDebit, totalCredit int

	for _, s := range summaries {
		entry := framework.Map{
			"transaction_type":  s.TransactionType,
			"transaction_date":  s.TransactionDate,
			"total_transaction": s.Total,
		}

		switch s.TransactionType {
		case string(models.Sale), string(models.Income), string(models.BuyReturn):
			debit = append(debit, entry)
			totalDebit += s.Total
		case string(models.Purchase), string(models.Expense), string(models.SaleReturn):
			credit = append(credit, entry)
			totalCredit += s.Total
		}
	}

	totalSaldo := totalDebit - totalCredit

	return c.JSON(http.StatusOK, framework.Map{
		"debit":        debit,
		"credit":       credit,
		"total_debit":  totalDebit,
		"total_credit": totalCredit,
		"total_saldo":  totalSaldo,
	})
}

// GetProfitGraphByMonth get profit graph by selected month
func GetProfitGraphByMonth(c *framework.Ctx) error {

	db := config.DB
	branchID, _ := middlewares.GetBranchID(c.Request)
	month := c.Query("month") // format: YYYY-MM

	parsedMonth, err := time.Parse("2006-01", month)
	if err != nil {
		return responses.BadRequest(c, "Invalid month format. Use YYYY-MM.", nil)
	}

	startOfMonth := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	var summaries []models.DailySummaryDB

	err = db.Table("daily_profit_reports").
		Select("report_date, SUM(total_sales) AS total_sales, SUM(profit_estimate) AS profit_estimate").
		Where("report_date BETWEEN ? AND ? AND branch_id = ?", startOfMonth, endOfMonth, branchID).
		Group("report_date").
		Order("report_date").
		Scan(&summaries).Error

	if err != nil {
		return responses.InternalServerError(c, "Gagal mengambil data laporan", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Sales & Profit Report on "+month, summaries)
}
