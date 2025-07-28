package notifications

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/storage"

	"github.com/SherClockHolmes/webpush-go"
)

type WebPushStrategy struct {
	config *config.Config
	db     *storage.Database
}

func NewWebPushStrategy(cfg *config.Config, db *storage.Database) *WebPushStrategy {
	return &WebPushStrategy{
		config: cfg,
		db:     db,
	}
}

func (w *WebPushStrategy) Send(data *NotificationData) error {
	if !w.config.EnableWebPushNotifications {
		return nil
	}
	if w.config.VAPIDPublicKey == "" || w.config.VAPIDPrivateKey == "" {
		return fmt.Errorf("VAPID keys not configured")
	}

	subscriptions, err := w.db.GetActiveWebPushSubscriptions()
	if err != nil {
		return fmt.Errorf("webpush_fetch: %w", err)
	}
	if len(subscriptions) == 0 {
		return nil
	}

	payload := map[string]interface{}{
		"title": data.Title,
		"body":  fmt.Sprintf("%s\nPrecio actual: $%.2f", data.Message, data.Price),
		"icon":  "/static/images/bitcoin-icon.png",
		"badge": "/static/images/bitcoin-badge.png",
		"data": map[string]interface{}{
			"price":     data.Price,
			"alertID":   data.Alert.ID,
			"alertName": data.Alert.Name,
			"timestamp": time.Now().Unix(),
		},
		"actions": []map[string]string{
			{"action": "view", "title": "Ver Dashboard"},
			{"action": "close", "title": "Cerrar"},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}
	log.Printf("WebPush payload: %s", string(payloadBytes))

	var errors []error
	for _, subscription := range subscriptions {
		if !subscription.IsActive {
			continue
		}
		err := w.sendWebPushToSubscription(subscription, payloadBytes)
		if err != nil {
			log.Printf("Error enviando Web Push a %s: %v", subscription.Endpoint, err)
			errors = append(errors, err)
		} else {
			log.Printf("📨 Web Push enviado exitosamente a: %s", subscription.Endpoint)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("failed to send to %d subscriptions", len(errors))
	}
	return nil
}

func (w *WebPushStrategy) sendWebPushToSubscription(subscription storage.WebPushSubscription, payload []byte) error {
	sub := &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.P256dh,
			Auth:   subscription.Auth,
		},
	}
	resp, err := webpush.SendNotification(payload, sub, &webpush.Options{
		Subscriber:      w.config.VAPIDSubject,
		VAPIDPublicKey:  w.config.VAPIDPublicKey,
		VAPIDPrivateKey: w.config.VAPIDPrivateKey,
		TTL:             3600,
	})
	if err != nil {
		return fmt.Errorf("error sending web push: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("web push server returned status: %d", resp.StatusCode)
	}
	return nil
}

func (w *WebPushStrategy) IsEnabled(alert *storage.Alert) bool {
	return w.config.EnableWebPushNotifications && alert.EnableWebPush
}

func (w *WebPushStrategy) GetChannelName() string {
	return "webpush"
}
