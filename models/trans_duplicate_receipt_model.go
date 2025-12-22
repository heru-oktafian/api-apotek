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

// AllDuplicateReceipes represents a record for duplicate receipts without branch and user information.
type AllDuplicateReceipes struct {
	ID                    string        `json:"id" gorm:"primaryKey;type:varchar(15)" validate:"required"`
	MemberId              string        `gorm:"type:varchar(15);not null" json:"member_id"`
	MemberName            string        `gorm:"type:varchar(100);not null" json:"member_name" validate:"required"`
	Description           string        `json:"description" gorm:"type:text"`
	DuplicateReceipeDate  time.Time     `json:"duplicate_receipe_date" gorm:"not null"`
	TotalDuplicateReceipe int           `json:"total_duplicate_receipe" gorm:"not null" validate:"required" type:"int"`
	ProfitEstimate        int           `json:"profit_estimate" gorm:"not null" validate:"required" type:"int"`
	Payment               PaymentStatus `json:"payment" gorm:"type:payment_status; default: 'unpaid';not null" validate:"required"`
}

// DuplicateReceipeInput model for input data
type DuplicateReceipeInput struct {
	DuplicateReceipeDate string  `json:"duplicate_receipe_date" validate:"required"`
	MemberId             *string `json:"member_id"`
	Discount             *int    `json:"discount"`
	Payment              string  `json:"payment"`
}
