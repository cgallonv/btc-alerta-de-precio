package config

import (
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

	// Monitoreo
	CheckInterval time.Duration

	// Email
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string

	// Notificaciones
	EnableDesktopNotifications  bool
	EnableEmailNotifications    bool
	EnableWebPushNotifications  bool
	EnableTelegramNotifications bool

	// Telegram Bot
	TelegramBotToken string
	TelegramChatID   string

	// Web Push (para notificaciones de Chrome)
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
}

func Load() (*Config, error) {
	// Cargar .env si existe
	godotenv.Load()

	checkInterval, _ := time.ParseDuration(getEnv("CHECK_INTERVAL", "30s"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		Port:          getEnv("PORT", "8080"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		DatabasePath:  getEnv("DATABASE_PATH", "alerts.db"),
		BitcoinAPIURL: getEnv("BITCOIN_API_URL", "https://api.coindesk.com/v1/bpi/currentprice.json"),
		CheckInterval: checkInterval,

		// Email configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     smtpPort,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", ""),

		// Notification settings
		EnableDesktopNotifications:  getEnvBool("ENABLE_DESKTOP_NOTIFICATIONS", true),
		EnableEmailNotifications:    getEnvBool("ENABLE_EMAIL_NOTIFICATIONS", true),
		EnableWebPushNotifications:  getEnvBool("ENABLE_WEB_PUSH_NOTIFICATIONS", true),
		EnableTelegramNotifications: getEnvBool("ENABLE_TELEGRAM_NOTIFICATIONS", false),

		// Telegram Bot configuration
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:   getEnv("TELEGRAM_CHAT_ID", ""),

		// Web Push (VAPID keys)
		VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:    getEnv("VAPID_SUBJECT", "mailto:admin@btcalerts.com"),
	}, nil
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
