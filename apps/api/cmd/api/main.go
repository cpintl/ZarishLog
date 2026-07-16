package main

import (
	"log"
	"os"
	"runtime"

	"github.com/cpintl/zarishlog-api/internal/config"
	"github.com/cpintl/zarishlog-api/internal/database"
	"github.com/cpintl/zarishlog-api/internal/handler"
	"github.com/cpintl/zarishlog-api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	Version    = "0.3.0-dev"
	CommitHash = "unknown"
	BuildTime  = "unknown"
)

func main() {
	godotenv.Load()

	cfg := config.Load()

	log.Printf("╔═══════════════════════════════════════════════════════════════╗")
	log.Printf("║  ZarishLog API Server                                        ║")
	log.Printf("║  Version: %-45s ║", Version)
	log.Printf("║  Commit:  %-45s ║", CommitHash)
	log.Printf("║  Built:   %-45s ║", BuildTime)
	log.Printf("║  Go:      %-45s ║", runtime.Version())
	log.Printf("║  Arch:    %-45s ║", runtime.GOARCH+"/"+runtime.GOOS)
	log.Printf("╚═══════════════════════════════════════════════════════════════╝")

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Printf("✓ Database connected")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(gin.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.Tenant())

	api := r.Group("/api/v1")
	{
		api.GET("/version", handler.Version(Version, CommitHash, BuildTime))
		api.GET("/health", handler.Health(db))

		protected := api.Group("")
		protected.Use(middleware.Auth(cfg))
		protected.Use(middleware.Audit(db))

		{
			products := protected.Group("/products")
			products.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "logistics_officer"))
			{
				products.GET("", handler.ListProducts(db))
				products.POST("", handler.CreateProduct(db))
				products.POST("/import", handler.ImportProducts(db))
				products.GET("/search", handler.SearchProducts(db))
				products.GET("/:id", handler.GetProduct(db))
				products.PUT("/:id", handler.UpdateProduct(db))
				products.DELETE("/:id", handler.DeleteProduct(db))
			}

			categories := protected.Group("/categories")
			categories.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist"))
			{
				categories.GET("", handler.ListCategories(db))
				categories.POST("", handler.CreateCategory(db))
			}

			warehouses := protected.Group("/warehouses")
			warehouses.Use(middleware.RequireRole("admin", "warehouse_manager"))
			{
				warehouses.GET("", handler.ListWarehouses(db))
				warehouses.POST("", handler.CreateWarehouse(db))
				warehouses.GET("/:id", handler.GetWarehouse(db))
				warehouses.PUT("/:id", handler.UpdateWarehouse(db))
				warehouses.DELETE("/:id", handler.DeleteWarehouse(db))
				warehouses.GET("/:id/locations", handler.ListLocations(db))
				warehouses.GET("/:id/locations/tree", handler.ListLocationTree(db))
				warehouses.GET("/:warehouse_id/locations/:id", handler.GetLocation(db))
				warehouses.POST("/:warehouse_id/locations", handler.CreateLocation(db))
				warehouses.PUT("/:warehouse_id/locations/:id", handler.UpdateLocation(db))
				warehouses.DELETE("/:warehouse_id/locations/:id", handler.DeleteLocation(db))
				warehouses.GET("/:warehouse_id/locations/:id/constraints", handler.GetLocationConstraints(db))
				warehouses.PUT("/:warehouse_id/locations/:id/constraints", handler.UpsertLocationConstraints(db))
			}

			uoms := protected.Group("/uoms")
			uoms.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist"))
			{
				uoms.GET("", handler.ListUoMs(db))
				uoms.POST("", handler.CreateUoM(db))
				uoms.GET("/:id", handler.GetUoM(db))
				uoms.PUT("/:id", handler.UpdateUoM(db))
				uoms.DELETE("/:id", handler.DeleteUoM(db))
			}

			stock := protected.Group("/stock")
			stock.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "logistics_officer"))
			{
				stock.POST("/grn", handler.CreateGRN(db))
				stock.POST("/issue", handler.CreateIssue(db))
				stock.POST("/transfer", handler.CreateTransfer(db))
				stock.POST("/adjust", handler.CreateAdjustment(db))
				stock.GET("/levels", handler.GetStockLevels(db))
				stock.GET("/movements", handler.GetStockMovements(db))
				stock.GET("/batches/:id/trail", handler.GetBatchTrail(db))
				stock.GET("/expiring", handler.GetExpiringStock(db))
			}

			qa := protected.Group("/qa")
			qa.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				qa.POST("/inspections", handler.CreateInspection(db))
				qa.GET("/inspections", handler.ListInspections(db))
				qa.GET("/inspections/:id", handler.GetInspection(db))
				qa.POST("/inspections/:id/disposition", handler.CreateDisposition(db))
				qa.POST("/checklists", handler.CreateChecklistTemplate(db))
				qa.GET("/checklists", handler.ListChecklistTemplates(db))
				qa.GET("/checklists/:id", handler.GetChecklistTemplate(db))
			}
			assets := protected.Group("/assets")
			assets.Use(middleware.RequireRole("admin", "warehouse_manager", "logistics_officer"))
			{
				assets.POST("", handler.CreateAsset(db))
				assets.GET("", handler.ListAssets(db))
				assets.GET("/:id", handler.GetAsset(db))
				assets.PUT("/:id", handler.UpdateAsset(db))
				assets.DELETE("/:id", handler.DeleteAsset(db))
				assets.POST("/:id/custody", handler.TransferCustody(db))
				assets.POST("/:id/maintenance", handler.CreateAssetMaintenance(db))
				assets.GET("/:id/maintenance", handler.ListAssetMaintenance(db))
			}

			distributions := protected.Group("/distributions")
			distributions.Use(middleware.RequireRole("admin", "warehouse_manager", "logistics_officer"))
			{
				distributions.POST("", handler.CreateDistribution(db))
				distributions.GET("", handler.ListDistributions(db))
				distributions.GET("/:id", handler.GetDistribution(db))
			}
			replenishment := protected.Group("/replenishment")
			replenishment.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist"))
			{
				replenishment.POST("/amc", handler.CalculateAMC(db))
				replenishment.GET("/amc", handler.ListAMCCalculations(db))
				replenishment.GET("/amc/latest", handler.GetLatestAMC(db))
				replenishment.GET("/recommendations", handler.ListReorderRecommendations(db))
				replenishment.POST("/recommendations", handler.CreateReorderRecommendation(db))
				replenishment.PUT("/recommendations/:id/review", handler.MarkRecommendationReviewed(db))
				replenishment.GET("/forecasts", handler.ListForecastResults(db))
				replenishment.POST("/forecasts", handler.CreateForecastResult(db))
			}
		}
	}

	addr := ":" + cfg.APIPort
	log.Printf("✓ Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
		os.Exit(1)
	}
}
