package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/adapters"
	"btc-alerta-de-precio/internal/alerts"
	"btc-alerta-de-precio/internal/api"
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
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

	// Create adapters
	configAdapter := adapters.NewConfigAdapter(cfg)

	// Create services
	notificationService := notifications.NewService(cfg, db)

	// Create price monitor
	priceMonitor := alerts.NewPriceMonitor(configAdapter, 20)

	// Create alert manager
	alertManager, err := alerts.NewAlertManager(
		configAdapter,
		notificationService,
		&AlertEvaluator{},
		db,
		db,
	)
	if err != nil {
		log.Fatalf("Error creating alert manager: %v", err)
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

	// Start price monitoring
	if err := priceMonitor.Start(context.Background()); err != nil {
		log.Printf("Error starting price monitor: %v", err)
	}

	// Start server
	port := cfg.Port
	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
