package tools

import (
	"log"
	"net/http"
	"time"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/api-retail/reports"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
	"gorm.io/gorm"
)

// AutoCleanupOpnames will delete any opnames older than 2 hours without opname items
func AutoCleanupOpnames(db *gorm.DB) error {
	var opnames []models.Opnames

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Ambil semua opnames yang tidak punya opname_items
	err := db.
		Where("created_at < ?", nowWIB.Add(-2*time.Hour)).
		Find(&opnames).Error
	if err != nil {
		return err
	}

	for _, opname := range opnames {
		var itemCount int64
		db.Model(&models.OpnameItems{}).
			Where("opname_id = ?", opname.ID).
			Count(&itemCount)

		if itemCount == 0 {
			log.Printf("🧹 Auto-cleaning orphan opname: %s\n", opname.ID)

			// Hapus transaction_report
			db.Where("id = ?", opname.ID).Delete(&models.TransactionReports{})

			// Hapus opname
			db.Where("id = ?", opname.ID).Delete(&models.Opnames{})
		}
	}

	return nil
}

// Input struct untuk request body
type CreateOpnameItemInput struct {
	OpnameId    string `json:"opname_id" validate:"required"`
	ProductId   string `json:"product_id" validate:"required"`
	Qty         int    `json:"qty" validate:"required"`
	ExpiredDate string `json:"expired_date" validate:"required"`
}

// Input struct untuk request body
type CreateOpnameItemUpdate struct {
	OpnameId    string `json:"opname_id" validate:"required"`
	ProductId   string `json:"product_id" validate:"required"`
	Qty         int    `json:"qty" validate:"required"`
	Price       int    `json:"price" validate:"required"`
	ExpiredDate string `json:"expired_date" validate:"required"`
}

// Response menampilkan satu opname beserta semua item-nya
type ResponseOpnameWithItemsResponse struct {
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	OpnameId    string      `json:"opname_id"`
	Description string      `json:"description"`
	OpnameDate  string      `json:"opname_date"`
	TotalOpname int         `json:"total_opname"`
	Payment     string      `json:"payment"`
	Items       interface{} `json:"items"`
}

// JSONFirstStockWithItemsResponse sends a standard JSON response format / structure
func JSONOpnameWithItemsResponse(c *framework.Ctx, status int, message string, opname_id string, description string, opname_date string, total_opname int, payment string, items interface{}) error {
	resp := ResponseOpnameWithItemsResponse{
		Status:      http.StatusText(status),
		Message:     message,
		OpnameId:    opname_id,
		Description: description,
		OpnameDate:  opname_date,
		TotalOpname: total_opname,
		Payment:     payment,
		Items:       items,
	}
	return responses.JSONResponse(c, status, message, resp)
}

// Opname stock product
func OpnameProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}
	product.Stock = qty
	return db.Save(&product).Error
}

// RecalculateTotalOpname menghitung ulang total opname
func RecalculateTotalOpname(db *gorm.DB, opnameID string) error {
	var totalAdjustment int64 // Ubah nama variabel untuk merefleksikan 'adjustment'

	// Hitung total (sub_total_exist - sub_total) dari opname_items
	// Menggunakan alias untuk kolom-kolom agar jelas dan melakukan operasi pengurangan langsung di query SQL
	err := db.Table("opname_items").
		Where("opname_id = ?", opnameID).
		Select("COALESCE(SUM(sub_total - sub_total_exist), 0)"). // Perhitungan yang diminta
		Scan(&totalAdjustment).Error                             // Scan ke variabel totalAdjustment

	if err != nil {
		return err
	}

	// Update ke opnames
	// Gunakan totalAdjustment untuk mengupdate kolom total_opname
	if err := db.Model(&models.Opnames{}).
		Where("id = ?", opnameID).
		Update("total_opname", totalAdjustment).Error; err != nil {
		return err
	}

	// Ambil opname lengkap buat update report
	var opname models.Opnames
	if err := db.First(&opname, "id = ?", opnameID).Error; err != nil {
		return err
	}

	// Update transaction_reports juga
	if err := reports.SyncOpnameReport(db, opname); err != nil {
		return err
	}

	return nil
}
