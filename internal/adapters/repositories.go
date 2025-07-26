package adapters

import (
	"btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/storage"
	"time"
)

// GormAlertRepository adapts storage.Database to implement AlertRepository interface
type GormAlertRepository struct {
	db *storage.Database
}

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

// GormPriceRepository adapts storage.Database to implement PriceRepository interface
type GormPriceRepository struct {
	db *storage.Database
}

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

func (r *GormPriceRepository) GetPriceHistoryByDateRange(start, end time.Time) ([]storage.PriceHistory, error) {
	prices, err := r.db.GetPriceHistoryByDateRange(start, end)
	if err != nil {
		return nil, errors.WrapError(err, "DATABASE_GET_PRICE_HISTORY_RANGE", "Failed to get price history by date range").
			WithField("start", start).WithField("end", end)
	}
	return prices, nil
}

func (r *GormPriceRepository) CleanOldPriceHistory(days int) error {
	if err := r.db.CleanOldPriceHistory(days); err != nil {
		return errors.WrapError(err, "DATABASE_CLEAN_PRICE_HISTORY", "Failed to clean old price history").WithField("days", days)
	}
	return nil
}

// GormNotificationRepository adapts storage.Database to implement NotificationRepository interface
type GormNotificationRepository struct {
	db *storage.Database
}

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

// GormWebPushRepository adapts storage.Database to implement WebPushRepository interface
type GormWebPushRepository struct {
	db *storage.Database
}

func NewGormWebPushRepository(db *storage.Database) interfaces.WebPushRepository {
	return &GormWebPushRepository{db: db}
}

func (r *GormWebPushRepository) SaveWebPushSubscription(sub *storage.WebPushSubscription) error {
	if err := r.db.SaveWebPushSubscription(sub); err != nil {
		return errors.WrapError(err, "DATABASE_SAVE_WEBPUSH_SUBSCRIPTION", "Failed to save web push subscription")
	}
	return nil
}

func (r *GormWebPushRepository) GetActiveWebPushSubscriptions() ([]storage.WebPushSubscription, error) {
	subs, err := r.db.GetActiveWebPushSubscriptions()
	if err != nil {
		return nil, errors.WrapError(err, "DATABASE_GET_WEBPUSH_SUBSCRIPTIONS", "Failed to get active web push subscriptions")
	}
	return subs, nil
}

func (r *GormWebPushRepository) RemoveWebPushSubscription(endpoint string) error {
	if err := r.db.RemoveWebPushSubscription(endpoint); err != nil {
		return errors.WrapError(err, "DATABASE_REMOVE_WEBPUSH_SUBSCRIPTION", "Failed to remove web push subscription").
			WithField("endpoint", endpoint)
	}
	return nil
}

// GormStatsRepository adapts storage.Database to implement StatsRepository interface
type GormStatsRepository struct {
	db *storage.Database
}

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
