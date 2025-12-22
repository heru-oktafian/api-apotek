package models

// DuplicateReceipt Items model
type DuplicateReceiptItems struct {
	ID                 string `gorm:"type:varchar(15);primaryKey" json:"id"`
	DuplicateReceiptId string `gorm:"type:varchar(15);not null" json:"duplicate_receipt_id"`
	ProductId          string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Price              int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty                int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required,min=1"`
	SubTotal           int    `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
}

// All DuplicateReceipt Items model
type AllDuplicateReceiptItems struct {
	ID                 string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	DuplicateReceiptId string `gorm:"type:varchar(15);not null" json:"duplicate_receipt_id" validate:"required"`
	ProductId          string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	ProductName        string `gorm:"type:varchar(255);not null" json:"product_name" validate:"required"`
	UnitName           string `gorm:"type:varchar(255);not null" json:"unit_name" validate:"required"`
	Price              int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty                int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal           int    `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
}

// DuplicateDetailResponse adalah struct khusus untuk data detail penjualan duplikat resep,
// digunakan untuk item individu dalam list GetAllDuplicateReceipes.
type DuplicateDetailResponse struct {
	ID                    string `json:"id"`
	MemberId              string `json:"member_id"`
	MemberName            string `json:"member_name"`
	DuplicateReceiptDate  string `json:"duplicate_receipt_date"` // Ini akan menjadi STRING yang diformat
	TotalDuplicateReceipt int    `json:"total_duplicate_receipt"`
	ProfitEstimate        int    `json:"profit_estimate"`
	Payment               string `json:"payment"`
}

// DuplicateItemResponse adalah struct khusus untuk data detail penjualan duplikat resep,
// digunakan untuk item individu dalam list GetDuplicateReceiptWithItems.
type DuplicateItemResponse struct {
	ID                    string      `json:"id"`
	MemberId              string      `json:"member_id"`
	MemberName            string      `json:"member_name"`
	DuplicateReceiptDate  string      `json:"duplicate_receipt_date"` // Ini akan menjadi STRING yang diformat
	TotalDuplicateReceipt int         `json:"total_duplicate_receipt"`
	Discount              int         `json:"discount"`
	ProfitEstimate        int         `json:"profit_estimate"`
	Payment               string      `json:"payment"`
	Items                 interface{} `json:"items"` // Items bisa berupa []models.AllDuplicateReceiptItems
}
