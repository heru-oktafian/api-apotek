package models

import (
	"time"
)

// BuyReturns model
type BuyReturns struct {
	ID          string        `gorm:"type:varchar(15);primaryKey" json:"id"`
	PurchaseId  string        `gorm:"type:varchar(15);not null" json:"buy_id" validate:"required"`
	ReturnDate  time.Time     `gorm:"not null" json:"return_date" validate:"required"`
	BranchID    string        `gorm:"type:varchar(15);not null" json:"branch_id"`
	TotalReturn int           `gorm:"type:int;not null;default:0" json:"total_purchase"`
	Payment     PaymentStatus `gorm:"type:payment_status;not null;default:'paid_by_cash'" json:"payment"`
	UserID      string        `gorm:"type:varchar(15);not null" json:"user_id"`
	CreatedAt   time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// Buy Return Items model
type BuyReturnItems struct {
	ID          string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	BuyReturnId string    `gorm:"type:varchar(15);not null" json:"buy_return_id" validate:"required"`
	ProductId   string    `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Price       int       `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty         int       `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal    int       `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	ExpiredDate time.Time `gorm:"not null;default:(NOW() + interval '2 year')" json:"expired_date" validate:"required"`
}

type BuyReturnRequest struct {
	BuyReturn      BuyReturnInput       `json:"buy_return"`
	BuyReturnItems []BuyReturnItemInput `json:"buy_return_items"`
}

// Struct baru untuk menangani input buy_return, khususnya buy_return_date sebagai string
type BuyReturnInput struct {
	ID             string        `json:"id"` // ID tidak perlu diisi dari request, akan di-generate
	PurchaseId     string        `json:"purchase_id" validate:"required"`
	ReturnDate     string        `json:"return_date"` // Diubah menjadi string
	BranchID       string        `json:"branch_id"`
	TotalBuyReturn int           `json:"total_buy_return"` // Akan dikalkulasi
	Payment        PaymentStatus `json:"payment" validate:"required"`
	UserID         string        `json:"user_id"`
	// CreatedAt dan UpdatedAt tidak perlu di input dari request
}

// Struct baru untuk menangani input buy_return items dari request
type BuyReturnItemInput struct {
	ID          string `json:"id"`
	ProductId   string `json:"product_id" validate:"required"`
	Qty         int    `json:"qty" validate:"required"`
	Price       int    `json:"price"`
	ExpiredDate string `json:"expired_date" validate:"required"` // <--- Diubah menjadi string
}

// AllBuyReturns model
type AllBuyReturns struct {
	ID          string        `gorm:"type:varchar(15);primaryKey" json:"id"`
	PurchaseId  string        `gorm:"type:varchar(15);not null" json:"buy_id" validate:"required"`
	ReturnDate  time.Time     `json:"return_date"`
	TotalReturn int           `gorm:"type:int;not null;default:0" json:"total_purchase"`
	Payment     PaymentStatus `gorm:"type:payment_status;not null;default:'paid_by_cash'" json:"payment"`
}

// All Buy Return Items model
type AllBuyReturnItems struct {
	ID          string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	BuyReturnId string    `gorm:"type:varchar(15);not null" json:"buy_return_id" validate:"required"`
	ProId       string    `gorm:"type:varchar(15);not null" json:"pro_id" validate:"required"`
	ProName     string    `gorm:"type:varchar(255);not null" json:"pro_name" validate:"required"`
	UnitId      string    `gorm:"type:varchar(15);primaryKey" json:"unit_id"`
	UnitName    string    `gorm:"type:varchar(255);not null" json:"unit_name" validate:"required"`
	Price       int       `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty         int       `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal    int       `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	ExpiredDate time.Time `json:"expired_date"`
}

// digunakan untuk item individu dalam list GetBuyReturnWithItems.
type BuyReturnItemResponse struct {
	ID          string      `json:"id"`
	PurchaseId  string      `json:"buy_id"`
	ReturnDate  string      `json:"return_date"`
	TotalReturn int         `json:"total_purchase"`
	Payment     string      `json:"payment"`
	Items       interface{} `json:"items"` // Items bisa berupa []models.AllBuyItems
}

// AllBuyReturns model
type BuyReturnsResponse struct {
	ID          string `json:"id"`
	PurchaseId  string `json:"buy_id"`
	ReturnDate  string `json:"return_date"`
	TotalReturn int    `json:"total_purchase"`
	Payment     string `json:"payment"`
}
