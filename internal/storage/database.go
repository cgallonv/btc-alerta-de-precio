package storage

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configurar SQLite
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(1) // SQLite works better with single connection
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	database := &Database{db: db}

	// Migrar esquemas
	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return database, nil
}

func (d *Database) migrate() error {
	return d.db.AutoMigrate(
		&Alert{},
		&PriceHistory{},
		&NotificationLog{},
		&WebPushSubscription{},
	)
}

// Alert operations
func (d *Database) CreateAlert(alert *Alert) error {
	return d.db.Create(alert).Error
}

func (d *Database) GetAlert(id uint) (*Alert, error) {
	var alert Alert
	err := d.db.First(&alert, id).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (d *Database) GetAlerts() ([]Alert, error) {
	var alerts []Alert
	err := d.db.Find(&alerts).Error
	return alerts, err
}

func (d *Database) GetActiveAlerts() ([]Alert, error) {
	var alerts []Alert
	err := d.db.Where("is_active = ?", true).Find(&alerts).Error
	return alerts, err
}

func (d *Database) UpdateAlert(alert *Alert) error {
	return d.db.Save(alert).Error
}

func (d *Database) DeleteAlert(id uint) error {
	return d.db.Delete(&Alert{}, id).Error
}

func (d *Database) ToggleAlert(id uint) error {
	return d.db.Model(&Alert{}).Where("id = ?", id).Update("is_active", gorm.Expr("NOT is_active")).Error
}

// Price History operations
func (d *Database) SavePriceHistory(price *PriceHistory) error {
	return d.db.Create(price).Error
}

func (d *Database) GetLatestPrice() (*PriceHistory, error) {
	var price PriceHistory
	err := d.db.Order("timestamp desc").First(&price).Error
	if err != nil {
		return nil, err
	}
	return &price, nil
}

func (d *Database) GetPriceHistory(limit int) ([]PriceHistory, error) {
	var prices []PriceHistory
	err := d.db.Order("timestamp desc").Limit(limit).Find(&prices).Error
	return prices, err
}

func (d *Database) GetPriceHistoryByDateRange(start, end time.Time) ([]PriceHistory, error) {
	var prices []PriceHistory
	err := d.db.Where("timestamp BETWEEN ? AND ?", start, end).Order("timestamp desc").Find(&prices).Error
	return prices, err
}

func (d *Database) CleanOldPriceHistory(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return d.db.Where("timestamp < ?", cutoff).Delete(&PriceHistory{}).Error
}

// Notification Log operations
func (d *Database) LogNotification(log *NotificationLog) error {
	return d.db.Create(log).Error
}

func (d *Database) GetNotificationLogs(alertID uint, limit int) ([]NotificationLog, error) {
	var logs []NotificationLog
	query := d.db.Preload("Alert").Order("created_at desc")

	if alertID > 0 {
		query = query.Where("alert_id = ?", alertID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// Web Push Subscription operations
func (d *Database) SaveWebPushSubscription(sub *WebPushSubscription) error {
	// Intentar actualizar primero, si no existe, crear
	var existing WebPushSubscription
	err := d.db.Where("endpoint = ?", sub.Endpoint).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// No existe, crear nuevo
		return d.db.Create(sub).Error
	} else if err != nil {
		return err
	}

	// Existe, actualizar
	existing.P256dh = sub.P256dh
	existing.Auth = sub.Auth
	existing.UserID = sub.UserID
	existing.IsActive = sub.IsActive
	return d.db.Save(&existing).Error
}

func (d *Database) GetActiveWebPushSubscriptions() ([]WebPushSubscription, error) {
	var subs []WebPushSubscription
	err := d.db.Where("is_active = ?", true).Find(&subs).Error
	return subs, err
}

func (d *Database) RemoveWebPushSubscription(endpoint string) error {
	return d.db.Where("endpoint = ?", endpoint).Delete(&WebPushSubscription{}).Error
}

// Utility operations
func (d *Database) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Contar alertas
	var totalAlerts, activeAlerts int64
	d.db.Model(&Alert{}).Count(&totalAlerts)
	d.db.Model(&Alert{}).Where("is_active = ?", true).Count(&activeAlerts)

	// Contar notificaciones
	var totalNotifications int64
	d.db.Model(&NotificationLog{}).Count(&totalNotifications)

	// Último precio
	latestPrice, _ := d.GetLatestPrice()

	stats["total_alerts"] = totalAlerts
	stats["active_alerts"] = activeAlerts
	stats["total_notifications"] = totalNotifications

	if latestPrice != nil {
		stats["latest_price"] = latestPrice.Price
		stats["latest_price_time"] = latestPrice.Timestamp
	}

	return stats, nil
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
