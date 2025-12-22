package models

// DuplicateReceipe Items model
type DuplicateReceipeItems struct {
	ID                 string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	DuplicateReceipeId string `gorm:"type:varchar(15);not null" json:"duplicate_receipe_id" validate:"required"`
	ProductId          string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Price              int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty                int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal           int    `gorm:"type:int;not null;default:0" json:"sub_total"`
}

// All DuplicateReceipe Items model
type AllDuplicateReceipeItems struct {
	ID                 string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	DuplicateReceipeId string `gorm:"type:varchar(15);not null" json:"duplicate_receipe_id" validate:"required"`
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
	DuplicateReceipeDate  string `json:"duplicate_receipe_date"` // Ini akan menjadi STRING yang diformat
	TotalDuplicateReceipe int    `json:"total_duplicate_receipe"`
	ProfitEstimate        int    `json:"profit_estimate"`
	Payment               string `json:"payment"`
}
