package models

import "time"

// FirstStock Items model
type FirstStockItems struct {
	ID           string    `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	FirstStockId string    `gorm:"type:varchar(15);not null" json:"first_stock_id" validate:"required"`
	ProductId    string    `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Price        int       `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty          int       `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal     int       `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	ExpiredDate  time.Time `gorm:"not null;default:(NOW() + interval '2 year')" json:"expired_date" validate:"required"`
}

// All FirstStock Items model
type AllFirstStockItems struct {
	ID           string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	FirstStockId string `gorm:"type:varchar(15);not null" json:"first_stock_id" validate:"required"`
	ProductId    string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	ProductName  string `gorm:"type:varchar(255);not null" json:"product_name" validate:"required"`
	Price        int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty          int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	UnitName     string `gorm:"type:varchar(255);not null" json:"unit_name" validate:"required"`
	SubTotal     int    `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
}
