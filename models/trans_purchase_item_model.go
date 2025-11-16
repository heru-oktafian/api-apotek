package models

import "time"

// Purchase Items model
type PurchaseItems struct {
	ID          string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	PurchaseId  string    `gorm:"type:varchar(15);not null" json:"purchase_id" validate:"required"`
	ProductId   string    `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	UnitId      string    `gorm:"type:varchar(15);not null;default:'UNT250118123203'" json:"unit_id" validate:"required"`
	Price       int       `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty         int       `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal    int       `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	ExpiredDate time.Time `gorm:"not null;default:(NOW() + interval '2 year')" json:"expired_date" validate:"required"`
}

// All Purchase Items model
type AllPurchaseItems struct {
	ID          string    `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	PurchaseId  string    `gorm:"type:varchar(15);not null" json:"purchase_id" validate:"required"`
	ProductId   string    `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	ProductName string    `gorm:"type:varchar(255);not null" json:"product_name" validate:"required"`
	Price       int       `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty         int       `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	UnitId      string    `gorm:"type:varchar(15);not null" json:"unit_id" validate:"required"`
	UnitName    string    `gorm:"type:varchar(255);not null" json:"unit_name" validate:"required"`
	SubTotal    int       `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	ExpiredDate time.Time `gorm:"not null;default:(NOW() + interval '2 year')" json:"expired_date" validate:"required"`
}

// PurchaseItemResponse merepresentasikan detail setiap item dalam respons
type PurchaseItemResponse struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"` // <--- Tambahan
	UnitID      string `json:"unit_id"`
	UnitName    string `json:"unit_name"` // <--- Tambahan
	Price       int    `json:"price"`
	Qty         int    `json:"qty"`
	SubTotal    int    `json:"sub_total"`
	ExpiredDate string `json:"expired_date"` // <--- Diubah ke string dengan format kustom
}

// PurchaseResponse merepresentasikan detail pembelian utama dalam respons
type PurchaseResponse struct {
	ID            string                 `json:"id"`
	SupplierID    string                 `json:"supplier_id"`
	SupplierName  string                 `json:"supplier_name"` // <--- Tambahan
	PurchaseDate  string                 `json:"purchase_date"` // <--- Diubah ke string dengan format kustom
	TotalPurchase int                    `json:"total_purchase"`
	Payment       PaymentStatus          `json:"payment"`
	Items         []PurchaseItemResponse `json:"items"` // <--- Tambahan: slice dari item
}

// GetFixedPriceRequest digunakan untuk parsing query parameters
type GetFixedPriceRequest struct {
	ProductID string `query:"product_id" validate:"required"`
	InitID    string `query:"init_id" validate:"required"`
	FinalID   string `query:"final_id" validate:"required"` // Parameter opsional, jika tidak ada, diasumsikan UnitId dari Product
}

// FixedPriceResponse merepresentasikan struktur respons yang diinginkan
type FixedPriceResponse struct {
	FixPrice int `json:"fix_price"`
}

// Request body struct untuk transaksi pembelian
type PurchaseTransactionRequest struct {
	Purchase      PurchaseInput       `json:"purchase" validate:"required"`
	PurchaseItems []PurchaseItemInput `json:"purchase_items" validate:"required,min=1,dive"` // Menggunakan PurchaseItemInput
}

// Struct baru untuk menangani input purchase items dari request
type PurchaseItemInput struct {
	ID          string `json:"id"`
	ProductId   string `json:"product_id" validate:"required"`
	UnitId      string `json:"unit_id" validate:"required"`
	Qty         int    `json:"qty" validate:"required"`
	Price       int    `json:"price" validate:"required"`
	ExpiredDate string `json:"expired_date" validate:"required"` // <--- Diubah menjadi string
}

type FormatedPurchaseItems struct {
	ID          string `json:"id"`
	ProductId   string `json:"product_id"`
	ProductName string `json:"product_name"`
	UnitId      string `json:"unit_id"`
	UnitName    string `json:"unit_name"`
	Price       int    `json:"price"`
	Qty         int    `json:"qty"`
	SubTotal    int    `json:"sub_total"`
	ExpiredDate string `json:"expired_date"`
}
