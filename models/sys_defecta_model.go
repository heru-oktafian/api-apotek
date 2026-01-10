package models

import "time"

// Defectas model
type Defectas struct {
	ID            string     `gorm:"type:varchar(15);primaryKey" json:"id"`
	DefectaDate   time.Time  `gorm:"not null" json:"defecta_date" validate:"required"`
	TotalEstimate int        `gorm:"type:int;not null;default:0" json:"total_estimate"`
	DefectaStatus DataStatus `gorm:"type:data_status;not null;default:'inactive'" json:"defecta_status" validate:"required"`
	BranchID      string     `gorm:"type:varchar(15);not null" json:"branch_id"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// Defecta Items model
type DefectaItems struct {
	ID        string `gorm:"type:varchar(15);primaryKey" json:"id"`
	DefectaId string `gorm:"type:varchar(15);not null" json:"defecta_id" validate:"required"`
	ProductId string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	UnitId    string `gorm:"type:varchar(15);not null;default:'UNT250118123203'" json:"unit_id" validate:"required"`
	Price     int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty       int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal  int    `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
}

// Struct baru untuk menangani input defecta, khususnya defecta_date sebagai string
type DefectaInput struct {
	DefectaDate   string     `json:"defecta_date"`   // Diubah menjadi string
	TotalEstimate int        `json:"total_estimate"` // Akan dikalkulasi
	DefectaStatus DataStatus `json:"defecta_status"`
	BranchID      string     `json:"branch_id"`
}

// DefectaInputItem digunakan untuk menangani input item defecta
type DefectaInputItem struct {
	DefectaId string `json:"defecta_id" validate:"required"`
	ProductId string `json:"product_id" validate:"required"`
	UnitId    string `json:"unit_id" validate:"required"`
	Price     int    `json:"price" validate:"required"`
	Qty       int    `json:"qty" validate:"required"`
}

// DefectaDetailResponse merepresentasikan detail defecta dalam respons
type DefectaDetailResponse struct {
	ID            string `json:"id"`
	DefectaDate   string `json:"defecta_date"` // Ini akan menjadi STRING yang diformat
	TotalEstimate int    `json:"total_estimate"`
	DefectaStatus string `json:"defecta_status"`
}

// DefectaDetailResponse merepresentasikan detail defecta disertai items dalam respons
type DefectaDetailWithItemsResponse struct {
	ID            string      `json:"id"`
	DefectaDate   string      `json:"defecta_date"` // Ini akan menjadi STRING yang diformat
	TotalEstimate int         `json:"total_estimate"`
	DefectaStatus string      `json:"defecta_status"`
	Items         interface{} `json:"items"`
}
