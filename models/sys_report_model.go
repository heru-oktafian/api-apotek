package models

import "time"

// Custom type for ENUM TransactionType
type TransactionType string

const (
	Purchase   TransactionType = "purchase"
	Sale       TransactionType = "sale"
	Expense    TransactionType = "expense"
	Income     TransactionType = "income"
	FirstStock TransactionType = "first_stock"
	Ipname     TransactionType = "opname"
	SaleReturn TransactionType = "sale_return"
	BuyReturn  TransactionType = "buy_return"
)

// TransactionReport model
type TransactionReports struct {
	ID              string          `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	TransactionType TransactionType `gorm:"type:transaction_type;not null;default:'expense'" json:"transaction_type" validate:"required"`
	UserID          string          `gorm:"type:varchar(15);primaryKey" json:"user_id" validate:"required"`
	BranchID        string          `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	Total           int             `gorm:"type:int;not null;default:0" json:"total" validate:"required"`
	Payment         PaymentStatus   `gorm:"type:payment_status;not null;default:'unpaid'" json:"payment" validate:"required"`
	CreatedAt       time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// DailyProfitReport model
type DailyProfitReport struct {
	ID             string    `gorm:"primaryKey"`
	ReportDate     time.Time `gorm:"type:date;index:idx_report_date_branch"`
	UserID         string    `gorm:"type:varchar(15);primaryKey" json:"user_id" validate:"required"`
	BranchID       string    `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	TotalSales     int       `gorm:"type:int;not null;default:0" json:"total_sales" validate:"required"`
	ProfitEstimate int       `gorm:"type:int;not null;default:0" json:"profit_estimate" validate:"required"`
	CreatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type BalanceReport struct {
	ID              string          `json:"id"`
	TransactionType TransactionType `json:"transaction_type"`
	UserID          string          `json:"user_id"`
	BranchID        string          `json:"branch_id"`
	Total           int             `json:"total"`
	Payment         PaymentStatus   `json:"payment"`
	CreatedAt       time.Time       `json:"created_at"`
}

type NeracaResponse struct {
	Debit       []BalanceReport `json:"debit"`
	Credit      []BalanceReport `json:"credit"`
	TotalDebit  int             `json:"total_debit"`
	TotalCredit int             `json:"total_credit"`
}

// Custom type for ENUM MovementType
type MovementType string

const (
	PurchaseTrans       MovementType = "purchase"
	PurchaseReturnTrans MovementType = "purchase_return"
	SaleTrans           MovementType = "sale"
	SaleReturnTrans     MovementType = "sale_return"
	OpnameTrans         MovementType = "opname"
	FirstStockTrans     MovementType = "first_stock"
)

// stock_tracks model
type StockTracks struct {
	ID           string       `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	MovementType MovementType `gorm:"type:movement_type;not null;default:'purchase'" json:"movement_type" validate:"required"`
	ProductID    string       `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Stock        int          `gorm:"type:int;not null;default:0" json:"stock" validate:"required"`
	UserID       string       `gorm:"type:varchar(15);primaryKey" json:"user_id" validate:"required"`
	BranchID     string       `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	CreatedAt    time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}
