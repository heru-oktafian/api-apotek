package models

import "time"

// Product model yang akan disimpan di database
type Product struct {
	ID                string    `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	SKU               string    `gorm:"type:varchar(100);not null" json:"sku" validate:"required"`
	Name              string    `gorm:"type:varchar(255);not null" json:"name" validate:"required"`
	Description       string    `gorm:"type:text;" json:"description"`
	Ingredient        string    `gorm:"type:text;" json:"ingredient"`
	Dosage            string    `gorm:"type:text;" json:"dosage"`
	SideAffection     string    `gorm:"type:text;" json:"side_affection"`
	UnitId            string    `gorm:"type:varchar(15);not null" json:"unit_id" validate:"required"`
	Stock             int       `gorm:"type:int;not null;default:0" json:"stock"`
	PurchasePrice     int       `gorm:"type:int;not null;default:0" json:"purchase_price" validate:"required"`
	ExpiredDate       time.Time `gorm:"not null" json:"expired_date" validate:"required"`
	SalesPrice        int       `gorm:"type:int;not null;default:0" json:"sales_price" validate:"required"`
	AlternatePrice    int       `gorm:"type:int;not null;default:0" json:"alternate_price" validate:"required"`
	ProductCategoryId uint      `gorm:"not null" json:"product_category_id" validate:"required"`
	BranchID          string    `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// Product All model yang akan ditampilkan di data GetAll
type ProductAll struct {
	ID                  string    `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	SKU                 string    `gorm:"type:varchar(100);not null" json:"sku" validate:"required"`
	Name                string    `gorm:"type:varchar(255);not null" json:"name" validate:"required"`
	Description         string    `gorm:"type:text;" json:"description"`
	Ingredient          string    `gorm:"type:text;" json:"ingredient"`
	Dosage              string    `gorm:"type:text;" json:"dosage"`
	SideAffection       string    `gorm:"type:text;" json:"side_affection"`
	UnitName            string    `gorm:"type:varchar(100);not null" json:"unit_name" validate:"required"`
	Stock               int       `gorm:"type:int;not null;default:0" json:"stock"`
	PurchasePrice       int       `gorm:"type:int;not null;default:0" json:"purchase_price" validate:"required"`
	SalesPrice          int       `gorm:"type:int;not null;default:0" json:"sales_price" validate:"required"`
	AlternatePrice      int       `gorm:"type:int;not null;default:0" json:"alternate_price" validate:"required"`
	ExpiredDate         time.Time `gorm:"not null" json:"expired_date" validate:"required"`
	ProductCategoryName string    `gorm:"type:varchar(100);not null" json:"product_category_name" validate:"required"`
}

// Product Detail model yang akan ditampilkan di data detail
type ProductDetail struct {
	ID                  string    `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	SKU                 string    `gorm:"type:varchar(100);not null" json:"sku" validate:"required"`
	Name                string    `gorm:"type:varchar(255);not null" json:"name" validate:"required"`
	Description         string    `gorm:"type:text;" json:"description"`
	Ingredient          string    `gorm:"type:text;" json:"ingredient"`
	Dosage              string    `gorm:"type:text;" json:"dosage"`
	SideAffection       string    `gorm:"type:text;" json:"side_affection"`
	UnitId              string    `gorm:"type:varchar(15);not null" json:"unit_id" validate:"required"`
	UnitName            string    `gorm:"type:varchar(100);not null" json:"unit_name" validate:"required"`
	Stock               int       `gorm:"type:int;not null;default:0" json:"stock"`
	PurchasePrice       int       `gorm:"type:int;not null;default:0" json:"purchase_price" validate:"required"`
	ExpiredDate         time.Time `gorm:"not null" json:"expired_date" validate:"required"`
	SalesPrice          int       `gorm:"type:int;not null;default:0" json:"sales_price" validate:"required"`
	AlternatePrice      int       `gorm:"type:int;not null;default:0" json:"alternate_price" validate:"required"`
	ProductCategoryId   uint      `gorm:"not null" json:"product_category_id" validate:"required"`
	ProductCategoryName string    `gorm:"type:varchar(100);not null" json:"product_category_name" validate:"required"`
}

// ProdConvCombo adalah model untuk combo box konversi produk
type ProdConvCombo struct {
	ProductId   string `json:"product_id"`
	ProductName string `json:"product_name"`
}

// ProdSaleCombo adalah model untuk combo box penjualan produk
type ProdSaleCombo struct {
	ProductId   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`
	UnitName    string `json:"unit_name"`
}

// ProdPurchaseCombo adalah model untuk combo box pembelian produk
type ProdPurchaseCombo struct {
	ProductId   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int    `json:"price"`
	UnitId      string `json:"unit_id"`
	UnitName    string `json:"unit_name"`
}
