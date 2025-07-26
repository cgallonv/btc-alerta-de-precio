package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/errors"
	"btc-alerta-de-precio/internal/storage"
)

// TelegramStrategy implements Telegram notifications
type TelegramStrategy struct {
	config *config.Config
}

// NewTelegramStrategy creates a new Telegram notification strategy
func NewTelegramStrategy(cfg *config.Config) *TelegramStrategy {
	return &TelegramStrategy{
		config: cfg,
	}
}

// Send sends a Telegram notification
func (t *TelegramStrategy) Send(data *NotificationData) error {
	if t.config.TelegramBotToken == "" || t.config.TelegramChatID == "" {
		return errors.NewAppError("TELEGRAM_CONFIG_MISSING", "Telegram bot token or chat ID not configured")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.config.TelegramBotToken)

	// Create message with HTML formatting
	message := fmt.Sprintf(
		"🚨 <b>BITCOIN ALERT</b> 🚨\n\n"+
			"💰 <b>Price:</b> $%.2f\n"+
			"📊 <b>Condition:</b> %s\n"+
			"⏰ <b>Time:</b> %s\n\n"+
			"🤖 <i>Sent by BTC Price Alert</i>",
		data.Price,
		data.Alert.GetDescription(),
		time.Now().Format("15:04:05 02/01/2006"),
	)

	payload := map[string]interface{}{
		"chat_id":    t.config.TelegramChatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return errors.WrapError(err, "TELEGRAM_MARSHAL_ERROR", "Failed to marshal Telegram payload")
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.WrapError(err, "TELEGRAM_SEND_ERROR", "Failed to send Telegram request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.NewAppError("TELEGRAM_API_ERROR", fmt.Sprintf("Telegram API returned status %d", resp.StatusCode))
	}

	return nil
}

// IsEnabled checks if Telegram notifications are enabled for this alert
func (t *TelegramStrategy) IsEnabled(alert *storage.Alert) bool {
	return t.config.EnableTelegramNotifications && alert.EnableTelegram
}

// GetChannelName returns the channel name for this strategy
func (t *TelegramStrategy) GetChannelName() string {
	return "telegram"
}

// TestSend sends a test Telegram notification
func (t *TelegramStrategy) TestSend() error {
	testData := &NotificationData{
		Title:   "🧪 Test Telegram - BTC Price Alert",
		Message: "This is a test Telegram notification",
		Price:   50000.00,
		Alert: &storage.Alert{
			Type:        "above",
			TargetPrice: 49000,
			IsActive:    true,
		},
	}

	return t.Send(testData)
}
