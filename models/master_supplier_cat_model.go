package models

// SupplierCategory model
type SupplierCategory struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	BranchID string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// SupplierCategoryCombo model yang akan ditampilkan di data combobox
type SupplierCategoryCombo struct {
	SupplierCategoryID   uint   `gorm:"primaryKey;autoIncrement" json:"supplier_category_id"`
	SupplierCategoryName string `gorm:"type:varchar(100);not null" json:"supplier_category_name" validate:"required"`
}
