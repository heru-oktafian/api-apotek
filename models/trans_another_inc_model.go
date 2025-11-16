package models

import "time"

// Another Incomes model
type AnotherIncomes struct {
	ID          string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description string        `gorm:"type:text;" json:"description"`
	IncomeDate  time.Time     `gorm:"not null" json:"income_date" validate:"required"`
	BranchID    string        `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	TotalIncome int           `gorm:"type:int;not null;default:0" json:"total_income" validate:"required"`
	Payment     PaymentStatus `gorm:"type:payment_status;not null;default:'unpaid'" json:"payment" validate:"required"`
	UserID      string        `gorm:"type:varchar(15);not null" json:"user_id" validate:"required"`
	CreatedAt   time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// All Another Incomes model
type AllAnotherIncomes struct {
	ID          string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description string        `gorm:"type:text;" json:"description"`
	IncomeDate  time.Time     `gorm:"not null" json:"income_date" validate:"required"`
	TotalIncome int           `gorm:"type:int;not null;default:0" json:"total_income" validate:"required"`
	Payment     PaymentStatus `gorm:"type:payment_status;not null;default:'unpaid'" json:"payment" validate:"required"`
}

// Another Income Input struct
// Digunakan untuk input data another income
type AnotherIncomeInput struct {
	IncomeDate  string `json:"income_date" validate:"required"`
	Description string `gorm:"type:text;" json:"description"`
	TotalIncome int    `gorm:"type:int;not null;default:0" json:"total_income" validate:"required"`
	Payment     string `json:"payment"`
}

// AnotherIncomeDetailResponse adalah struct khusus untuk data detail income lain,
// digunakan untuk item individu dalam list GetAllAnotherIncomes.
type AnotherIncomeDetailResponse struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	IncomeDate  string `json:"income_date"` // Ini akan menjadi STRING yang diformat
	TotalIncome int    `json:"total_income"`
	Payment     string `json:"payment"`
}
