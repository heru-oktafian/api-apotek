package models

import "time"

// Expenses model
type Expenses struct {
	ID           string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description  string        `gorm:"type:text;" json:"description"`
	ExpenseDate  time.Time     `gorm:"not null" json:"expense_date" validate:"required"`
	BranchID     string        `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	TotalExpense int           `gorm:"type:int;not null;default:0" json:"total_expense" validate:"required"`
	Payment      PaymentStatus `gorm:"type:payment_status;not null;default:'unpaid'" json:"payment" validate:"required"`
	UserID       string        `gorm:"type:varchar(15);not null" json:"user_id" validate:"required"`
	CreatedAt    time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

type ExpenseInput struct {
	ExpenseDate  string `json:"expense_date" validate:"required"`
	Description  string `gorm:"type:text;" json:"description"`
	TotalExpense int    `gorm:"type:int;not null;default:0" json:"total_expense" validate:"required"`
	Payment      string `json:"payment"`
}

// ExpenseDetailResponse adalah struct khusus untuk data detail expenses,
// digunakan untuk item individu dalam list GetAllAnotherIncomes.
type ExpenseDetailResponse struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	ExpenseDate  string `json:"expense_date"` // Ini akan menjadi STRING yang diformat
	TotalExpense int    `json:"total_expense"`
	Payment      string `json:"payment"`
}
