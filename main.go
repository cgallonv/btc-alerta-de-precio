package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/adapters"
	"btc-alerta-de-precio/internal/alerts"
	"btc-alerta-de-precio/internal/api"
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize database
	db, err := storage.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Create repository adapters
	alertRepo := adapters.NewGormAlertRepository(db)
	priceRepo := adapters.NewGormPriceRepository(db)
	notificationRepo := adapters.NewGormNotificationRepository(db)
	statsRepo := adapters.NewGormStatsRepository(db)

	// Create service adapters
	bitcoinClient := bitcoin.NewClient()
	priceClient := adapters.NewBitcoinClientAdapter(bitcoinClient)
	notificationService := notifications.NewService(cfg, db)
	notificationSender := adapters.NewNotificationServiceAdapter(notificationService)
	configProvider := adapters.NewConfigAdapter(cfg)
	alertEvaluator := adapters.NewAlertEvaluator()

	// Create new refactored services
	priceMonitor := alerts.NewPriceMonitor(priceClient, priceRepo, configProvider)
	alertManager := alerts.NewAlertManager(
		alertRepo,
		notificationRepo,
		notificationSender,
		alertEvaluator,
		priceMonitor,
		configProvider,
	)

	// Create main context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start alert manager (which starts price monitoring)
	if err := alertManager.Start(ctx); err != nil {
		log.Fatalf("Error starting alert manager: %v", err)
	}

	// Configure and start web server
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Create a service adapter for the API handler (temporary compatibility layer)
	serviceAdapter := &AlertServiceAdapter{
		alertManager: alertManager,
		statsRepo:    statsRepo.(*adapters.GormStatsRepository),
		priceRepo:    priceRepo.(*adapters.GormPriceRepository),
	}

	apiHandler := api.NewHandler(serviceAdapter)
	apiHandler.SetupRoutes(router)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server started on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down application...")

	// Cancel context to stop monitoring
	cancel()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	// Stop alert manager
	if err := alertManager.Stop(); err != nil {
		log.Printf("Error stopping alert manager: %v", err)
	}

	log.Println("Application shut down successfully")
}

// AlertServiceAdapter provides compatibility with the existing API handler
// This is a temporary adapter that will be removed once we refactor the API layer
type AlertServiceAdapter struct {
	alertManager *alerts.AlertManager
	statsRepo    *adapters.GormStatsRepository
	priceRepo    *adapters.GormPriceRepository
}

// Methods to satisfy the alerts.Service interface that the API handler expects

// Add missing IsMonitoring method
func (a *AlertServiceAdapter) IsMonitoring() bool {
	return a.alertManager.IsMonitoring()
}

func (a *AlertServiceAdapter) CreateAlert(alert *storage.Alert) error {
	return a.alertManager.CreateAlert(alert)
}

func (a *AlertServiceAdapter) GetAlert(id uint) (*storage.Alert, error) {
	return a.alertManager.GetAlert(id)
}

func (a *AlertServiceAdapter) GetAlerts() ([]storage.Alert, error) {
	return a.alertManager.GetAlerts()
}

func (a *AlertServiceAdapter) UpdateAlert(alert *storage.Alert) error {
	return a.alertManager.UpdateAlert(alert)
}

func (a *AlertServiceAdapter) DeleteAlert(id uint) error {
	return a.alertManager.DeleteAlert(id)
}

func (a *AlertServiceAdapter) ToggleAlert(id uint) error {
	return a.alertManager.ToggleAlert(id)
}

func (a *AlertServiceAdapter) GetCurrentPrice() (*bitcoin.PriceData, error) {
	return a.alertManager.GetCurrentPrice()
}

func (a *AlertServiceAdapter) GetPriceHistory(limit int) ([]storage.PriceHistory, error) {
	return a.priceRepo.GetPriceHistory(limit)
}

func (a *AlertServiceAdapter) GetStats() (map[string]interface{}, error) {
	stats, err := a.statsRepo.GetStats()
	if err != nil {
		return nil, err
	}

	// Add monitoring information
	stats["monitoring_active"] = a.alertManager.IsMonitoring()

	// Add current price information
	if currentPrice, err := a.alertManager.GetCurrentPrice(); err == nil {
		stats["current_price"] = currentPrice.Price
		stats["current_price_source"] = currentPrice.Source
		stats["current_price_time"] = currentPrice.Timestamp
	}

	return stats, nil
}

func (a *AlertServiceAdapter) TestAlert(id uint) error {
	return a.alertManager.TestAlert(id)
}

func (a *AlertServiceAdapter) ResetAlert(alertID uint) error {
	return a.alertManager.ResetAlert(alertID)
}

func (a *AlertServiceAdapter) GetCurrentPercentage() float64 {
	return a.alertManager.GetCurrentPercentage()
}
