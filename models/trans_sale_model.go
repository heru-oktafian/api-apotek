package models

import (
	"time"
)

// Sales model
type Sales struct {
	ID             string        `gorm:"type:varchar(15);primaryKey" json:"id"` // Hapus validate:"required"
	MemberId       string        `gorm:"type:varchar(15);not null" json:"member_id"`
	SaleDate       time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP" json:"sale_date"`
	BranchID       string        `gorm:"type:varchar(15);not null" json:"branch_id"`
	TotalSale      int           `gorm:"type:int;not null;default:0" json:"total_sale"`      // Hapus validate:"required"
	Discount       int           `gorm:"type:int;not null;default:0" json:"discount"`        // Tetap ada jika diskon wajib diisi klien
	ProfitEstimate int           `gorm:"type:int;not null;default:0" json:"profit_estimate"` // Hapus validate:"required"
	Payment        PaymentStatus `gorm:"type:payment_status;not null;default:'unpaid'" json:"payment"`
	UserID         string        `gorm:"type:varchar(15);not null" json:"user_id"`
	CreatedAt      time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// AllSales model is a combination of Sales and Members
type AllSales struct {
	ID             string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	MemberId       string        `gorm:"type:varchar(15);not null" json:"member_id" validate:"required"`
	MemberName     string        `gorm:"type:varchar(100);not null" json:"member_name" validate:"required"`
	Discount       int           `gorm:"type:int;not null;default:0" json:"discount"`
	ProfitEstimate int           `gorm:"type:int;not null;default:0" json:"profit_estimate" validate:"required"`
	SaleDate       time.Time     `gorm:"not null" json:"sale_date" validate:"required"` // Tetap time.Time
	TotalSale      int           `gorm:"type:int;not null;default:0" json:"total_sale" validate:"required"`
	Payment        PaymentStatus `gorm:"type:payment_status;not null;default:'unpaid'" json:"payment" validate:"required"`
}

// SaleInput model for input data
type SaleInput struct {
	SaleDate string  `json:"sale_date" validate:"required"`
	MemberId *string `json:"member_id"` // ubah jadi pointer
	Discount *int    `json:"discount"`  // tetap pointer seperti sebelumnya
	Payment  string  `json:"payment"`
}
