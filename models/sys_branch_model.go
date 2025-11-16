package models

import "time"

// Branch Struct untuk branch model yang akan disimpan di database
type Branch struct {
	ID               string          `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	BranchName       string          `gorm:"unique;not null" json:"branch_name"`
	Address          string          `gorm:"type:text;" json:"address"`
	Phone            string          `gorm:"type:varchar(100);" json:"phone"`
	Email            string          `gorm:"type:varchar(100);" json:"email"`
	OwnerId          string          `gorm:"type:varchar(100);" json:"owner_id"`
	OwnerName        string          `gorm:"type:varchar(255);" json:"owner_name"`
	BankName         string          `gorm:"type:varchar(255);" json:"bank_name"`
	AccountName      string          `gorm:"type:varchar(255);" json:"account_name"`
	AccountNumber    string          `gorm:"type:varchar(100);" json:"account_number"`
	TaxPercentage    int             `gorm:"type:int;default:0" json:"tax_percentage"`
	JournalMethod    JournalMethod   `gorm:"type:journal_method; default:'automatic'" json:"journal_method" validate:"required"`
	BranchStatus     DataStatus      `gorm:"type:data_status;default:'inactive'" json:"branch_status"`
	LicenseDate      time.Time       `gorm:"not null" json:"license_date" validate:"required"`
	DefaultMember    string          `gorm:"type:varchar(15);not null" json:"default_member" validate:"required"`
	SubscriptionType SubcriptionType `gorm:"type:subscription_type; default:'month'" json:"subscription_type" validate:"required"` // Kolom baru
	Quota            int             `gorm:"type:integer;default:0" json:"quota"`
}

// SetID is function to set ID into Branch
func (b *Branch) SetID(id string) {
	b.ID = id
}

// Profile Struct untuk profile model yang akan ditampilkan pada function GetDetail
type Profile struct {
	UserID        string        `gorm:"type:varchar(15);primaryKey" json:"user_id" validate:"required"`
	ProfileName   string        `gorm:"type:varchar(255);not null" json:"profile_name" validate:"required"`
	BranchID      string        `gorm:"type:varchar(15);primaryKey" json:"branch_id" validate:"required"`
	BranchName    string        `gorm:"unique;not null" json:"branch_name"`
	Address       string        `gorm:"type:text;" json:"address"`
	Phone         string        `gorm:"type:varchar(100);" json:"phone"`
	Email         string        `gorm:"type:varchar(100);" json:"email"`
	OwnerId       string        `gorm:"type:varchar(100);" json:"owner_id"`
	OwnerName     string        `gorm:"type:varchar(255);" json:"owner_name"`
	BankName      string        `gorm:"type:varchar(255);" json:"bank_name"`
	AccountName   string        `gorm:"type:varchar(255);" json:"account_name"`
	AccountNumber string        `gorm:"type:varchar(100);" json:"account_number"`
	TaxPercentage int           `gorm:"type:int;default:0" json:"tax_percentage"`
	JournalMethod JournalMethod `gorm:"type:journal_method; default:'automatic'" json:"journal_method" validate:"required"`
	BranchStatus  DataStatus    `gorm:"type:data_status;default:'inactive'" json:"branch_status"`
	LicenseDate   time.Time     `gorm:"not null" json:"license_date" validate:"required"`
	DefaultMember string        `gorm:"type:varchar(15);not null;" json:"default_member" validate:"required"`
	MemberName    string        `gorm:"type:varchar(100);not null" json:"member_name" validate:"required"`
}
