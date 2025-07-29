package adapters

import (
	"btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/storage"
)

// GormAlertRepository adapts storage.Database to implement AlertRepository interface.
//
// Example usage:
//
//	repo := NewGormAlertRepository(db)
//	err := repo.CreateAlert(alert)
type GormAlertRepository struct {
	db *storage.Database
}

// NewGormAlertRepository creates a new GormAlertRepository.
//
// Example usage:
//
//	repo := NewGormAlertRepository(db)
func NewGormAlertRepository(db *storage.Database) interfaces.AlertRepository {
	return &GormAlertRepository{db: db}
}

func (r *GormAlertRepository) CreateAlert(alert *storage.Alert) error {
	if err := r.db.CreateAlert(alert); err != nil {
		return errors.WrapError(err, "DATABASE_CREATE_ALERT", "Failed to create alert")
	}
	return nil
}

func (r *GormAlertRepository) GetAlert(id uint) (*storage.Alert, error) {
	alert, err := r.db.GetAlert(id)
	if err != nil {
		return nil, errors.WrapError(err, "DATABASE_GET_ALERT", "Failed to get alert").WithField("alert_id", id)
	}
	return alert, nil
}

func (r *GormAlertRepository) GetAlerts() ([]storage.Alert, error) {
	alerts, err := r.db.GetAlerts()
	if err != nil {
		return nil, errors.WrapError(err, "DATABASE_GET_ALERTS", "Failed to get alerts")
	}
	return alerts, nil
}

func (r *GormAlertRepository) GetActiveAlerts() ([]storage.Alert, error) {
	alerts, err := r.db.GetActiveAlerts()
	if err != nil {
		return nil, errors.WrapError(err, "DATABASE_GET_ACTIVE_ALERTS", "Failed to get active alerts")
	}
	return alerts, nil
}

func (r *GormAlertRepository) UpdateAlert(alert *storage.Alert) error {
	if err := r.db.UpdateAlert(alert); err != nil {
		return errors.WrapError(err, "DATABASE_UPDATE_ALERT", "Failed to update alert").WithField("alert_id", alert.ID)
	}
	return nil
}

func (r *GormAlertRepository) DeleteAlert(id uint) error {
	if err := r.db.DeleteAlert(id); err != nil {
		return errors.WrapError(err, "DATABASE_DELETE_ALERT", "Failed to delete alert").WithField("alert_id", id)
	}
	return nil
}

func (r *GormAlertRepository) ToggleAlert(id uint) error {
	if err := r.db.ToggleAlert(id); err != nil {
		return errors.WrapError(err, "DATABASE_TOGGLE_ALERT", "Failed to toggle alert").WithField("alert_id", id)
	}
	return nil
}

// GormPriceRepository adapts storage.Database to implement PriceRepository interface.
//
// Example usage:
//
//	repo := NewGormPriceRepository(db)
//	price, err := repo.GetLatestPrice()
type GormPriceRepository struct {
	db *storage.Database
}

// NewGormPriceRepository creates a new GormPriceRepository.
//
// Example usage:
//
//	repo := NewGormPriceRepository(db)
func NewGormPriceRepository(db *storage.Database) interfaces.PriceRepository {
	return &GormPriceRepository{db: db}
}

func (r *GormPriceRepository) SavePriceHistory(price *storage.PriceHistory) error {
	if err := r.db.SavePriceHistory(price); err != nil {
		return errors.WrapError(err, "DATABASE_SAVE_PRICE", "Failed to save price history")
	}
	return nil
}

func (r *GormPriceRepository) GetLatestPrice() (*storage.PriceHistory, error) {
	price, err := r.db.GetLatestPrice()
	if err != nil {
		return nil, errors.WrapError(err, "DATABASE_GET_LATEST_PRICE", "Failed to get latest price")
	}
	return price, nil
}

func (r *GormPriceRepository) GetPriceHistory(limit int) ([]storage.PriceHistory, error) {
	prices, err := r.db.GetPriceHistory(limit)
	if err != nil {
		return nil, errors.WrapError(err, "DATABASE_GET_PRICE_HISTORY", "Failed to get price history").WithField("limit", limit)
	}
	return prices, nil
}

// GormNotificationRepository adapts storage.Database to implement NotificationRepository interface.
//
// Example usage:
//
//	repo := NewGormNotificationRepository(db)
//	err := repo.LogNotification(log)
type GormNotificationRepository struct {
	db *storage.Database
}

// NewGormNotificationRepository creates a new GormNotificationRepository.
//
// Example usage:
//
//	repo := NewGormNotificationRepository(db)
func NewGormNotificationRepository(db *storage.Database) interfaces.NotificationRepository {
	return &GormNotificationRepository{db: db}
}

func (r *GormNotificationRepository) LogNotification(log *storage.NotificationLog) error {
	if err := r.db.LogNotification(log); err != nil {
		return errors.WrapError(err, "DATABASE_LOG_NOTIFICATION", "Failed to log notification")
	}
	return nil
}

func (r *GormNotificationRepository) GetNotificationLogs(alertID uint, limit int) ([]storage.NotificationLog, error) {
	logs, err := r.db.GetNotificationLogs(alertID, limit)
	if err != nil {
		return nil, errors.WrapError(err, "DATABASE_GET_NOTIFICATION_LOGS", "Failed to get notification logs").
			WithField("alert_id", alertID).WithField("limit", limit)
	}
	return logs, nil
}

// GormStatsRepository adapts storage.Database to implement StatsRepository interface.
//
// Example usage:
//
//	repo := NewGormStatsRepository(db)
//	stats, err := repo.GetStats()
type GormStatsRepository struct {
	db *storage.Database
}

// NewGormStatsRepository creates a new GormStatsRepository.
//
// Example usage:
//
//	repo := NewGormStatsRepository(db)
func NewGormStatsRepository(db *storage.Database) interfaces.StatsRepository {
	return &GormStatsRepository{db: db}
}

func (r *GormStatsRepository) GetStats() (map[string]interface{}, error) {
	stats, err := r.db.GetStats()
	if err != nil {
		return nil, errors.WrapError(err, "DATABASE_GET_STATS", "Failed to get stats")
	}
	return stats, nil
}
