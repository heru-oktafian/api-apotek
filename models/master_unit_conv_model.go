package models

// UnitConversion model yang akan disimpan di database
type UnitConversion struct {
	ID        string `gorm:"type:varchar(15);primaryKey" json:"id"`
	ProductId string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	InitId    string `gorm:"type:varchar(15);not null" json:"init_id" validate:"required"`
	FinalId   string `gorm:"type:varchar(15);not null" json:"final_id" validate:"required"`
	ValueConv int    `gorm:"type:int;not null;default:0" json:"value_conv" validate:"required"`
	BranchID  string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// UnitConversionDetail model yang akan ditampilkan di data detail
type UnitConversionDetail struct {
	ID          string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	ProductName string `gorm:"type:varchar(255);not null" json:"product_name" validate:"required"`
	InitName    string `gorm:"type:varchar(100);not null" json:"init_name" validate:"required"`
	FinalName   string `gorm:"type:varchar(100);not null" json:"final_name" validate:"required"`
	ValueConv   int    `gorm:"type:int;not null;default:0" json:"value_conv" validate:"required"`
	ProductId   string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	InitId      string `gorm:"type:varchar(15);not null" json:"init_id" validate:"required"`
	FinalId     string `gorm:"type:varchar(15);not null" json:"final_id" validate:"required"`
	BranchID    string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// UnitConversionRequest merepresentasikan request body untuk membuat UnitConversion baru
type UnitConversionRequest struct {
	ProductId string `json:"product_id" validate:"required"`
	InitId    string `json:"init_id" validate:"required"`
	FinalId   string `json:"final_id" validate:"required"`
	ValueConv int    `json:"value_conv" validate:"required,min=1"` // ValueConv harus minimal 1
}

// GetUnitsByProductIdRequest merepresentasikan body request untuk endpoint GET ini
type GetUnitsByProductIdRequest struct {
	ProductID string `json:"product_id" validate:"required"`
}
