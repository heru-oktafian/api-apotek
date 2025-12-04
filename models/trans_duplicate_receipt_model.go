package models

import "time"

// DuplicateReceipes represents a record for duplicate receipts in the system.
type DuplicateReceipes struct {
	ID                    string        `json:"id" gorm:"primaryKey;type:varchar(15)" validate:"required"`
	MemberId              string        `gorm:"type:varchar(15);not null" json:"member_id"`
	Description           string        `json:"description" gorm:"type:text"`
	DuplicateReceipeDate  time.Time     `json:"duplicate_receipe_date" gorm:"not null"`
	TotalDuplicateReceipe int           `json:"total_duplicate_receipe" gorm:"not null" validate:"required" type:"int"`
	ProfitEstimate        int           `json:"profit_estimate" gorm:"not null" validate:"required" type:"int"`
	Payment               PaymentStatus `json:"payment" gorm:"type:payment_status; default: 'unpaid';not null" validate:"required"`
	BranchID              string        `json:"branch_id" gorm:"type:varchar(15);not null" validate:"required"`
	UserID                string        `json:"user_id" gorm:"type:varchar(15);not null" validate:"required"`
	CreatedAt             time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// AllDuplicateRecipes represents a record for duplicate receipts without branch and user information.
type AllDuplicateRecipes struct {
	ID                    string        `json:"id" gorm:"primaryKey;type:varchar(15)" validate:"required"`
	MemberId              string        `gorm:"type:varchar(15);not null" json:"member_id"`
	Description           string        `json:"description" gorm:"type:text"`
	DuplicateReceipeDate  time.Time     `json:"duplicate_receipe_date" gorm:"not null"`
	TotalDuplicateReceipe int           `json:"total_duplicate_receipe" gorm:"not null" validate:"required" type:"int"`
	ProfitEstimate        int           `json:"profit_estimate" gorm:"not null" validate:"required" type:"int"`
	Payment               PaymentStatus `json:"payment" gorm:"type:payment_status; default: 'unpaid';not null" validate:"required"`
}

// DuplicateRecipe Items model
type DuplicateRecipeItems struct {
	ID                string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	DuplicateRecipeId string `gorm:"type:varchar(15);not null" json:"duplicate_recipe_id" validate:"required"`
	ProductId         string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Price             int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty               int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal          int    `gorm:"type:int;not null;default:0" json:"sub_total"`
}

// All DuplicateRecipe Items model
type AllDuplicateRecipeItems struct {
	ID                string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	DuplicateRecipeId string `gorm:"type:varchar(15);not null" json:"duplicate_recipe_id" validate:"required"`
	ProductId         string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	ProductName       string `gorm:"type:varchar(255);not null" json:"product_name" validate:"required"`
	UnitName          string `gorm:"type:varchar(255);not null" json:"unit_name" validate:"required"`
	Price             int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty               int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal          int    `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
}
