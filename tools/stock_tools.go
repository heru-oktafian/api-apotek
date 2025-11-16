package tools

import (
	"errors"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/api-retail/reports"
	"gorm.io/gorm"
)

// Tambah stock product
func AddProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}
	product.Stock += qty
	return db.Save(&product).Error
}

// Kurangi stock product
func ReduceProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}

	if product.Stock < qty {
		return errors.New("insufficient stock")
	}
	product.Stock -= qty

	if err := db.Save(&product).Error; err != nil {
		return err
	}

	return nil
}

// SubtractProductStock menambah stok produk
func SubtractProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}

	product.Stock += qty

	if err := db.Save(&product).Error; err != nil {
		return err
	}

	return nil
}

// ZeroProductStock kosongkan stok produk
func ZeroProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}

	product.Stock = 0

	if err := db.Save(&product).Error; err != nil {
		return err
	}

	return nil
}

// UpdateProductPriceIfHigher memperbarui harga produk jika harga baru lebih tinggi
func UpdateProductPriceIfHigher(db *gorm.DB, productId string, newPrice int) error {
	var product models.Product

	if err := db.First(&product, "id = ?", productId).Error; err != nil {
		return err
	}

	if newPrice > product.PurchasePrice {
		product.PurchasePrice = newPrice
		return db.Save(&product).Error
	}

	// Tidak update jika harga baru lebih rendah
	return nil
}

// RecalculateTotalPurchase menghitung ulang total pembelian berdasarkan item
func RecalculateTotalPurchase(db *gorm.DB, purchaseID string) error {
	var total int64

	// Hitung total sub_total dari purchase_items
	err := db.Model(&models.PurchaseItems{}).
		Where("purchase_id = ?", purchaseID).
		Select("COALESCE(SUM(sub_total), 0)").
		Scan(&total).Error

	if err != nil {
		return err
	}

	// Update ke purchases
	if err := db.Model(&models.Purchases{}).
		Where("id = ?", purchaseID).
		Update("total_purchase", total).Error; err != nil {
		return err
	}

	// Ambil purchase lengkap buat update report
	var purchase models.Purchases
	if err := db.First(&purchase, "id = ?", purchaseID).Error; err != nil {
		return err
	}

	// Update transaction_reports juga
	if err := reports.SyncPurchaseReport(db, purchase); err != nil {
		return err
	}

	return nil
}
