package models

type ComboboxUnits struct {
	UnitID   string `gorm:"type:varchar(15);primaryKey" json:"unit_id" validate:"required"`
	UnitName string `gorm:"type:varchar(255);not null" json:"unit_name" validate:"required"`
}

type ComboboxProductCategories struct {
	PCID     uint   `gorm:"primaryKey;autoIncrement" json:"pc_id" validate:"required"`
	Category string `gorm:"type:varchar(255);not null" json:"category" validate:"required"`
}

type ComboboxSupplierCategories struct {
	SCID     uint   `gorm:"primaryKey;autoIncrement" json:"sc_id" validate:"required"`
	Category string `gorm:"type:varchar(255);not null" json:"category" validate:"required"`
}

type ComboboxMemberCategories struct {
	MCID     uint   `gorm:"primaryKey;autoIncrement" json:"mc_id" validate:"required"`
	Category string `gorm:"type:varchar(255);not null" json:"category" validate:"required"`
}

type ComboboxSuppliers struct {
	SupplierID   string `gorm:"type:varchar(15);primaryKey" json:"supplier_id" validate:"required"`
	SupplierName string `gorm:"type:varchar(255);not null" json:"supplier_name" validate:"required"`
}

type ComboboxMembers struct {
	MemberID   string `gorm:"type:varchar(15);primaryKey" json:"member_id" validate:"required"`
	MemberName string `gorm:"type:varchar(255);not null" json:"member_name" validate:"required"`
}

// ProductCombo model
type ComboboxProducts struct {
	ProID    string `gorm:"type:varchar(15);primaryKey" json:"pro_id" validate:"required"`
	ProName  string `gorm:"type:varchar(100);not null" json:"pro_name" validate:"required"`
	Stock    int    `gorm:"type:int;not null;default:0" json:"stock"`
	UnitId   string `gorm:"type:varchar(15);not null" json:"unit_id" validate:"required"`
	UnitName string `gorm:"type:varchar(100);not null" json:"unit_name" validate:"required"`
	Price    int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
}
