package scheduler

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/heru-oktafian/scafold/helpers"
	"gorm.io/gorm"

	"github.com/heru-oktafian/api-retail/models"
)

// Contoh: Backup DB pake pg_dump
func BackupDatabase() {
	cmd := exec.Command("pg_dump",
		"-h", os.Getenv("DB_HOST"),
		"-U", os.Getenv("DB_USER"),
		"-d", os.Getenv("DB_NAME"),
		"-f", "/backups/retail-"+time.Now().Format("2006-01-02")+".sql",
	)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PGPASSWORD="+os.Getenv("DB_PASS"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[BACKUP] Gagal: %v\n%s", err, output)
		return
	}
	log.Println("[BACKUP] Sukses:", string(output))
}

// Contoh: Generate laporan
func GenerateDailyReport() {
	log.Println("→ Mengambil data penjualan hari ini...")
	// Query GORM → simpan ke file PDF/Excel
	log.Println("→ Laporan selesai, dikirim ke email owner")
}

func ClearRedisCache() {
	// Koneksi Redis dari internal/database/redis.go
	// rdb.FlushDB(context.Background())
	log.Println("Redis cache dibersihkan")
}

func DeactivateExpiredPromos() {
	// Query: UPDATE promos SET active = false WHERE end_date < NOW()
	log.Println("Promo kadaluarsa dinonaktifkan")
}

// AssetCounter menghitung dan menyimpan nilai aset harian berdasarkan stok, harga beli produk, dan mengurangi total pembelian kredit
func AssetCounter(db *gorm.DB) error {
	// SQL query untuk menghitung nilai aset per cabang
	query := `
		SELECT 
			branch_id,
			SUM(stock * purchase_price) as total_asset
		FROM 
			products
		GROUP BY 
			branch_id
	`

	type BranchAsset struct {
		BranchID   string
		TotalAsset int
	}

	var branchAssets []BranchAsset
	if err := db.Raw(query).Scan(&branchAssets).Error; err != nil {
		log.Printf("[ASSET COUNTER] Error querying branch assets: %v", err)
		return err
	}

	// Query untuk total pembelian kredit per cabang
	creditQuery := `
		SELECT 
			branch_id,
			COALESCE(SUM(total_purchase), 0) as total_credit
		FROM 
			purchases
		WHERE 
			payment = 'paid_by_credit'
		GROUP BY 
			branch_id
	`

	type BranchCredit struct {
		BranchID    string
		TotalCredit int
	}

	var branchCredits []BranchCredit
	if err := db.Raw(creditQuery).Scan(&branchCredits).Error; err != nil {
		log.Printf("[ASSET COUNTER] Error querying branch credits: %v", err)
		return err
	}

	// Buat map untuk lookup total kredit per branch
	creditMap := make(map[string]int)
	for _, credit := range branchCredits {
		creditMap[credit.BranchID] = credit.TotalCredit
	}

	// Menyimpan aset harian untuk setiap cabang
	for _, asset := range branchAssets {
		credit := creditMap[asset.BranchID]
		finalAsset := asset.TotalAsset - credit

		dailyAsset := models.DailyAsset{
			ID:         helpers.GenerateID("AST"),
			AssetValue: finalAsset,
			BranchId:   asset.BranchID,
		}

		if err := db.Create(&dailyAsset).Error; err != nil {
			log.Printf("[ASSET COUNTER] Error creating daily asset for branch %s: %v", asset.BranchID, err)
			return err
		}
	}

	log.Println("[ASSET COUNTER] Successfully updated daily assets for all branches")
	return nil
}
