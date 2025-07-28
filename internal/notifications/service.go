package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/storage"
)

type Service struct {
	config *config.Config
	db     *storage.Database
}

type NotificationData struct {
	Title   string
	Message string
	Price   float64
	Alert   *storage.Alert
}

func NewService(cfg *config.Config, db *storage.Database) *Service {
	return &Service{
		config: cfg,
		db:     db,
	}
}

func (s *Service) SendAlert(data *NotificationData) error {
	// Create strategies
	strategies := []NotificationStrategy{
		NewEmailStrategy(s.config),
		NewTelegramStrategy(s.config),
		NewWebPushStrategy(s.config, s.db),
	}
	manager := NewNotificationManager(strategies...)
	return manager.SendAlert(data)
}

// Telegram Notifications
func (s *Service) sendTelegramNotification(data *NotificationData) error {
	if s.config.TelegramBotToken == "" || s.config.TelegramChatID == "" {
		return fmt.Errorf("telegram bot token o chat ID no configurados")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.config.TelegramBotToken)

	// Crear mensaje con formato HTML
	message := fmt.Sprintf(
		"🚨 <b>BITCOIN ALERT - %s</b> 🚨\n\n"+
			"💰 <b>Precio:</b> $%.2f\n"+
			"📊 <b>Condición:</b> %s\n"+
			"⏰ <b>Hora:</b> %s\n\n"+
			"🤖 <i>Enviado por BTC Price Alert</i>",
		data.Alert.Name,
		data.Price,
		data.Alert.GetDescription(),
		time.Now().Format("15:04:05 02/01/2006"),
	)

	payload := map[string]interface{}{
		"chat_id":    s.config.TelegramChatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error enviando request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API error: status %d", resp.StatusCode)
	}

	log.Println("📱 Notificación de Telegram enviada exitosamente")
	return nil
}

// Test de notificación de Telegram
func (s *Service) TestTelegramNotification() error {
	if s.config.TelegramBotToken == "" || s.config.TelegramChatID == "" {
		return fmt.Errorf("telegram no configurado - revisa TELEGRAM_BOT_TOKEN y TELEGRAM_CHAT_ID en .env")
	}

	testData := &NotificationData{
		Title:   "🧪 Test de Telegram - BTC Price Alert",
		Message: "Esta es una prueba de notificación de Telegram",
		Price:   50000.00,
		Alert: &storage.Alert{
			Type:        "above",
			TargetPrice: 49000,
			IsActive:    true,
		},
	}

	log.Println("📱 Enviando notificación de prueba a Telegram...")
	return s.sendTelegramNotification(testData)
}
