package controllers

import (
	"net/http"
	"time"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
)

// MonthlyProfitReport get monthly profit report grouped per day
func MonthlyProfitReport(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	branchID, _ := middlewares.GetBranchID(c.Request)

	now := nowWIB
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	var summariesDB []models.DailySummaryDB
	var summariesResponse []models.DailySummaryResponse

	err := db.Table("daily_profit_reports").
		Select("report_date, SUM(total_sales) AS total_sales, SUM(profit_estimate) AS profit_estimate").
		Where("report_date BETWEEN ? AND ? AND branch_id = ?", startOfMonth, endOfMonth, branchID).
		Group("report_date").
		Order("report_date").
		Scan(&summariesDB).Error // Scan ke DailySummaryDB

	if err != nil {
		return responses.InternalServerError(c, "Failed to retrieve monthly profit report", err)
	}

	// Format ulang data sebelum dikirim sebagai respons
	for _, s := range summariesDB {
		summariesResponse = append(summariesResponse, models.DailySummaryResponse{
			ReportDate:     s.ReportDate.Format("02"), // Format hanya hari (DD)
			TotalSales:     s.TotalSales,
			ProfitEstimate: s.ProfitEstimate,
		})
	}

	// Hitung total bulan
	var monthSales, monthProfit int
	for _, s := range summariesResponse { // Gunakan summariesResponse untuk perhitungan
		monthSales += s.TotalSales
		monthProfit += s.ProfitEstimate
	}

	return JSONProfitReportMonthly(c, http.StatusOK, "Sales & Profit Report this month", monthSales, monthProfit, summariesResponse)
}

// WeeklyProfitReport gets weekly profit report grouped by date
func WeeklyProfitReport(c *framework.Ctx) error {
	// Calculate current time in WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	branchID, _ := middlewares.GetBranchID(c.Request)

	// Determine the start and end of the week (Monday - Sunday)
	now := nowWIB
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is considered the 7th day
	}
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	var summariesDB []models.DailySummaryDB

	err := db.Table("daily_profit_reports").
		Select("report_date, SUM(total_sales) as total_sales, SUM(profit_estimate) as profit_estimate").
		Where("report_date BETWEEN ? AND ? AND branch_id = ?", startOfWeek, endOfWeek, branchID).
		Group("report_date").
		Order("report_date").
		Scan(&summariesDB).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to retrieve weekly profit report", err)
	}

	// Calculate total omset and profit
	var totalOmset, totalProfit int
	for _, s := range summariesDB {
		totalOmset += s.TotalSales
		totalProfit += s.ProfitEstimate
	}

	totalHPP := totalOmset - totalProfit
	hppPercentage := 0
	profitPercentage := 0

	if totalOmset > 0 { // Prevent division by zero
		hppPercentage = (totalHPP * 100) / totalOmset
		profitPercentage = (totalProfit * 100) / totalOmset
	}

	// Build the final response using the struct
	response := models.WeeklyProfitReportResponse{
		Omset:            totalOmset,
		Profit:           totalProfit,
		TotalHPP:         totalHPP,
		ProfitPercentage: profitPercentage,
		HPPPercentage:    hppPercentage,
	}

	return responses.JSONResponse(c, http.StatusOK, "Weekly sales & profit report", []models.WeeklyProfitReportResponse{response})
}

// DailyProfitReport get summarized daily profit report (by branch, for today)
func DailyProfitReport(c *framework.Ctx) error {

	db := config.DB
	branchID, _ := middlewares.GetBranchID(c.Request)

	today := time.Now().In(utils.Location).Format("2006-01-02")

	var summary models.DailySummaryResponse

	err := db.Table("daily_profit_reports").
		Select("report_date, SUM(total_sales) AS total_sales, SUM(profit_estimate) AS profit_estimate").
		Where("report_date = ? AND branch_id = ?", today, branchID).
		Group("report_date").
		Scan(&summary).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to retrieve daily profit report", err)
	}

	// Jika tidak ada transaksi hari ini, kembalikan default kosong
	if summary.ReportDate == "" {
		summary = models.DailySummaryResponse{
			ReportDate:     today,
			TotalSales:     0,
			ProfitEstimate: 0,
		}
	}

	return responses.JSONResponse(c, http.StatusOK, "Sales & Profit Report today", summary)
}

