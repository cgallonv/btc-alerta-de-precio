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
	"btc-alerta-de-precio/internal/alerts"
	"btc-alerta-de-precio/internal/api"
	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	// Inicializar base de datos
	db, err := storage.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	// Inicializar servicios
	bitcoinClient := bitcoin.NewClient()
	notificationService := notifications.NewService(cfg, db)
	alertService := alerts.NewService(db, bitcoinClient, notificationService, cfg)

	// Iniciar el monitoreo de precios
	go alertService.StartMonitoring(cfg.CheckInterval)

	// Iniciar el monitoreo de porcentaje de cambio
	go alertService.StartPercentageMonitoring()

	// Configurar y iniciar servidor web
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	apiHandler := api.NewHandler(alertService)
	apiHandler.SetupRoutes(router)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Iniciar servidor en goroutine
	go func() {
		log.Printf("Servidor iniciado en puerto %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error iniciando servidor: %v", err)
		}
	}()

	// Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Cerrando aplicación...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error cerrando servidor: %v", err)
	}

	alertService.Stop()
	alertService.StopPercentageMonitoring()
	log.Println("Aplicación cerrada correctamente")
}
