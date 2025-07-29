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

const (
	whatsappAPIVersion = "v17.0"
)

var whatsappAPIBaseURL = "https://graph.facebook.com"

// WhatsAppStrategy implements WhatsApp notifications using Meta's WhatsApp Business API
type WhatsAppStrategy struct {
	config *config.Config
	client *http.Client
}

// NewWhatsAppStrategy creates a new WhatsApp notification strategy
func NewWhatsAppStrategy(cfg *config.Config) *WhatsAppStrategy {
	return &WhatsAppStrategy{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends a WhatsApp notification using Meta's WhatsApp Business API
func (w *WhatsAppStrategy) Send(data *NotificationData) error {
	if !w.config.EnableWhatsAppNotifications {
		return nil
	}

	if w.config.WhatsAppAccessToken == "" || w.config.WhatsAppPhoneNumberID == "" {
		return fmt.Errorf("WhatsApp API configuration missing")
	}

	// Get template name based on alert language
	templateName := w.config.WhatsAppTemplateNameES
	if data.Alert.Language == "en" {
		templateName = w.config.WhatsAppTemplateNameEN
	}

	// Prepare message payload
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                data.Alert.WhatsAppNumber,
		"type":              "template",
		"template": map[string]interface{}{
			"name": templateName,
			"language": map[string]string{
				"code": data.Alert.Language,
			},
			"components": []map[string]interface{}{
				{
					"type": "body",
					"parameters": []map[string]interface{}{
						{"type": "text", "text": data.Alert.Name},
						{"type": "text", "text": fmt.Sprintf("$%.2f", data.Price)},
						{"type": "text", "text": data.Alert.GetDescription()},
						{"type": "text", "text": time.Now().Format("15:04:05 02/01/2006")},
					},
				},
			},
		},
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling WhatsApp payload: %w", err)
	}

	// Prepare request
	url := fmt.Sprintf("%s/%s/%s/messages",
		whatsappAPIBaseURL,
		whatsappAPIVersion,
		w.config.WhatsAppPhoneNumberID,
	)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating WhatsApp request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+w.config.WhatsAppAccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending WhatsApp request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return fmt.Errorf("WhatsApp API error: status %d", resp.StatusCode)
		}
		return fmt.Errorf("WhatsApp API error: %v", errorResponse)
	}

	log.Printf("📱 WhatsApp notification sent successfully to %s", data.Alert.WhatsAppNumber)
	return nil
}

// IsEnabled checks if WhatsApp notifications are enabled for this alert
func (w *WhatsAppStrategy) IsEnabled(alert *storage.Alert) bool {
	return w.config.EnableWhatsAppNotifications &&
		alert.EnableWhatsApp &&
		alert.WhatsAppNumber != ""
}

// GetChannelName returns the channel name for this strategy
func (w *WhatsAppStrategy) GetChannelName() string {
	return "whatsapp"
}
