package migrations

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/cgallonv/btc-alerta-de-precio/internal/storage/models"
)

// MigrateTickerData creates or updates the ticker_data table schema.
// It also creates necessary indexes for efficient querying.
//
// Example usage:
//
//	if err := migrations.MigrateTickerData(db); err != nil {
//	    log.Fatalf("Failed to migrate ticker data: %v", err)
//	}
func MigrateTickerData(db *gorm.DB) error {
	// Create or update table schema
	if err := db.AutoMigrate(&models.TickerData{}); err != nil {
		return fmt.Errorf("failed to migrate ticker_data table: %w", err)
	}

	// Create indexes
	for _, idx := range (models.TickerData{}).Indexes() {
		indexName := fmt.Sprintf("idx_%s_%s", "ticker_data", idx[0])
		if err := db.Exec(fmt.Sprintf(
			"CREATE INDEX IF NOT EXISTS %s ON ticker_data (%s)",
			indexName,
			idx[0],
		)).Error; err != nil {
			return fmt.Errorf("failed to create index %s: %w", indexName, err)
		}
	}

	return nil
}
