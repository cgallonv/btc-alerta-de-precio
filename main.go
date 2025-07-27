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
	"btc-alerta-de-precio/internal/interfaces"
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
	notificationRepo := adapters.NewGormNotificationRepository(db)
	statsRepo := adapters.NewGormStatsRepository(db)

	// Create service adapters
	bitcoinClient := bitcoin.NewClient()
	priceClient := adapters.NewBitcoinClientAdapter(bitcoinClient)
	notificationService := notifications.NewService(cfg, db)
	notificationSender := adapters.NewNotificationServiceAdapter(notificationService)
	configProvider := adapters.NewConfigAdapter(cfg)
	alertEvaluator := adapters.NewAlertEvaluator()

	// Create new refactored services with cache (20 entries for price history)
	priceMonitor := alerts.NewPriceMonitor(priceClient, configProvider, 20)
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
	}

	apiHandler := api.NewHandler(serviceAdapter, configProvider)
	apiHandler.SetupRoutes(router)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Graceful shutdown handling
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down server...")
		cancel() // Cancel context to stop monitoring

		// Stop alert manager
		if err := alertManager.Stop(); err != nil {
			log.Printf("Error stopping alert manager: %v", err)
		}

		// Shutdown HTTP server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("📊 Visit http://localhost:%s for the web interface", cfg.Port)
	log.Printf("🔔 Bitcoin price monitoring is active every %v", cfg.CheckInterval)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %v", err)
	}

	log.Println("✅ Server shutdown complete")
}

// AlertServiceAdapter adapts the new AlertManager to the old AlertService interface
// This provides backward compatibility for the API handler
type AlertServiceAdapter struct {
	alertManager *alerts.AlertManager
	statsRepo    *adapters.GormStatsRepository
}

// Methods to satisfy the alerts.Service interface that the API handler expects

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

func (a *AlertServiceAdapter) ResetAlert(alertID uint) error {
	return a.alertManager.ResetAlert(alertID)
}

func (a *AlertServiceAdapter) TestAlert(id uint) error {
	return a.alertManager.TestAlert(id)
}

func (a *AlertServiceAdapter) GetCurrentPrice() (*bitcoin.PriceData, error) {
	return a.alertManager.GetCurrentPrice()
}

func (a *AlertServiceAdapter) GetPriceHistory(limit int) ([]interfaces.PriceCacheEntry, error) {
	return a.alertManager.GetPriceHistory(limit), nil
}

func (a *AlertServiceAdapter) GetCurrentPercentage() float64 {
	return a.alertManager.GetCurrentPercentage()
}

func (a *AlertServiceAdapter) GetStats() (map[string]interface{}, error) {
	return a.statsRepo.GetStats()
}

func (a *AlertServiceAdapter) IsMonitoring() bool {
	return a.alertManager.IsMonitoring()
}
