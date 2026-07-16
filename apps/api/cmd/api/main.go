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
	Version    = "0.2.0-dev"
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
	r.Use(gin.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.Tenant())

	api := r.Group("/api/v1")
	{
		api.GET("/version", handler.Version(Version, CommitHash, BuildTime))
		api.GET("/health", handler.Health(db))

		products := api.Group("/products")
		products.Use(middleware.Auth(cfg))
		{
			products.GET("", handler.ListProducts(db))
			products.POST("", handler.CreateProduct(db))
			products.GET("/:id", handler.GetProduct(db))
			products.PUT("/:id", handler.UpdateProduct(db))
			products.DELETE("/:id", handler.DeleteProduct(db))
		}

		categories := api.Group("/categories")
		categories.Use(middleware.Auth(cfg))
		{
			categories.GET("", handler.ListCategories(db))
			categories.POST("", handler.CreateCategory(db))
		}

		warehouses := api.Group("/warehouses")
		warehouses.Use(middleware.Auth(cfg))
		{
			warehouses.GET("", handler.ListWarehouses(db))
			warehouses.POST("", handler.CreateWarehouse(db))
		}

		stock := api.Group("/stock")
		stock.Use(middleware.Auth(cfg))
		{
			stock.POST("/grn", handler.CreateGRN(db))
			stock.POST("/issue", handler.CreateIssue(db))
			stock.POST("/transfer", handler.CreateTransfer(db))
			stock.POST("/adjust", handler.CreateAdjustment(db))
			stock.GET("/levels", handler.GetStockLevels(db))
			stock.GET("/movements", handler.GetStockMovements(db))
		}
	}

	addr := ":" + cfg.APIPort
	log.Printf("✓ Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
		os.Exit(1)
	}
}
