package main

import (
	"log"
	"time"

	"github.com/cgallonv/btc-alerta-de-precio/config"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage"
	"gorm.io/gorm"
)

// TickerAnalysis representa un registro de la tabla ticker_data con los campos necesarios
type TickerAnalysis struct {
	ID          uint      `gorm:"primaryKey"`
	CloseTime   time.Time `gorm:"column:close_time"`
	LastPrice   float64   `gorm:"column:last_price"`
	MaxDiscount *float64  `gorm:"column:max_discount"` // Permitimos NULL usando puntero para descuentos < -1%
}

func (TickerAnalysis) TableName() string {
	return "ticker_data"
}

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Error cargando configuración: %v", err)
	}

	// Conectar a la base de datos
	db, err := storage.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("❌ Error conectando a la base de datos: %v", err)
	}
	defer db.Close()

	// Procesar registros en lotes para evitar sobrecarga de memoria
	const batchSize = 1000
	var lastID uint = 0

	for {
		var records []TickerAnalysis
		result := db.DB().
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&records)

		if result.Error != nil {
			log.Fatalf("❌ Error leyendo registros: %v", result.Error)
		}

		if len(records) == 0 {
			break // No hay más registros para procesar
		}

		log.Printf("🔄 Procesando batch desde ID %d", lastID)

		descuentosEncontrados := 0
		// Procesar cada registro del lote
		for _, record := range records {
			discount := calculateMaxDiscount(db.DB(), record)
			if discount != nil {
				descuentosEncontrados++
				// Solo actualizamos si encontramos un descuento válido (negativo)
				if err := updateMaxDiscount(db.DB(), record.ID, *discount); err != nil {
					log.Printf("⚠️ Error actualizando descuento para ID %d: %v", record.ID, err)
				}
			}
		}

		lastID = records[len(records)-1].ID
		log.Printf("✅ Batch completado: %d registros procesados, %d descuentos encontrados", len(records), descuentosEncontrados)
	}

	log.Println("✅ Proceso completado exitosamente!")
}

func calculateMaxDiscount(db *gorm.DB, record TickerAnalysis) *float64 {
	// Calcular descuento para ventana de 24 horas
	discount24h := calculate24hDiscount(db, record)

	// Calcular descuento para ventana de 5 horas
	discount5h := calculate5hDiscount(db, record)

	// Si no hay descuentos negativos, retornamos nil
	if discount24h == nil && discount5h == nil {
		return nil
	}

	// Comparar los descuentos y retornar el más negativo
	if discount24h == nil {
		return discount5h
	}
	if discount5h == nil {
		return discount24h
	}
	if *discount24h < *discount5h {
		return discount24h
	}
	return discount5h
}

func calculate24hDiscount(db *gorm.DB, record TickerAnalysis) *float64 {
	var prevRecord TickerAnalysis

	// Calcular exactamente 24 horas antes para comparación día a día
	exactTimeWindow := record.CloseTime.Add(-24 * time.Hour)

	// Buscar el registro más cercano a exactamente 24 horas antes
	err := db.Where("close_time <= ?", exactTimeWindow).
		Order(db.Raw("ABS(STRFTIME('%s', close_time) - STRFTIME('%s', ?))", exactTimeWindow)).
		First(&prevRecord).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {

			return nil
		}
		log.Printf("⚠️ Error buscando registro 24h anterior: %v", err)
		return nil
	}

	// Calcular el descuento porcentual comparando con el precio de hace 24h
	discount := calculatePercentageDiscount(record.LastPrice, prevRecord.LastPrice)

	if discount >= -1.0 {

		return nil
	}

	// Convertir a float con 2 decimales para almacenamiento
	discountFloat := float64(int(discount*100)) / 100.0

	return &discountFloat
}

func calculate5hDiscount(db *gorm.DB, record TickerAnalysis) *float64 {
	var highestPrice5h float64

	// Buscar el precio más alto en las últimas 5 horas para detectar caídas
	timeWindow := record.CloseTime.Add(-5 * time.Hour)

	err := db.Model(&TickerAnalysis{}).
		Where("close_time BETWEEN ? AND ?", timeWindow, record.CloseTime).
		Select("MAX(last_price)"). // Buscamos el máximo para comparar contra caídas
		Row().
		Scan(&highestPrice5h)

	if err != nil {
		log.Printf("⚠️ Error buscando precio más alto en 5h: %v", err)
		return nil
	}

	if highestPrice5h == 0 {

		return nil
	}

	// Si el precio actual es menor que el máximo, tenemos un descuento
	discount := calculatePercentageDiscount(record.LastPrice, highestPrice5h)

	if discount >= -1.0 {

		return nil
	}

	// Convertir a float con 2 decimales
	discountFloat := float64(int(discount*100)) / 100.0

	return &discountFloat
}

func calculatePercentageDiscount(currentPrice, previousPrice float64) float64 {
	if previousPrice == 0 {
		return 0
	}
	return ((currentPrice - previousPrice) / previousPrice) * 100
}

func updateMaxDiscount(db *gorm.DB, id uint, discount float64) error {
	return db.Model(&TickerAnalysis{}).
		Where("id = ?", id).
		Update("max_discount", discount).
		Error
}
