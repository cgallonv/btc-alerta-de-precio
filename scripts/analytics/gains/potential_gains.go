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
	ID                  uint      `gorm:"primaryKey"`
	CloseTime           time.Time `gorm:"column:close_time"`
	LastPrice           float64   `gorm:"column:last_price"`
	MaxDiscount         *float64  `gorm:"column:max_discount"`
	TimeTo2PercentGains *float64  `gorm:"column:time_to_2_percent_gains"` // Tiempo en horas
	TimeTo3PercentGains *float64  `gorm:"column:time_to_3_percent_gains"` // Tiempo en horas
	TimeTo4PercentGains *float64  `gorm:"column:time_to_4_percent_gains"` // Tiempo en horas
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
		var opportunities []TickerAnalysis
		result := db.DB().
			Where("id > ? AND max_discount IS NOT NULL", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&opportunities)

		if result.Error != nil {
			log.Fatalf("❌ Error leyendo oportunidades: %v", result.Error)
		}

		if len(opportunities) == 0 {
			break // No hay más registros para procesar
		}

		log.Printf("🔄 Procesando batch desde ID %d", lastID)

		oportunidadesAnalizadas := 0
		// Procesar cada oportunidad del lote
		for _, opportunity := range opportunities {
			if err := analyzeGainsPotential(db.DB(), &opportunity); err != nil {
				log.Printf("⚠️ Error analizando ganancias para ID %d: %v", opportunity.ID, err)
				continue
			}
			oportunidadesAnalizadas++
		}

		lastID = opportunities[len(opportunities)-1].ID
		log.Printf("✅ Batch completado: %d oportunidades analizadas", oportunidadesAnalizadas)
	}

	log.Println("✅ Proceso completado exitosamente!")
}

func analyzeGainsPotential(db *gorm.DB, opportunity *TickerAnalysis) error {
	// Definir ventana máxima de búsqueda (2 semanas)
	maxSearchWindow := opportunity.CloseTime.Add(14 * 24 * time.Hour)

	// Buscar todos los tiempos de ganancia en un solo query
	time2p, time3p, time4p := findAllGainThresholds(db, opportunity.CloseTime, maxSearchWindow, opportunity.LastPrice)

	// Debug: Mostrar tiempos calculados
	log.Printf("⏱️ Tiempos de ganancia calculados:")
	if time2p != nil {
		log.Printf("   - 2%%: %.2f horas", *time2p)
	}
	if time3p != nil {
		log.Printf("   - 3%%: %.2f horas", *time3p)
	}
	if time4p != nil {
		log.Printf("   - 4%%: %.2f horas", *time4p)
	}

	// Actualizar los tiempos encontrados
	err := db.Model(opportunity).Updates(map[string]interface{}{
		"time_to_2_percent_gains": time2p,
		"time_to_3_percent_gains": time3p,
		"time_to_4_percent_gains": time4p,
	}).Error

	if err != nil {
		log.Printf("❌ Error actualizando tiempos: %v", err)
	} else if time2p != nil || time3p != nil || time4p != nil {
		log.Printf("✅ Tiempos actualizados para ID %d", opportunity.ID)
	}

	return err
}

func findAllGainThresholds(db *gorm.DB, baseTime, maxTime time.Time, basePrice float64) (time2p, time3p, time4p *float64) {
	// Calcular precios objetivo para cada porcentaje
	price2p := basePrice * (1 + 1.99/100)
	price3p := basePrice * (1 + 2.99/100)
	price4p := basePrice * (1 + 3.99/100)

	// Query optimizado que busca los tres umbrales en una sola consulta
	query := `
		WITH FirstGains AS (
			SELECT 
				MIN(CASE WHEN last_price >= ? THEN close_time END) as time_2p,
				MIN(CASE WHEN last_price >= ? THEN close_time END) as time_3p,
				MIN(CASE WHEN last_price >= ? THEN close_time END) as time_4p
			FROM ticker_data 
			WHERE close_time > ? AND close_time <= ?
		)
		SELECT time_2p, time_3p, time_4p FROM FirstGains`

	// Debug: Mostrar query con valores
	debugQuery := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Raw(query, price2p, price3p, price4p, baseTime, maxTime)
	})
	log.Printf("🔍 Query ejecutado:\n%s", debugQuery)

	var result struct {
		Time2p *time.Time
		Time3p *time.Time
		Time4p *time.Time
	}

	// Debug: Mostrar precios objetivo y ventana de tiempo
	log.Printf("🎯 Buscando ganancias para precio base %.2f:", basePrice)
	log.Printf("   - Objetivo 2%%: %.2f", price2p)
	log.Printf("   - Objetivo 3%%: %.2f", price3p)
	log.Printf("   - Objetivo 4%%: %.2f", price4p)
	log.Printf("📅 Ventana de búsqueda: %s → %s", 
		baseTime.Format("2006-01-02 15:04:05"),
		maxTime.Format("2006-01-02 15:04:05"))

	// Ejecutar el query con todos los parámetros
	err := db.Raw(query, price2p, price3p, price4p, baseTime, maxTime).Scan(&result).Error
	if err != nil {
		log.Printf("⚠️ Error buscando ganancias: %v", err)
		return nil, nil, nil
	}

	// Debug: Mostrar resultados encontrados
	log.Printf("🔍 Resultados encontrados:")
	if result.Time2p != nil {
		log.Printf("   - 2%%: %s", result.Time2p.Format("2006-01-02 15:04:05"))
	}
	if result.Time3p != nil {
		log.Printf("   - 3%%: %s", result.Time3p.Format("2006-01-02 15:04:05"))
	}
	if result.Time4p != nil {
		log.Printf("   - 4%%: %s", result.Time4p.Format("2006-01-02 15:04:05"))
	}

	// Convertir los resultados a horas
	if result.Time2p != nil {
		hours := result.Time2p.Sub(baseTime).Hours()
		time2p = &hours
	}
	if result.Time3p != nil {
		hours := result.Time3p.Sub(baseTime).Hours()
		time3p = &hours
	}
	if result.Time4p != nil {
		hours := result.Time4p.Sub(baseTime).Hours()
		time4p = &hours
	}

	return
}
