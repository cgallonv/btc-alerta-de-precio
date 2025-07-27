package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/storage"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Migrar el modelo por si acaso
	db.AutoMigrate(&storage.Alert{})

	alerts := []storage.Alert{
		{
			Name:        "Precio por debajo de 17000",
			Type:        "below",
			TargetPrice: 17000,
			Percentage:  0,
			IsActive:    true,
			Email:       "cgallonv@gmail.com",
			EnableEmail: true,
		},
		{
			Name:        "Precio por debajo de 16000",
			Type:        "below",
			TargetPrice: 16000,
			Percentage:  0,
			IsActive:    true,
			Email:       "cgallonv@gmail.com",
			EnableEmail: true,
		},
		{
			Name:        "Precio por debajo de 15000",
			Type:        "below",
			TargetPrice: 15000,
			Percentage:  0,
			IsActive:    true,
			Email:       "cgallonv@gmail.com",
			EnableEmail: true,
		},
		{
			Name:        "Bajó 3%",
			Type:        "change",
			TargetPrice: 0,
			Percentage:  -3,
			IsActive:    true,
			Email:       "cgallonv@gmail.com",
			EnableEmail: true,
		},
		{
			Name:        "Bajó 4%",
			Type:        "change",
			TargetPrice: 0,
			Percentage:  -4,
			IsActive:    true,
			Email:       "cgallonv@gmail.com",
			EnableEmail: true,
		},
		{
			Name:        "Bajó 5%",
			Type:        "change",
			TargetPrice: 0,
			Percentage:  -5,
			IsActive:    true,
			Email:       "cgallonv@gmail.com",
			EnableEmail: true,
		},
		{
			Name:        "PRECIO BAJÓ 6%!!!!!!",
			Type:        "change",
			TargetPrice: 0,
			Percentage:  -6,
			IsActive:    true,
			Email:       "cgallonv@gmail.com",
			EnableEmail: true,
		},
		{
			Name:        "PRECIO BAJÓ 7%!!!!!!",
			Type:        "change",
			TargetPrice: 0,
			Percentage:  -7,
			IsActive:    true,
			Email:       "cgallonv@gmail.com",
			EnableEmail: true,
		},
	}

	for i, alert := range alerts {
		alert.CreatedAt = time.Now()
		alert.UpdatedAt = time.Now()
		if err := db.Create(&alert).Error; err != nil {
			log.Printf("❌ Error inserting alert #%d (%s): %v", i+1, alert.Name, err)
		} else {
			fmt.Printf("✅ Alerta precargada: %s\n", alert.Name)
		}
	}

	fmt.Println("\n¡Precarga de alertas completada!")
	os.Exit(0)
}
