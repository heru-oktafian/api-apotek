package models

import "time"

// SysDaylyAsset model yang akan disimpan di database
type DailyAsset struct {
	ID         string    `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	AssetDate  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"asset_date"`
	AssetValue int       `gorm:"type:int;not null;default:0" json:"asset_value" validate:"required"`
	BranchId   string    `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// SysDaylyAsset model yang akan disimpan di database
type AllDailyAsset struct {
	ID         string    `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	AssetDate  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"asset_date"`
	AssetValue int       `gorm:"type:int;not null;default:0" json:"asset_value" validate:"required"`
	BranchId   string    `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	BranchName string    `gorm:"unique;not null" json:"branch_name"`
}

// SysDaylyAsset model yang akan disimpan di database
type DetailDailyAsset struct {
	ID         string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	AssetDate  string `json:"asset_date"`
	AssetValue int    `gorm:"type:int;not null;default:0" json:"asset_value" validate:"required"`
	BranchId   string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	BranchName string `gorm:"unique;not null" json:"branch_name"`
}
