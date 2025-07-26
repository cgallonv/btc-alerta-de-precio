package alerts

import (
	"fmt"
	"log"
	"sync"
	"time"

	"btc-alerta-de-precio/internal/bitcoin"
	"btc-alerta-de-precio/internal/notifications"
	"btc-alerta-de-precio/internal/storage"
)

type Service struct {
	db                  *storage.Database
	bitcoinClient       *bitcoin.Client
	notificationService *notifications.Service

	// Control del monitoreo
	isMonitoring  bool
	stopChannel   chan bool
	monitoringMux sync.RWMutex

	// Cache del último precio
	lastPrice    *bitcoin.PriceData
	lastPriceMux sync.RWMutex
}

func NewService(db *storage.Database, bitcoinClient *bitcoin.Client, notificationService *notifications.Service) *Service {
	return &Service{
		db:                  db,
		bitcoinClient:       bitcoinClient,
		notificationService: notificationService,
		stopChannel:         make(chan bool),
	}
}

func (s *Service) StartMonitoring(interval time.Duration) {
	s.monitoringMux.Lock()
	defer s.monitoringMux.Unlock()

	if s.isMonitoring {
		log.Println("El monitoreo ya está activo")
		return
	}

	s.isMonitoring = true
	log.Printf("Iniciando monitoreo de precio de Bitcoin cada %v", interval)

	go s.monitoringLoop(interval)
}

func (s *Service) Stop() {
	s.monitoringMux.Lock()
	defer s.monitoringMux.Unlock()

	if !s.isMonitoring {
		return
	}

	log.Println("Deteniendo monitoreo...")
	s.isMonitoring = false
	s.stopChannel <- true
}

func (s *Service) IsMonitoring() bool {
	s.monitoringMux.RLock()
	defer s.monitoringMux.RUnlock()
	return s.isMonitoring
}

func (s *Service) monitoringLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Primera verificación inmediata
	s.checkPriceAndAlerts()

	for {
		select {
		case <-ticker.C:
			s.checkPriceAndAlerts()
		case <-s.stopChannel:
			log.Println("Monitoreo detenido")
			return
		}
	}
}

func (s *Service) checkPriceAndAlerts() {
	// Obtener precio actual
	currentPrice, err := s.bitcoinClient.GetCurrentPrice()
	if err != nil {
		log.Printf("Error obteniendo precio de Bitcoin: %v", err)
		return
	}

	log.Printf("Precio actual de Bitcoin: %s", currentPrice.String())

	// Guardar en historial
	priceHistory := &storage.PriceHistory{
		Price:     currentPrice.Price,
		Currency:  currentPrice.Currency,
		Source:    currentPrice.Source,
		Timestamp: currentPrice.Timestamp,
	}

	if err := s.db.SavePriceHistory(priceHistory); err != nil {
		log.Printf("Error guardando historial de precios: %v", err)
	}

	// Obtener precio anterior para comparaciones
	s.lastPriceMux.RLock()
	lastPrice := s.lastPrice
	s.lastPriceMux.RUnlock()

	var previousPrice float64
	if lastPrice != nil {
		previousPrice = lastPrice.Price
	}

	// Actualizar último precio
	s.lastPriceMux.Lock()
	s.lastPrice = currentPrice
	s.lastPriceMux.Unlock()

	// Verificar alertas
	s.checkAlerts(currentPrice.Price, previousPrice)
}

func (s *Service) checkAlerts(currentPrice, previousPrice float64) {
	alerts, err := s.db.GetActiveAlerts()
	if err != nil {
		log.Printf("Error obteniendo alertas activas: %v", err)
		return
	}

	for _, alert := range alerts {
		if alert.ShouldTrigger(currentPrice, previousPrice) {
			s.triggerAlert(&alert, currentPrice)
		}
	}
}

func (s *Service) triggerAlert(alert *storage.Alert, currentPrice float64) {
	log.Printf("Disparando alerta: %s (Precio: $%.2f)", alert.Name, currentPrice)

	// Marcar alerta como activada
	alert.MarkTriggered()
	if err := s.db.UpdateAlert(alert); err != nil {
		log.Printf("Error actualizando alerta: %v", err)
	}

	// Preparar datos de notificación
	notificationData := &notifications.NotificationData{
		Title:   fmt.Sprintf("🚨 Alerta de Bitcoin: %s", alert.Name),
		Message: s.generateAlertMessage(alert, currentPrice),
		Price:   currentPrice,
		Alert:   alert,
	}

	// Enviar notificación
	if err := s.notificationService.SendAlert(notificationData); err != nil {
		log.Printf("Error enviando notificación para alerta %s: %v", alert.Name, err)

		// Registrar error en log de notificaciones
		s.logNotification(alert.ID, "error", err.Error())
	} else {
		log.Printf("Notificación enviada exitosamente para alerta: %s", alert.Name)

		// Registrar éxito en log de notificaciones
		s.logNotification(alert.ID, "sent", "Notification sent successfully")
	}
}

