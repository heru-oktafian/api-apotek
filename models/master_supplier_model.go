package models

// Supplier model yang akan disimpan di database
type Supplier struct {
	ID                 string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Name               string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	Phone              string `gorm:"type:varchar(100);" json:"phone"`
	Address            string `gorm:"type:text;" json:"address"`
	PIC                string `gorm:"type:varchar(255);" json:"pic"`
	SupplierCategoryId uint   `gorm:"not null" json:"supplier_category_id" validate:"required"`
	BranchID           string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// Supplier Detail model yang akan ditampilkan di data detail
type SupplierDetail struct {
	ID                 string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Name               string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	Phone              string `gorm:"type:varchar(100);" json:"phone"`
	Address            string `gorm:"type:text;" json:"address"`
	PIC                string `gorm:"type:varchar(255);" json:"pic"`
	SupplierCategoryId uint   `gorm:"not null" json:"supplier_category_id" validate:"required"`
	SupplierCategory   string `gorm:"type:varchar(100);not null" json:"supplier_category" validate:"required"`
}

// Supplier All model yang akan ditampilkan di data detail
type SupplierAll struct {
	ID               string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Name             string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	Phone            string `gorm:"type:varchar(100);" json:"phone"`
	Address          string `gorm:"type:text;" json:"address"`
	PIC              string `gorm:"type:varchar(255);" json:"pic"`
	SupplierCategory string `gorm:"type:varchar(100);not null" json:"supplier_category" validate:"required"`
}

// CmbSupplierModel adalah model untuk combo box kategori supplier
type CmbSupplierModel struct {
	SupplierId   string `json:"supplier_id"`
	SupplierName string `json:"supplier_name"`
}