// GetTopSellingProducts get top selling products
func GetTopSellingProducts(c *framework.Ctx) error {
	db := config.DB
	branchID, _ := middlewares.GetBranchID(c.Request)
	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	type Result struct {
		ProductID string
		Name      string
		TotalQty  int
	}

	var results []Result
	err := db.
		Table("sale_items").
		Select("products.id as product_id, products.name, SUM(sale_items.qty) as total_qty").
		Joins("JOIN sales ON sales.id = sale_items.sale_id").
		Joins("JOIN products ON products.id = sale_items.product_id").
		Where("sales.sale_date >= ? AND sales.branch_id = ?", oneMonthAgo, branchID).
		Group("products.id, products.name").
		Order("total_qty DESC").
		Limit(10).
		Scan(&results).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to fetch top selling products", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Top Selling Products Last Month", results)
}

// GetLeastSellingProducts get least selling products
func GetLeastSellingProducts(c *framework.Ctx) error {
	db := config.DB
	branchID, _ := middlewares.GetBranchID(c.Request)
	// Hitung rentang waktu 1 bulan ke belakang
	now := time.Now()
	oneMonthAgo := now.AddDate(0, -1, 0)

	// Subquery: Ambil total qty penjualan per product dalam 1 bulan terakhir
	subQuery := db.
		Table("sale_items").
		Select("product_id, SUM(qty) as total_sold").
		Joins("JOIN sales ON sales.id = sale_items.sale_id").
		Where("sales.sale_date BETWEEN ? AND ?", oneMonthAgo, now).
		Group("product_id")

	// Query utama: ambil semua produk yang memiliki stok >= 1
	// dan gabungkan dengan subQuery untuk mengetahui jumlah penjualannya
	type Result struct {
		ProductID   string `json:"product_id"`
		ProductName string `json:"product_name"`
		Stock       int    `json:"stock"`
		TotalSold   int    `json:"total_sold"`
	}

	var results []Result

	if err := db.
		Table("products p").
		Select("p.id as product_id, p.name as product_name, p.stock, COALESCE(s.total_sold, 0) as total_sold").
		Joins("LEFT JOIN (?) as s ON p.id = s.product_id", subQuery).
		Where("p.stock >= ? AND p.branch_id = ?", 1, branchID).
		Order("total_sold ASC").
		Limit(25).
		Scan(&results).Error; err != nil {
		return responses.InternalServerError(c, "Failed to retrieve least selling products", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Least selling products (1 month)", results)
}

// GetExpiringProducts mendapatkan produk yang akan kadaluarsa dengan bidang spesifik,
// nama unit yang digabung, dan disaring berdasarkan stok.
func GetExpiringProducts(c *framework.Ctx) error {
	db := config.DB

	// Tentukan batas tanggal maksimal: 3 bulan dari hari ini
	// Waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)
	// 3 bulan kemudian dari hari ini.
	// Untuk memastikan perbandingan mencakup seluruh hari ke-3 bulan,
	// kita bisa mengatur jam, menit, detik, nanodetik ke 23:59:59.
	// Namun, untuk `expired_date <= ?`, `time.Now().AddDate(0, 3, 0)` sudah cukup.
	threeMonthsLater := nowWIB.AddDate(0, 3, 0)

	// Gunakan struct ini untuk menampung hasil kueri dari database
	// Karena kita ingin mengakses expired_date sebagai time.Time sebelum memformatnya
	type ProductQueryResult struct {
		ID          string    `gorm:"column:id"`
		SKU         string    `gorm:"column:sku"`
		Name        string    `gorm:"column:name"`
		Stock       int       `gorm:"column:stock"`
		Unit        string    `gorm:"column:unit"` // Alias untuk units.name
		ExpiredDate time.Time `gorm:"column:expired_date"`
	}

	var rawProducts []ProductQueryResult // Slice untuk menampung hasil kueri mentah

	// Lakukan satu kali kueri ke database
	err := db.Table("products").
		Select("products.id, products.sku, products.name, products.stock, units.name as unit, products.expired_date").
		Joins("LEFT JOIN units ON products.unit_id = units.id").                          // Gabungkan dengan tabel units
		Where("products.expired_date <= ? AND products.stock >= ?", threeMonthsLater, 1). // Filter berdasarkan tanggal kadaluarsa dan stok >= 1
		Order("products.expired_date ASC").
		Scan(&rawProducts).Error // Pindai langsung ke slice `rawProducts`

	if err != nil {
		return responses.InternalServerError(c, "Failed to fetch expiring products", err)
	}

	// Setelah mendapatkan data mentah, kita format sesuai respons yang diinginkan
	var productsResponse []models.ProductExpiredResponse
	for _, p := range rawProducts {
		productsResponse = append(productsResponse, models.ProductExpiredResponse{
			ID:          p.ID,
			SKU:         p.SKU,
			Name:        p.Name,
			Stock:       p.Stock,
			Unit:        p.Unit,
			ExpiredDate: p.ExpiredDate.Format("2006-01-02"), // Format ke "YYYY-MM-DD"
		})
	}

	return responses.JSONResponse(c, http.StatusOK, "All Product Near Expired (<= 3 month)", productsResponse)
}

// GetDailyProfitReportByUser get daily profit report by user
func GetDailyProfitReportByUser(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	branchID, _ := middlewares.GetBranchID(c.Request)

	today := nowWIB.Format("2006-01-02")

	type Result struct {
		UserID   string
		UserName string
		Profit   int
		Sales    int
	}

	var results []Result

	// Group by user_id to sum profit and sales per user
	err := db.Table("daily_profit_reports").
		Select("users.user_id, users.name AS user_name, SUM(daily_profit_reports.profit_estimate) AS profit, SUM(daily_profit_reports.total_sales) AS sales").
		Joins("JOIN users ON users.user_id = daily_profit_reports.user_id").
		Where("daily_profit_reports.report_date = ? AND daily_profit_reports.branch_id = ?", today, branchID).
		Group("users.user_id, users.name").
		Scan(&results).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to fetch report", err)
	}

	// Hitung total keseluruhan profit dan sales
	var totalProfit, totalSales int
	for _, r := range results {
		totalProfit += r.Profit
		totalSales += r.Sales
	}

	// Ambil jumlah transaksi
	var qtyTransactions int64
	err = db.Table("sales").
		Where("DATE(created_at) = ? AND branch_id = ?", today, branchID).
		Count(&qtyTransactions).Error
	if err != nil {
		return responses.InternalServerError(c, "Failed to count transactions", err)
	}

	// Hitung average per transaksi
	abvTransactions := 0
	if qtyTransactions > 0 {
		abvTransactions = int(totalSales) / int(qtyTransactions)
	}

	// Siapkan data akhir
	var reportData []framework.Map
	for _, r := range results {
		percentage := 0
		if totalProfit > 0 {
			percentage = int(float64(r.Profit) / float64(totalProfit) * 100)
		}

		reportData = append(reportData, framework.Map{
			"user_id":           r.UserID,
			"user_name":         r.UserName,
			"profit":            r.Profit,
			"sales":             r.Sales,
			"profit_percentage": percentage,
		})
	}

	return JSONProfitReportToDay(c, http.StatusOK, "Profit Report Successfully", "daily", totalProfit, totalSales, qtyTransactions, abvTransactions, reportData)
}

// JSONProfitReportToDay sends a standard JSON response format / structure
func JSONProfitReportToDay(c *framework.Ctx, status int, message string, report_type string, total_profit int, total_sales int, qty_transactions int64, abv_transactions int, data interface{}) error {
	resp := models.ResponseProfitReportToDay{
		Status:          http.StatusText(status),
		Message:         message,
		ReportType:      report_type,
		TotalProfit:     total_profit,
		TotalSales:      total_sales,
		QtyTransactions: qty_transactions,
		AbvTransactions: abv_transactions,
		Data:            data,
	}
	return responses.JSONResponse(c, status, message, resp)

}

// JSONProfitReportMonthly sends a standard JSON response format / structure
func JSONProfitReportMonthly(c *framework.Ctx, status int, message string, month_sales int, month_profit int, data interface{}) error {
	resp := models.ResponseProfitReportMonthly{
		Status:      http.StatusText(status),
		Message:     message,
		MonthSales:  month_sales,
		MonthProfit: month_profit,
		Data:        data,
	}
	return responses.JSONResponse(c, status, message, resp)

}