func (s *Service) generateAlertMessage(alert *storage.Alert, currentPrice float64) string {
	switch alert.Type {
	case "above":
		return fmt.Sprintf("El precio de Bitcoin ha superado $%.2f. Precio actual: $%.2f", alert.TargetPrice, currentPrice)
	case "below":
		return fmt.Sprintf("El precio de Bitcoin ha caído por debajo de $%.2f. Precio actual: $%.2f", alert.TargetPrice, currentPrice)
	case "change":
		// Calcular cambio porcentual (requiere precio anterior)
		s.lastPriceMux.RLock()
		lastPrice := s.lastPrice
		s.lastPriceMux.RUnlock()

		if lastPrice != nil && lastPrice.Price > 0 {
			changePercent := ((currentPrice - lastPrice.Price) / lastPrice.Price) * 100
			direction := "subido"
			if changePercent < 0 {
				direction = "bajado"
				changePercent = -changePercent
			}
			return fmt.Sprintf("El precio de Bitcoin ha %s %.2f%%. Precio actual: $%.2f", direction, changePercent, currentPrice)
		}
		return fmt.Sprintf("Cambio significativo en el precio de Bitcoin. Precio actual: $%.2f", currentPrice)
	default:
		return fmt.Sprintf("Alerta de Bitcoin activada. Precio actual: $%.2f", currentPrice)
	}
}

func (s *Service) logNotification(alertID uint, status, message string) {
	notificationLog := &storage.NotificationLog{
		AlertID: alertID,
		Type:    "combined", // email + desktop
		Status:  status,
		Message: message,
		SentAt:  time.Now(),
	}

	if err := s.db.LogNotification(notificationLog); err != nil {
		log.Printf("Error registrando log de notificación: %v", err)
	}
}

// Métodos de la API para gestionar alertas

func (s *Service) CreateAlert(alert *storage.Alert) error {
	return s.db.CreateAlert(alert)
}

func (s *Service) GetAlert(id uint) (*storage.Alert, error) {
	return s.db.GetAlert(id)
}

func (s *Service) GetAlerts() ([]storage.Alert, error) {
	return s.db.GetAlerts()
}

func (s *Service) UpdateAlert(alert *storage.Alert) error {
	return s.db.UpdateAlert(alert)
}

func (s *Service) DeleteAlert(id uint) error {
	return s.db.DeleteAlert(id)
}

func (s *Service) ToggleAlert(id uint) error {
	return s.db.ToggleAlert(id)
}

func (s *Service) GetCurrentPrice() (*bitcoin.PriceData, error) {
	return s.bitcoinClient.GetCurrentPrice()
}

func (s *Service) GetPriceHistory(limit int) ([]storage.PriceHistory, error) {
	return s.db.GetPriceHistory(limit)
}

func (s *Service) GetStats() (map[string]interface{}, error) {
	stats, err := s.db.GetStats()
	if err != nil {
		return nil, err
	}

	// Agregar información del monitoreo
	stats["monitoring_active"] = s.IsMonitoring()

	// Agregar precio actual si está disponible
	s.lastPriceMux.RLock()
	if s.lastPrice != nil {
		stats["current_price"] = s.lastPrice.Price
		stats["current_price_source"] = s.lastPrice.Source
		stats["current_price_time"] = s.lastPrice.Timestamp
	}
	s.lastPriceMux.RUnlock()

	return stats, nil
}

func (s *Service) TestAlert(id uint) error {
	alert, err := s.db.GetAlert(id)
	if err != nil {
		return err
	}

	currentPrice, err := s.bitcoinClient.GetCurrentPrice()
	if err != nil {
		return err
	}

	// Crear una copia de la alerta para testing (no actualizar la original)
	testAlert := *alert
	testAlert.Name = "🧪 TEST: " + alert.Name

	notificationData := &notifications.NotificationData{
		Title:   fmt.Sprintf("🧪 Test de Alerta: %s", alert.Name),
		Message: fmt.Sprintf("Esta es una prueba de la alerta '%s'. Precio actual: $%.2f", alert.Name, currentPrice.Price),
		Price:   currentPrice.Price,
		Alert:   &testAlert,
	}

	return s.notificationService.SendAlert(notificationData)
}

// ResetAlert resetea una alerta para que pueda dispararse de nuevo
func (s *Service) ResetAlert(alertID uint) error {
	alert, err := s.db.GetAlert(alertID)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}

	// Resetear la alerta usando el método del modelo
	alert.Reset()

	// Guardar los cambios en la base de datos
	if err := s.db.UpdateAlert(alert); err != nil {
		return fmt.Errorf("failed to reset alert: %w", err)
	}

	log.Printf("🔄 Alerta reseteada: %s (ID: %d)", alert.Name, alertID)
	return nil
}
