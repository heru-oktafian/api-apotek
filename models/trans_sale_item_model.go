package models

// Sale Items model
type SaleItems struct {
	ID        string `gorm:"type:varchar(15);primaryKey" json:"id"`    // Hapus validate:"required"
	SaleId    string `gorm:"type:varchar(15);not null" json:"sale_id"` // Hapus validate:"required"
	ProductId string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Price     int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty       int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required,min=1"`
	SubTotal  int    `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
}

// All Sale Items model
type AllSaleItems struct {
	ID          string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	SaleId      string `gorm:"type:varchar(15);not null" json:"sale_id" validate:"required"`
	ProductId   string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	ProductName string `gorm:"type:varchar(255);not null" json:"product_name" validate:"required"`
	Price       int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty         int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	UnitName    string `gorm:"type:varchar(255);not null" json:"unit_name" validate:"required"`
	SubTotal    int    `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
}

// SaleItemResponse adalah struct khusus untuk data detail penjualan,
// digunakan untuk item individu dalam list GetSaleWithItems.
type SaleItemResponse struct {
	ID             string      `json:"id"`
	MemberId       string      `json:"member_id"`
	MemberName     string      `json:"member_name"`
	SaleDate       string      `json:"sale_date"` // Ini akan menjadi STRING yang diformat
	TotalSale      int         `json:"total_sale"`
	Discount       int         `json:"discount"`
	ProfitEstimate int         `json:"profit_estimate"`
	Payment        string      `json:"payment"`
	Items          interface{} `json:"items"` // Items bisa berupa []models.AllSaleItems
}

// SaleDetailResponse adalah struct khusus untuk data detail penjualan,
// digunakan untuk item individu dalam list GetAllSales.
type SaleDetailResponse struct {
	ID             string `json:"id"`
	MemberId       string `json:"member_id"`
	MemberName     string `json:"member_name"`
	SaleDate       string `json:"sale_date"` // Ini akan menjadi STRING yang diformat
	TotalSale      int    `json:"total_sale"`
	Discount       int    `json:"discount"`
	ProfitEstimate int    `json:"profit_estimate"`
	Payment        string `json:"payment"`
}
