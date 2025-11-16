package models

import (
	"time"
)

// FirstStocks model
type FirstStocks struct {
	ID              string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description     string        `gorm:"type:text;" json:"description"`
	FirstStockDate  time.Time     `gorm:"not null" json:"first_stock_date" validate:"required"`
	BranchID        string        `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	TotalFirstStock int           `gorm:"type:int;not null;default:0" json:"total_first_stock" validate:"required"`
	Payment         PaymentStatus `gorm:"type:payment_status;not null;default:'nocost'" json:"payment" validate:"required"`
	UserID          string        `gorm:"type:varchar(15);not null" json:"user_id" validate:"required"`
	CreatedAt       time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// All FirstStocks model
type AllFirstStocks struct {
	ID              string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description     string        `gorm:"type:text;" json:"description"`
	FirstStockDate  time.Time     `gorm:"not null" json:"first_stock_date" validate:"required"`
	TotalFirstStock int           `gorm:"type:int;not null;default:0" json:"total_first_stock" validate:"required"`
	Payment         PaymentStatus `gorm:"type:payment_status;not null;default:'nocost'" json:"payment" validate:"required"`
}

// FirstStockInput is the input struct for creating or updating a first stock
type FirstStockInput struct {
	FirstStockDate  string `json:"first_stock_date" validate:"required"`
	Description     string `gorm:"type:text;" json:"description"`
	TotalFirstStock int    `gorm:"type:int;not null;default:0" json:"total_first_stock" validate:"required"`
	Payment         string `json:"payment"`
}
