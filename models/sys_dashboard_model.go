package models

import "time"

// Struktur untuk menampung hasil query mentah dari database
type DailySummaryDB struct {
	ReportDate     time.Time `json:"report_date"`
	TotalSales     int       `json:"total_sales"`
	ProfitEstimate int       `json:"profit_estimate"`
}

// Struktur untuk respons JSON yang dimodifikasi
type DailySummaryResponse struct {
	ReportDate     string `json:"report_date"` // Ubah tipe data menjadi string
	TotalSales     int    `json:"total_sales"`
	ProfitEstimate int    `json:"profit_estimate"`
}

// WeeklyProfitReportResponse defines the structure for the weekly profit report response
type WeeklyProfitReportResponse struct {
	Omset            int `json:"omset"`
	Profit           int `json:"profit"`
	TotalHPP         int `json:"total_hpp"`
	ProfitPercentage int `json:"profit_percentage"`
	HPPPercentage    int `json:"hpp_percentage"`
}

// ProductExpiredResponse defines the structure for each expiring product item in the response
type ProductExpiredResponse struct {
	ID          string `json:"id"`
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Stock       int    `json:"stock"`
	Unit        string `json:"unit"`         // This will hold the unit name from the 'units' table
	ExpiredDate string `json:"expired_date"` // Formatted date string
}

type ResponseProfitReportMonthly struct {
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	MonthSales  int         `json:"month_sales"`
	MonthProfit int         `json:"month_profit"`
	Data        interface{} `json:"data"`
}

type ResponseProfitReportToDay struct {
	Status          string      `json:"status"`
	Message         string      `json:"message"`
	ReportType      string      `json:"report_type"`
	TotalProfit     int         `json:"total_profit"`
	TotalSales      int         `json:"total_sales"`
	QtyTransactions int64       `json:"qty_transactions"`
	AbvTransactions int         `json:"abv_transactions"`
	Data            interface{} `json:"data"`
}
