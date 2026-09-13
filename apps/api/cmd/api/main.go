package main

import (
	"log"
	"os"
	"runtime"

	"github.com/cpintl/ZarishLog/apps/api/internal/config"
	"github.com/cpintl/ZarishLog/apps/api/internal/database"
	"github.com/cpintl/ZarishLog/apps/api/internal/handler"
	"github.com/cpintl/ZarishLog/apps/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	Version    = "1.0.0"
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

			regulatory := protected.Group("/policy/regulatory-approvals")
			regulatory.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				regulatory.POST("", handler.CreateRegulatoryApproval(db))
				regulatory.GET("", handler.ListRegulatoryApprovals(db))
				regulatory.GET("/expiring", handler.ListExpiringRegulatoryApprovals(db))
				regulatory.GET("/:id", handler.GetRegulatoryApproval(db))
			}

			deviations := protected.Group("/policy/deviations")
			deviations.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				deviations.POST("", handler.CreateDeviation(db))
				deviations.GET("", handler.ListDeviations(db))
				deviations.GET("/:id", handler.GetDeviation(db))
			}

			capa := protected.Group("/policy/capa")
			capa.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				capa.POST("", handler.CreateCAPAAction(db))
				capa.GET("", handler.ListCAPAActions(db))
				capa.GET("/:id", handler.GetCAPAAction(db))
				capa.POST("/:id/effectiveness", handler.VerifyCAPAEffectiveness(db))
			}

			training := protected.Group("/policy/training")
			training.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				training.POST("", handler.CreateTrainingRecord(db))
				training.GET("", handler.ListTrainingRecords(db))
			}

			tempMon := protected.Group("/temperature-monitoring")
			tempMon.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "logistics_officer", "quality_officer"))
			{
				tempMon.POST("", handler.CreateTemperatureMonitoringEntry(db))
				tempMon.GET("", handler.ListTemperatureMonitoringEntries(db))
			}

			tempExc := protected.Group("/temperature-excursions")
			tempExc.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				tempExc.POST("", handler.CreateTemperatureExcursion(db))
				tempExc.GET("", handler.ListTemperatureExcursions(db))
				tempExc.GET("/:id", handler.GetTemperatureExcursion(db))
				tempExc.POST("/:id/disposition", handler.UpdateTemperatureExcursionDisposition(db))
			}

			stockReleases := protected.Group("/stock-releases")
			stockReleases.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				stockReleases.POST("", handler.CreateStockReleaseRecord(db))
				stockReleases.GET("", handler.ListStockReleaseRecords(db))
				stockReleases.GET("/:id", handler.GetStockReleaseRecord(db))
			}

			controlledStock := protected.Group("/controlled-stock")
			controlledStock.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist"))
			{
				controlledStock.POST("", handler.CreateControlledStockRegisterEntry(db))
				controlledStock.GET("", handler.ListControlledStockRegister(db))
				controlledStock.GET("/:id", handler.GetControlledStockRegisterEntry(db))
			}

			donations := protected.Group("/donations")
			donations.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				donations.POST("", handler.CreateDonation(db))
				donations.GET("", handler.ListDonations(db))
				donations.GET("/:id", handler.GetDonation(db))
				donations.POST("/:id/decision", handler.UpdateDonationDecision(db))
			}

			complaints := protected.Group("/complaints")
			complaints.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				complaints.POST("", handler.CreateComplaint(db))
				complaints.GET("", handler.ListComplaints(db))
				complaints.GET("/:id", handler.GetComplaint(db))
			}

			recalls := protected.Group("/recalls")
			recalls.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist", "quality_officer"))
			{
				recalls.POST("", handler.CreateRecall(db))
				recalls.GET("", handler.ListRecalls(db))
				recalls.GET("/:id", handler.GetRecall(db))
			}

			waybills := protected.Group("/dispatch-waybills")
			waybills.Use(middleware.RequireRole("admin", "warehouse_manager", "logistics_officer"))
			{
				waybills.POST("", handler.CreateDispatchWaybill(db))
				waybills.GET("", handler.ListDispatchWaybills(db))
				waybills.GET("/:id", handler.GetDispatchWaybill(db))
			}

			deliveryConfirm := protected.Group("/delivery-confirmations")
			deliveryConfirm.Use(middleware.RequireRole("admin", "warehouse_manager", "logistics_officer"))
			{
				deliveryConfirm.POST("", handler.CreateDeliveryConfirmation(db))
				deliveryConfirm.GET("", handler.ListDeliveryConfirmations(db))
				deliveryConfirm.GET("/:id", handler.GetDeliveryConfirmation(db))
			}

			shortExpiry := protected.Group("/short-expiry-reviews")
			shortExpiry.Use(middleware.RequireRole("admin", "warehouse_manager", "pharmacist"))
			{
				shortExpiry.POST("", handler.CreateShortExpiryReview(db))
				shortExpiry.GET("", handler.ListShortExpiryReviews(db))
			}

			emergencyPlans := protected.Group("/emergency-plans")
			emergencyPlans.Use(middleware.RequireRole("admin", "warehouse_manager"))
			{
				emergencyPlans.POST("", handler.CreateEmergencyPlan(db))
				emergencyPlans.GET("", handler.ListEmergencyPlans(db))
				emergencyPlans.GET("/:id", handler.GetEmergencyPlan(db))
			}

			changeControls := protected.Group("/change-controls")
			changeControls.Use(middleware.RequireRole("admin", "warehouse_manager", "quality_officer"))
			{
				changeControls.POST("", handler.CreateChangeControl(db))
				changeControls.GET("", handler.ListChangeControls(db))
				changeControls.GET("/:id", handler.GetChangeControl(db))
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

			users := protected.Group("/users")
			users.Use(middleware.RequireRole("admin"))
			{
				users.GET("", handler.ListUsers(db))
				users.POST("", handler.CreateUser(db))
				users.GET("/:id", handler.GetUser(db))
				users.PUT("/:id", handler.UpdateUser(db))
				users.DELETE("/:id", handler.DeactivateUser(db))
				users.POST("/:id/roles", handler.AssignUserRole(db))
				users.DELETE("/:id/roles", handler.RemoveUserRole(db))
			}

			protected.GET("/roles", handler.ListRoles(db))
			protected.GET("/permissions", handler.ListPermissions(db))
		}
	}

	addr := ":" + cfg.APIPort
	log.Printf("✓ Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
		os.Exit(1)
	}
}
