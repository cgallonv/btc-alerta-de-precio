package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/cgallonv/btc-alerta-de-precio/config"
	"github.com/cgallonv/btc-alerta-de-precio/internal/adapters"
	"github.com/cgallonv/btc-alerta-de-precio/internal/alerts"
	"github.com/cgallonv/btc-alerta-de-precio/internal/api"
	"github.com/cgallonv/btc-alerta-de-precio/internal/bitcoin"
	"github.com/cgallonv/btc-alerta-de-precio/internal/interfaces"
	"github.com/cgallonv/btc-alerta-de-precio/internal/notifications"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage/migrations"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage/repositories"
)

// AlertServiceAdapter adapts AlertManager to AlertService interface
type AlertServiceAdapter struct {
	*alerts.AlertManager
	db *storage.Database
}

func (a *AlertServiceAdapter) GetPriceHistory(limit int) ([]interfaces.PriceCacheEntry, error) {
	history, err := a.AlertManager.GetPriceHistory(limit)
	return history, err
}

func (a *AlertServiceAdapter) GetStats() (map[string]interface{}, error) {
	return a.db.GetStats()
}

// AlertEvaluator implements the AlertEvaluator interface
type AlertEvaluator struct{}

func (e *AlertEvaluator) ShouldTrigger(alert *storage.Alert, priceData *bitcoin.PriceData) bool {
	return alert.ShouldTrigger(priceData.Price, 0)
}

func main() {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize database
	db, err := storage.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Run migrations
	if err := migrations.MigrateTickerData(db.DB()); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// Create repositories
	tickerRepo := repositories.NewTickerRepository(db.DB())

	// Create storage handlers
	tickerStorage := bitcoin.NewTickerStorage(tickerRepo)

	// Create adapters
	configAdapter := adapters.NewConfigAdapter(cfg)

	// Create services
	notificationService := notifications.NewService(cfg, db)

	// Create alert manager (which creates its own price monitor)
	alertManager, err := alerts.NewAlertManager(
		configAdapter,
		notificationService,
		&AlertEvaluator{},
		db,
		db,
		tickerStorage,
	)
	if err != nil {
		log.Fatalf("Error creating alert manager: %v", err)
	}

	// Start alert manager (which starts price monitoring)
	if err := alertManager.Start(context.Background()); err != nil {
		log.Printf("Error starting alert manager: %v", err)
	}

	// Create alert service adapter
	alertService := &AlertServiceAdapter{
		AlertManager: alertManager,
		db:           db,
	}

	// Create API handler
	handler := api.NewHandler(alertService, configAdapter)

	// Create router
	router := gin.Default()

	// Configure router
	handler.SetupRoutes(router)

	// Start server
	port := cfg.Port
	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
