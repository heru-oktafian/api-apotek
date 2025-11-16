package models

// Unit model yang akan disimpan di database
type Unit struct {
	ID       string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Name     string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	BranchID string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// Unit model yang akan disimpan ditampilkan
type AllUnit struct {
	UnitID   string `gorm:"type:varchar(15);primaryKey" json:"unit_id" validate:"required"`
	UnitName string `gorm:"type:varchar(100);not null" json:"unit_name" validate:"required"`
}

// UnitCombo model yang akan ditampilkan di data combobox
type UnitCombo struct {
	UnitId   string `json:"unit_id"`
	UnitName string `json:"unit_name"`
}

// ProductUnitResponseItem merupakan representasi dari unit yang ditampilkan dalam transaksi pembelian berdasarkan id produk
type ProductUnitResponseItem struct {
	UnitId        string `json:"unit_id"`
	UnitName      string `json:"unit_name"`
	PurchasePrice int    `json:"purchase_price"`
}
