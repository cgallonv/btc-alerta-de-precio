package notifications

import (
	"crypto/tls"
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"btc-alerta-de-precio/config"
	"btc-alerta-de-precio/internal/storage"

	"gopkg.in/gomail.v2"
)

type Service struct {
	config *config.Config
}

type NotificationData struct {
	Title   string
	Message string
	Price   float64
	Alert   *storage.Alert
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
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

	// Enviar notificación de escritorio
	if s.config.EnableDesktopNotifications && data.Alert.EnableDesktop {
		if err := s.sendDesktopNotification(data); err != nil {
			log.Printf("Error enviando notificación de escritorio: %v", err)
			errors = append(errors, fmt.Errorf("desktop: %w", err))
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

func (s *Service) sendDesktopNotification(data *NotificationData) error {
	title := data.Title
	message := fmt.Sprintf("%s\nPrecio actual: $%.2f", data.Message, data.Price)

	switch runtime.GOOS {
	case "darwin": // macOS
		return s.sendMacOSNotification(title, message)
	case "linux":
		return s.sendLinuxNotification(title, message)
	case "windows":
		return s.sendWindowsNotification(title, message)
	default:
		return fmt.Errorf("desktop notifications not supported on %s", runtime.GOOS)
	}
}

func (s *Service) sendMacOSNotification(title, message string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s" sound name "Glass"`, message, title)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func (s *Service) sendLinuxNotification(title, message string) error {
	// Intentar con notify-send (más común)
	cmd := exec.Command("notify-send", title, message, "-i", "dialog-information")
	if err := cmd.Run(); err != nil {
		// Si falla, intentar con zenity
		cmd = exec.Command("zenity", "--info", "--title="+title, "--text="+message)
		return cmd.Run()
	}
	return nil
}

func (s *Service) sendWindowsNotification(title, message string) error {
	// Usar PowerShell para mostrar notificación en Windows
	script := fmt.Sprintf(`
		Add-Type -AssemblyName System.Windows.Forms
		$notification = New-Object System.Windows.Forms.NotifyIcon
		$notification.Icon = [System.Drawing.SystemIcons]::Information
		$notification.BalloonTipTitle = "%s"
		$notification.BalloonTipText = "%s"
		$notification.Visible = $true
		$notification.ShowBalloonTip(5000)
	`, title, message)

	cmd := exec.Command("powershell", "-Command", script)
	return cmd.Run()
}

// Método para testing
func (s *Service) TestNotifications() error {
	testData := &NotificationData{
		Title:   "🧪 Test de Notificación",
		Message: "Esta es una notificación de prueba del sistema de alertas de Bitcoin.",
		Price:   50000.00,
		Alert: &storage.Alert{
			Name:          "Test Alert",
			Email:         s.config.FromEmail,
			EnableEmail:   true,
			EnableDesktop: true,
		},
	}

	log.Println("Enviando notificación de prueba...")
	return s.SendAlert(testData)
}

// Web Push Notifications (implementación básica)
func (s *Service) SendWebPushNotification(subscriptions []storage.WebPushSubscription, data *NotificationData) error {
	// Nota: Para una implementación completa de Web Push, necesitarías usar una librería
	// como github.com/SherClockHolmes/webpush-go
	// Por simplicidad, aquí se proporciona la estructura base

	log.Printf("Web Push notifications no implementadas completamente. Datos: %+v", data)
	log.Printf("Subscriptions: %d", len(subscriptions))

	// TODO: Implementar envío real de web push notifications
	// Requiere VAPID keys y manejo de subscriptions de service workers

	return nil
}
