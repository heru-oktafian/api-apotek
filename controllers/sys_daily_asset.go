package controllers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
)

func GetAllAssets(c *framework.Ctx) error {
	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Ambil ID cabang
	branchID, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")

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

	var dailyAssetFromDB []models.AllDailyAsset // Gunakan models.DailyAsset untuk mengambil data dari DB
	var total int64

	query := config.DB.Table("daily_assets ast").
		Select("ast.id, ast.asset_date, ast.asset_value, ast.branch_id, bc.branch_name").
		Joins("LEFT JOIN branches bc on bc.id = ast.branch_id").
		Where("ast.branch_id = ? ", branchID).
		Order("ast.asset_date DESC")

	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			// return helpers.JSONResponse(c, framework.StatusBadRequest, "Invalid month format", "Month should be in format YYYY-MM")
			return responses.BadRequest(c, "Invalid month format. Month should be in format YYYY-MM", err)
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where("ast.asset_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Get assets failed", err)
	}

	if err := query.Offset(offset).Limit(limit).Scan(&dailyAssetFromDB).Error; err != nil {
		return responses.InternalServerError(c, "Get assets failed", err)
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedDailyAsset []models.DetailDailyAsset
	for _, daily := range dailyAssetFromDB {
		formattedDailyAsset = append(formattedDailyAsset, models.DetailDailyAsset{
			ID:         daily.ID,
			AssetDate:  utils.FormatIndonesianDate(daily.AssetDate), // Format tanggal di sini
			AssetValue: daily.AssetValue,
			BranchId:   daily.BranchId,
			BranchName: daily.BranchName,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Gunakan JSONResponseGetAll helper dengan data yang sudah diformat
	return responses.JSONResponseGetAll(
		c,
		http.StatusOK,
		"Sales retrieved successfully",
		"",
		int(total),
		page,
		totalPages,
		limit,
		formattedDailyAsset, // Kirim data yang sudah diformat (slice dari SaleDetailResponse)
	)
}
