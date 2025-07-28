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
	db, err := storage.NewDatabase()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Initialize services
	priceClient := bitcoin.NewBinanceClient()
	notificationService := notifications.NewService(cfg, db)

	// Initialize alert manager
	alertManager := alerts.NewAlertManager(db, notificationService)
	if err := alertManager.Start(); err != nil {
		log.Fatalf("Error starting alert manager: %v", err)
	}

	// Initialize price monitor
	priceMonitor := alerts.NewPriceMonitor(priceClient, cfg)
	priceMonitor.AddCallback(alertManager.CheckAlerts)

	// Initialize API handlers
	handler := api.NewHandler(
		adapters.NewConfigAdapter(cfg),
		adapters.NewPriceMonitorAdapter(priceMonitor),
		adapters.NewAlertManagerAdapter(alertManager),
		adapters.NewNotificationServiceAdapter(notificationService),
		db,
	)

	// Initialize router
	router := gin.Default()
	handler.SetupRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Start price monitoring
	ctx := context.Background()
	if err := priceMonitor.Start(ctx); err != nil {
		log.Fatalf("Error starting price monitor: %v", err)
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	log.Println("✅ Server shutdown complete")
}
