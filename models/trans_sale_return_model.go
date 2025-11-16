package models

import (
	"time"
)

// SaleReturns model
type SaleReturns struct {
	ID          string        `gorm:"type:varchar(15);primaryKey" json:"id"`
	SaleId      string        `gorm:"type:varchar(15);not null" json:"sale_id" validate:"required"`
	ReturnDate  time.Time     `gorm:"not null" json:"return_date" validate:"required"`
	BranchID    string        `gorm:"type:varchar(15);not null" json:"branch_id"`
	TotalReturn int           `gorm:"type:int;not null;default:0" json:"total_purchase"`
	Payment     PaymentStatus `gorm:"type:payment_status;not null;default:'paid_by_cash'" json:"payment"`
	UserID      string        `gorm:"type:varchar(15);not null" json:"user_id"`
	CreatedAt   time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// Sale Return Items model
type SaleReturnItems struct {
	ID           string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	SaleReturnId string    `gorm:"type:varchar(15);not null" json:"sale_return_id" validate:"required"`
	ProductId    string    `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Price        int       `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty          int       `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal     int       `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	ExpiredDate  time.Time `gorm:"not null;default:(NOW() + interval '2 year')" json:"expired_date" validate:"required"`
}

type SaleReturnRequest struct {
	SaleReturn      SaleReturnInput       `json:"sale_return"`
	SaleReturnItems []SaleReturnItemInput `json:"sale_return_items"`
}

// Struct baru untuk menangani input sale_return, khususnya sale_return_date sebagai string
type SaleReturnInput struct {
	ID              string        `json:"id"` // ID tidak perlu diisi dari request, akan di-generate
	SaleId          string        `json:"sale_id" validate:"required"`
	ReturnDate      string        `json:"return_date"` // Diubah menjadi string
	BranchID        string        `json:"branch_id"`
	TotalSaleReturn int           `json:"total_sale_return"` // Akan dikalkulasi
	Payment         PaymentStatus `json:"payment" validate:"required"`
	UserID          string        `json:"user_id"`
	// CreatedAt dan UpdatedAt tidak perlu di input dari request
}

// Struct baru untuk menangani input sale_return items dari request
type SaleReturnItemInput struct {
	ID          string `json:"id"`
	ProductId   string `json:"product_id" validate:"required"`
	Qty         int    `json:"qty" validate:"required"`
	Price       int    `json:"price"`
	ExpiredDate string `json:"expired_date" validate:"required"` // <--- Diubah menjadi string
}

// AllSaleReturns model
type AllSaleReturns struct {
	ID          string        `gorm:"type:varchar(15);primaryKey" json:"id"`
	SaleId      string        `gorm:"type:varchar(15);not null" json:"sale_id" validate:"required"`
	ReturnDate  time.Time     `json:"return_date"`
	TotalReturn int           `gorm:"type:int;not null;default:0" json:"total_purchase"`
	Payment     PaymentStatus `gorm:"type:payment_status;not null;default:'paid_by_cash'" json:"payment"`
}

// All Sale Return Items model
type AllSaleReturnItems struct {
	ID           string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	SaleReturnId string    `gorm:"type:varchar(15);not null" json:"sale_return_id" validate:"required"`
	ProId        string    `gorm:"type:varchar(15);not null" json:"pro_id" validate:"required"`
	ProName      string    `gorm:"type:varchar(255);not null" json:"pro_name" validate:"required"`
	UnitId       string    `gorm:"type:varchar(15);primaryKey" json:"unit_id"`
	UnitName     string    `gorm:"type:varchar(255);not null" json:"unit_name" validate:"required"`
	Price        int       `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty          int       `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal     int       `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	ExpiredDate  time.Time `json:"expired_date"`
}

// digunakan untuk item individu dalam list GetSaleReturnWithItems.
type SaleReturnItemResponse struct {
	ID          string      `json:"id"`
	SaleId      string      `json:"sale_id"`
	ReturnDate  string      `json:"return_date"`
	TotalReturn int         `json:"total_purchase"`
	Payment     string      `json:"payment"`
	Items       interface{} `json:"items"` // Items bisa berupa []models.AllSaleItems
}

// AllSaleReturns model
type SaleReturnsResponse struct {
	ID          string `json:"id"`
	SaleId      string `json:"sale_id"`
	ReturnDate  string `json:"return_date"`
	TotalReturn int    `json:"total_purchase"`
	Payment     string `json:"payment"`
}
