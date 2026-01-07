package models

import "time"

// DuplicateReceipts represents a record for duplicate receipts in the system.
type DuplicateReceipts struct {
	ID                    string        `json:"id" gorm:"primaryKey;type:varchar(15)"`
	MemberId              string        `gorm:"type:varchar(15);not null" json:"member_id"`
	Description           string        `json:"description" gorm:"type:text"`
	DuplicateReceiptDate  time.Time     `json:"duplicate_receipt_date" gorm:"not null"`
	TotalDuplicateReceipt int           `json:"total_duplicate_receipt" gorm:"not null" type:"int"`
	ProfitEstimate        int           `json:"profit_estimate" gorm:"not null" type:"int"`
	Payment               PaymentStatus `json:"payment" gorm:"type:payment_status; default: 'paid_by_cash';not null" validate:"required"`
	BranchID              string        `json:"branch_id" gorm:"type:varchar(15);not null"`
	UserID                string        `json:"user_id" gorm:"type:varchar(15);not null"`
	CreatedAt             time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// AllDuplicateReceipts represents a record for duplicate receipts without branch and user information.
type AllDuplicateReceipts struct {
	ID                    string        `json:"id" gorm:"primaryKey;type:varchar(15)" validate:"required"`
	MemberId              string        `gorm:"type:varchar(15);not null" json:"member_id"`
	MemberName            string        `gorm:"type:varchar(100);not null" json:"member_name" validate:"required"`
	Description           string        `json:"description" gorm:"type:text"`
	DuplicateReceiptDate  time.Time     `json:"duplicate_receipt_date" gorm:"not null"`
	TotalDuplicateReceipt int           `json:"total_duplicate_receipt" gorm:"not null" validate:"required" type:"int"`
	ProfitEstimate        int           `json:"profit_estimate" gorm:"not null" validate:"required" type:"int"`
	Payment               PaymentStatus `gorm:"type:payment_status;not null;default:'paid_by_cash'" json:"payment" validate:"required"`
}

// DuplicateReceiptInput model for input data
type DuplicateReceiptInput struct {
	MemberId              string        `gorm:"type:varchar(15);not null" json:"member_id"`
	Description           string        `json:"description" gorm:"type:text"`
	DuplicateReceiptDate  string        `json:"duplicate_receipt_date" validate:"required"`
	TotalDuplicateReceipt int           `json:"total_duplicate_receipt" gorm:"not null" type:"int"`
	ProfitEstimate        int           `json:"profit_estimate" gorm:"not null" type:"int"`
	Payment               PaymentStatus `json:"payment" gorm:"type:payment_status; default: 'paid_by_cash';not null" validate:"required"`
}
