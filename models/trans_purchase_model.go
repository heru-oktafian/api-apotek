package models

import "time"

// Purchases model
type Purchases struct {
	ID            string        `gorm:"type:varchar(15);primaryKey" json:"id"`
	SupplierId    string        `gorm:"type:varchar(15);not null" json:"supplier_id" validate:"required"`
	PurchaseDate  time.Time     `gorm:"not null" json:"purchase_date" validate:"required"`
	BranchID      string        `gorm:"type:varchar(15);not null" json:"branch_id"`
	TotalPurchase int           `gorm:"type:int;not null;default:0" json:"total_purchase"`
	Payment       PaymentStatus `gorm:"type:payment_status;not null;default:'unpaid'" json:"payment"`
	UserID        string        `gorm:"type:varchar(15);not null" json:"user_id"`
	CreatedAt     time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// All Purchases model
type AllPurchases struct {
	ID            string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	SupplierId    string        `gorm:"type:varchar(15);not null" json:"supplier_id" validate:"required"`
	SupplierName  string        `gorm:"type:varchar(100);not null" json:"supplier_name" validate:"required"`
	PurchaseDate  time.Time     `gorm:"not null" json:"purchase_date" validate:"required"`
	TotalPurchase int           `gorm:"type:int;not null;default:0" json:"total_purchase" validate:"required"`
	Payment       PaymentStatus `gorm:"type:payment_status;not null;default:'unpaid'" json:"payment" validate:"required"`
}

// Struct baru untuk menangani input purchase, khususnya purchase_date sebagai string
type PurchaseInput struct {
	ID            string        `json:"id"` // ID tidak perlu diisi dari request, akan di-generate
	SupplierId    string        `json:"supplier_id" validate:"required"`
	PurchaseDate  string        `json:"purchase_date"` // Diubah menjadi string
	BranchID      string        `json:"branch_id"`
	TotalPurchase int           `json:"total_purchase"` // Akan dikalkulasi
	Payment       PaymentStatus `json:"payment" validate:"required"`
	UserID        string        `json:"user_id"`
	// CreatedAt dan UpdatedAt tidak perlu di input dari request
}

type PurchaseDetailResponse struct {
	ID            string `json:"id"`
	SupplierId    string `json:"supplier_id"`
	SupplierName  string `json:"supplier_name"`
	PurchaseDate  string `json:"purchase_date"` // Ini akan menjadi STRING yang diformat
	TotalPurchase int    `json:"total_purchase"`
	Payment       string `json:"payment"`
}

type PurchaseDetailWithItemsResponse struct {
	ID            string      `json:"id"`
	SupplierId    string      `json:"supplier_id"`
	SupplierName  string      `json:"supplier_name"`
	PurchaseDate  string      `json:"purchase_date"` // Ini akan menjadi STRING yang diformat
	TotalPurchase int         `json:"total_purchase"`
	Payment       string      `json:"payment"`
	Items         interface{} `json:"items"`
}
