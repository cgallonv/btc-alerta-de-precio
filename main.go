package main

import (
	"context"
	"fmt"
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

// AlertServiceAdapter adapts AlertManager to AlertService interface
type AlertServiceAdapter struct {
	*alerts.AlertManager
	statsRepo *adapters.GormStatsRepository
}

func (a *AlertServiceAdapter) GetPriceHistory(limit int) ([]interfaces.PriceCacheEntry, error) {
	return a.AlertManager.GetPriceHistory(limit), nil
}

func (a *AlertServiceAdapter) GetStats() (map[string]interface{}, error) {
	return a.statsRepo.GetStats()
}

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

	// Create API handler with adapter
	alertService := &AlertServiceAdapter{
		AlertManager: alertManager,
		statsRepo:    statsRepo.(*adapters.GormStatsRepository),
	}
	apiHandler := api.NewHandler(alertService, configProvider)
	apiHandler.SetupRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		log.Printf("📊 Visit http://localhost:%s for the web interface", cfg.Port)
		log.Printf("🔔 Bitcoin price monitoring is active every %v", cfg.CheckInterval)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	// Stop alert manager
	if err := alertManager.Stop(); err != nil {
		log.Printf("Error stopping alert manager: %v", err)
	}

	// Stop price monitor
	if err := priceMonitor.Stop(); err != nil {
		log.Printf("Error stopping price monitor: %v", err)
	}

	// Shutdown server
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	log.Println("✅ Server shutdown complete")
}
