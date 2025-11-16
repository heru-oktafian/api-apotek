package models

import (
	time "time"

	gorm "gorm.io/gorm"
)

// UserBranch Struct untuk user_branch model yang akan disimpan di database
type UserBranch struct {
	UserID    string `gorm:"type:varchar(15);primaryKey" json:"user_id" validate:"required"`
	BranchID  string `gorm:"type:varchar(15);primaryKey" json:"branch_id" validate:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// UserBranchDetail Struct untuk user_branch model yang akan digunakan untuk menampilkan data pada function GetDetail
type UserBranchDetail struct {
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	BranchID   string `json:"branch_id"`
	BranchName string `json:"branch_name"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
}
