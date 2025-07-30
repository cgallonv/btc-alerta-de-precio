package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Servidor
	Port        string
	Environment string

	// Base de datos
	DatabasePath string

	// Bitcoin API
	BitcoinAPIURL string

	// Monitoreo - Intervalo único para todo (precio, porcentaje, backend y frontend)
	CheckInterval time.Duration

	// Email
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string

	// Notificaciones
	EnableEmailNotifications bool

	EnableTelegramNotifications bool
	EnableWhatsAppNotifications bool

	// Telegram Bot
	TelegramBotToken string
	TelegramChatID   string

	// Web Push (para notificaciones de Chrome)
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string

	// WhatsApp Business API (Meta)
	WhatsAppAccessToken    string
	WhatsAppPhoneNumberID  string
	WhatsAppBusinessAccID  string
	WhatsAppTemplateNameES string
	WhatsAppTemplateNameEN string

	// Binance API
	BinanceAPIKey    string
	BinanceAPISecret string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	} else {
		log.Printf(".env file loaded successfully")
	}

	checkInterval, _ := time.ParseDuration(getEnv("CHECK_INTERVAL", "30s"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	// Load Binance API credentials
	binanceKey := getEnv("BINANCE_API_KEY", "")
	binanceSecret := getEnv("BINANCE_API_SECRET", "")

	// Log Binance API configuration status
	if binanceKey == "" || binanceSecret == "" {
		log.Printf("⚠️ Binance API credentials not configured")
	} else {
		log.Printf("✅ Binance API credentials loaded (key length: %d, secret length: %d)", len(binanceKey), len(binanceSecret))
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		DatabasePath:  getEnv("DATABASE_PATH", "btc_market_data.db"),
		BitcoinAPIURL: getEnv("BITCOIN_API_URL", "https://api.coindesk.com/v1/bpi/currentprice.json"),
		CheckInterval: checkInterval,

		// Email configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     smtpPort,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", ""),

		// Notification settings
		EnableEmailNotifications: getEnvBool("ENABLE_EMAIL_NOTIFICATIONS", true),

		EnableTelegramNotifications: getEnvBool("ENABLE_TELEGRAM_NOTIFICATIONS", false),
		EnableWhatsAppNotifications: getEnvBool("ENABLE_WHATSAPP_NOTIFICATIONS", false),

		// Telegram Bot configuration
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:   getEnv("TELEGRAM_CHAT_ID", ""),

		// Web Push (VAPID keys)
		VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:    getEnv("VAPID_SUBJECT", "mailto:admin@btcalerts.com"),

		// WhatsApp Business API configuration
		WhatsAppAccessToken:    getEnv("WHATSAPP_ACCESS_TOKEN", ""),
		WhatsAppPhoneNumberID:  getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppBusinessAccID:  getEnv("WHATSAPP_BUSINESS_ACCOUNT_ID", ""),
		WhatsAppTemplateNameES: getEnv("WHATSAPP_TEMPLATE_NAME_ES", "btc_alert_es"),
		WhatsAppTemplateNameEN: getEnv("WHATSAPP_TEMPLATE_NAME_EN", "btc_alert_en"),

		// Binance API configuration
		BinanceAPIKey:    binanceKey,
		BinanceAPISecret: binanceSecret,
	}, nil
}

func (c *Config) GetString(key string) string {
	switch key {
	case "binance.api_key":
		return c.BinanceAPIKey
	case "binance.api_secret":
		return c.BinanceAPISecret
	default:
		return ""
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		result, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return result
	}
	return defaultValue
}
