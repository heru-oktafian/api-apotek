package models

// ProductCategory model
type ProductCategory struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	BranchID string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

type ComboProductCategory struct {
	ProductCategoryID   uint   `gorm:"primaryKey;autoIncrement" json:"product_category_id"`
	ProductCategoryName string `gorm:"type:varchar(100);not null" json:"product_category_name" validate:"required"`
}
