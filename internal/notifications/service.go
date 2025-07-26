package notifications

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/storage"

	"github.com/SherClockHolmes/webpush-go"
	"gopkg.in/gomail.v2"
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
	var errors []error

	// Enviar email
	if s.config.EnableEmailNotifications && data.Alert.EnableEmail && data.Alert.Email != "" {
		if err := s.sendEmail(data); err != nil {
			log.Printf("Error enviando email: %v", err)
			errors = append(errors, fmt.Errorf("email: %w", err))
		}
	}

	// Enviar notificación de Telegram
	if s.config.EnableTelegramNotifications && data.Alert.EnableTelegram {
		if err := s.sendTelegramNotification(data); err != nil {
			log.Printf("Error enviando notificación de Telegram: %v", err)
			errors = append(errors, fmt.Errorf("telegram: %w", err))
		}
	}

	// Enviar notificación Web Push
	if s.config.EnableWebPushNotifications && data.Alert.EnableWebPush {
		subscriptions, err := s.db.GetActiveWebPushSubscriptions()
		if err != nil {
			log.Printf("Error obteniendo subscripciones Web Push: %v", err)
			errors = append(errors, fmt.Errorf("webpush_fetch: %w", err))
		} else if len(subscriptions) > 0 {
			if err := s.SendWebPushNotification(subscriptions, data); err != nil {
				log.Printf("Error enviando notificación Web Push: %v", err)
				errors = append(errors, fmt.Errorf("webpush: %w", err))
			}
		}
	}

	// Si hay errores, retornar el primero
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func (s *Service) sendEmail(data *NotificationData) error {
	if s.config.SMTPUsername == "" || s.config.SMTPPassword == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.config.FromEmail)
	m.SetHeader("To", data.Alert.Email)
	m.SetHeader("Subject", data.Title)

	// Crear contenido HTML del email
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { background-color: #f7931a; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; margin: -20px -20px 20px -20px; }
        .price { font-size: 2em; font-weight: bold; color: #f7931a; text-align: center; margin: 20px 0; }
        .message { font-size: 1.1em; line-height: 1.6; margin: 20px 0; }
        .alert-info { background-color: #f8f9fa; padding: 15px; border-left: 4px solid #f7931a; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚨 Bitcoin Price Alert</h1>
        </div>
        
        <div class="price">$%.2f USD</div>
        
        <div class="message">%s</div>
        
        <div class="alert-info">
            <strong>Alert:</strong> %s<br>
            <strong>Triggered:</strong> %s
        </div>
        
        <div class="footer">
            <p>Esta alerta fue generada por BTC Price Alert</p>
            <p>Para desactivar las alertas, accede a tu panel de control.</p>
        </div>
    </div>
</body>
</html>
	`, data.Title, data.Price, data.Message, data.Alert.GetDescription(), data.Alert.Name)

	m.SetBody("text/html", htmlBody)

	// Configurar SMTP
	d := gomail.NewDialer(s.config.SMTPHost, s.config.SMTPPort, s.config.SMTPUsername, s.config.SMTPPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}

// Web Push Notifications (implementación completa)
func (s *Service) SendWebPushNotification(subscriptions []storage.WebPushSubscription, data *NotificationData) error {
	if !s.config.EnableWebPushNotifications {
		return nil
	}

	if s.config.VAPIDPublicKey == "" || s.config.VAPIDPrivateKey == "" {
		return fmt.Errorf("VAPID keys not configured")
	}

	var errors []error

	// Preparar el payload de la notificación
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
			{
				"action": "view",
				"title":  "Ver Dashboard",
			},
			{
				"action": "close",
				"title":  "Cerrar",
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	for _, subscription := range subscriptions {
		if !subscription.IsActive {
			continue
		}

		err := s.sendWebPushToSubscription(subscription, payloadBytes)
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

func (s *Service) sendWebPushToSubscription(subscription storage.WebPushSubscription, payload []byte) error {
	// Crear la subscripción
	sub := &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.P256dh,
			Auth:   subscription.Auth,
		},
	}

	// Enviar la notificación usando la función global
	resp, err := webpush.SendNotification(payload, sub, &webpush.Options{
		Subscriber:      s.config.VAPIDSubject,
		VAPIDPublicKey:  s.config.VAPIDPublicKey,
		VAPIDPrivateKey: s.config.VAPIDPrivateKey,
		TTL:             3600, // 1 hora
	})
	if err != nil {
		return fmt.Errorf("error sending web push: %w", err)
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("web push server returned status: %d", resp.StatusCode)
	}

	return nil
}

// Telegram Notifications
func (s *Service) sendTelegramNotification(data *NotificationData) error {
	if s.config.TelegramBotToken == "" || s.config.TelegramChatID == "" {
		return fmt.Errorf("telegram bot token o chat ID no configurados")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.config.TelegramBotToken)

	// Crear mensaje con formato HTML
	message := fmt.Sprintf(
		"🚨 <b>BITCOIN ALERT</b> 🚨\n\n"+
			"💰 <b>Precio:</b> $%.2f\n"+
			"📊 <b>Condición:</b> %s\n"+
			"⏰ <b>Hora:</b> %s\n\n"+
			"🤖 <i>Enviado por BTC Price Alert</i>",
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
