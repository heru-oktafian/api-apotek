package main

import (
	log "log"
	os "os"
	"strconv"

	models "github.com/heru-oktafian/api-retail/models"
	routes "github.com/heru-oktafian/api-retail/routes"
	scheduler "github.com/heru-oktafian/api-retail/scheduler"
	config "github.com/heru-oktafian/scafold/config"
	env "github.com/heru-oktafian/scafold/env"
	framework "github.com/heru-oktafian/scafold/framework"
	middlewares "github.com/heru-oktafian/scafold/middlewares"
	utils "github.com/heru-oktafian/scafold/utils"
)

func main() {
	// Initialize timezone
	utils.InitTimezone()

	// Load .env file
	env.Load(".env")

	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET_KEY not set in .env file")
	}

	// Initialize database connection
	config.KoneksiPG(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	// Migrasi model dengan pengecekan tabel yang sudah ada
	for _, model := range []interface{}{
		&models.AnotherIncomes{},
		&models.BalanceReport{},
		&models.Branch{},
		&models.BuyReturnItems{},
		&models.BuyReturns{},
		&models.DailyProfitReport{},
		&models.DailyAsset{},
		&models.Expenses{},
		&models.FirstStockItems{},
		&models.FirstStocks{},
		&models.MemberCategory{},
		&models.Member{},
		&models.OpnameItems{},
		&models.Opnames{},
		&models.ProductCategory{},
		&models.Product{},
		&models.PurchaseItems{},
		&models.Purchases{},
		&models.SaleItems{},
		&models.Sales{},
		&models.SupplierCategory{},
		&models.Supplier{},
		&models.TransactionReports{},
		&models.UnitConversion{},
		&models.Unit{},
		&models.UserBranch{},
		&models.User{},
	} {
		// Cek apakah tabel sudah ada
		if !config.DB.Migrator().HasTable(model) {
			log.Printf("Membuat tabel untuk model %T...", model)
			if err := config.DB.AutoMigrate(model); err != nil {
				log.Fatalf("Gagal migrasi model %T: %v", model, err)
			}
			log.Printf("Tabel untuk model %T berhasil dibuat", model)
		} else {
			log.Printf("Tabel untuk model %T sudah ada, melanjutkan ke model berikutnya", model)
		}
	}

	// Initialize Redis connection
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if dbInt, err := strconv.Atoi(dbStr); err == nil {
			redisDB = dbInt
		}
	}

	// Default to DB 0 if REDIS_DB is not set or invalid
	config.KoneksiRedis(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), redisDB)

	// Initialize Scheduler
	scheduler.InitScheduler(config.DB)

	// Get port from environment
	serverPort := os.Getenv("PORT")

	// Start the application
	app := framework.New()

	app.Use(middlewares.CORS())
	// app.Use(middlewares.Logger())

	// Routes
	routes.SysAuthRoutes(app)
	routes.SysMenuRoutes(app)
	routes.SysBranchRoutes(app)
	routes.SysUserBranchRoutes(app)
	routes.SysUserRoutes(app)
	routes.SysMemberCategoryRoutes(app)
	routes.SysMemberRoutes(app)
	routes.SysDashboardRoutes(app)
	routes.SysReportRoutes(app)
	routes.DailyAssetRoutes(app)
	routes.AuditFirstStockRoutes(app)
	routes.AuditFirstStockWithItems(app)
	routes.AuditFirstStockItemRoutes(app)
	routes.AuditMobileOpnamesRoutes(app)
	routes.AuditOpnameRoutes(app)
	routes.AuditOpnameItemRoutes(app)
	routes.CmbProductOpnameRoutes(app)
	routes.MasterProductCategoryRoutes(app)
	routes.MasterSupplierCategoryRoutes(app)
	routes.MasterSupplierRoutes(app)
	routes.MasterUnitRoutes(app)
	routes.MasterProductRoutes(app)
	routes.MasterUnitConversionRoutes(app)
	routes.TransAnotherIncomeRoutes(app)
	routes.TransExpenseRoutes(app)
	routes.TransPurchaseRoutes(app)
	routes.TransPurchaseItemRoutes(app)
	routes.TransSaleRoutes(app)
	routes.TransSaleDetailRoutes(app)
	routes.TransSaleItemRoutes(app)
	routes.TransBuyReturnRoutes(app)
	routes.TransSaleReturnRoutes(app)
	routes.CmbProdSaleReturn(app)
	routes.CmbSaleRoute(app)
	routes.CmbProdBuyReturn(app)
	routes.CmbPurchaseRoute(app)
	routes.CmbMemberCategoryRoutes(app)
	routes.CmbMemberRoutes(app)
	routes.CmbSupplierCategoryRoutes(app)
	routes.CmbSupplierRoutes(app)
	routes.CmbUnitRoutes(app)
	routes.CmbProductCategoryRoutes(app)
	routes.CmbProdSaleRoutes(app)
	routes.CmbProdPurchaseRoutes(app)
	routes.CmbProdConvRoutes(app)
	routes.CmbUnitConvRoutes(app)
	routes.GetBranchByUserId(app)
	routes.CobaRoutes(app)

	// Start listening on the specified port
	app.Listen(":"+serverPort, os.Getenv("APPNAME"))
}
