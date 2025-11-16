package scheduler

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// InitScheduler mulaiin semua job terjadwal
func InitScheduler(db *gorm.DB) *cron.Cron {
	c := cron.New(cron.WithLocation(time.UTC)) // Akan kita konversi ke WIB

	// === CONTOH JOB ===

	// 1. Backup DB setiap pukul 02:00 WIB (UTC+7 → 19:00 UTC kemarin)
	// c.AddFunc("0 19 * * *", func() {
	// 	log.Println("[SCHEDULER] Mulai backup database...")
	// 	BackupDatabase()
	// })

	// 2. Hitung asset value & simpan dalam tabel sys_dayly_asset setiap pukul 21:35 WIB (UTC-7 → 15:35 UTC kemarin)
	c.AddFunc("30 17 * * *", func() {
		err := AssetCounter(db) // Ganti 'nil' dengan instance *gorm.DB Anda
		if err != nil {
			log.Println("[SCHEDULER] Gagal menghitung asset:", err)
		} else {
			log.Println("[SCHEDULER] Perhitungan asset berhasil dihitung dan disimpan.")
		}
	})

	// 3. Generate laporan harian pukul 06:00 WIB (23:00 UTC)
	// c.AddFunc("0 23 * * *", func() {
	// 	log.Println("[SCHEDULER] Generate laporan harian...")
	// 	GenerateDailyReport()
	// })

	// 4. Reset cache Redis setiap pukul 00:00 WIB
	// c.AddFunc("0 17 * * *", func() { // 17:00 UTC = 00:00 WIB
	// 	log.Println("[SCHEDULER] Clearing Redis cache...")
	// 	ClearRedisCache()
	// })

	// 5. Cek expired promo tiap 10 menit
	// c.AddFunc("*/10 * * * *", func() {
	// 	log.Println("[SCHEDULER] Cek promo kadaluarsa...")
	// 	DeactivateExpiredPromos()
	// })

	c.Start()
	log.Println("[SCHEDULER] Semua job terjadwal aktif!")
	return c
}
