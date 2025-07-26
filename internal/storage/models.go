package storage

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Alert struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Type        string    `json:"type" gorm:"not null"` // "above", "below", "change"
	TargetPrice float64   `json:"target_price"`
	Percentage  float64   `json:"percentage"` // Para alertas de cambio porcentual
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Configuración de notificaciones
	EnableEmail    bool `json:"enable_email" gorm:"default:true"`
	EnableWebPush  bool `json:"enable_web_push" gorm:"default:false"`
	EnableTelegram bool `json:"enable_telegram" gorm:"default:false"`

	// Tracking de activaciones
	LastTriggered *time.Time `json:"last_triggered"`
	TriggerCount  int        `json:"trigger_count" gorm:"default:0"`
}

type PriceHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Price     float64   `json:"price" gorm:"not null"`
	Currency  string    `json:"currency" gorm:"default:'USD'"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	AlertID   uint      `json:"alert_id" gorm:"not null"`
	Type      string    `json:"type" gorm:"not null"`   // "email", "desktop", "web_push"
	Status    string    `json:"status" gorm:"not null"` // "sent", "failed", "pending"
	Message   string    `json:"message"`
	Error     string    `json:"error"`
	SentAt    time.Time `json:"sent_at"`
	CreatedAt time.Time `json:"created_at"`

	// Relación con Alert
	Alert Alert `json:"alert" gorm:"foreignKey:AlertID"`
}

type WebPushSubscription struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Endpoint  string    `json:"endpoint" gorm:"not null;unique"`
	P256dh    string    `json:"p256dh" gorm:"not null"`
	Auth      string    `json:"auth" gorm:"not null"`
	UserID    string    `json:"user_id"` // Opcional, para asociar con usuarios
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Métodos para Alert
func (a *Alert) ShouldTrigger(currentPrice float64, previousPrice float64) bool {
	if !a.IsActive {
		return false
	}

	// One-Shot: Solo disparar si nunca se ha activado antes
	if a.LastTriggered != nil {
		return false
	}

	switch a.Type {
	case "above":
		return currentPrice >= a.TargetPrice
	case "below":
		return currentPrice <= a.TargetPrice
	case "change":
		if previousPrice == 0 {
			return false
		}
		changePercent := ((currentPrice - previousPrice) / previousPrice) * 100
		return changePercent >= a.Percentage || changePercent <= -a.Percentage
	default:
		return false
	}
}

func (a *Alert) GetDescription() string {
	switch a.Type {
	case "above":
		return fmt.Sprintf("Bitcoin price above $%.2f", a.TargetPrice)
	case "below":
		return fmt.Sprintf("Bitcoin price below $%.2f", a.TargetPrice)
	case "change":
		return fmt.Sprintf("Bitcoin price change of %.2f%%", a.Percentage)
	default:
		return "Unknown alert type"
	}
}

func (a *Alert) MarkTriggered() {
	now := time.Now()
	a.LastTriggered = &now
	a.TriggerCount++
}

// ResetAlert resetea una alerta para poder dispararse de nuevo
func (a *Alert) Reset() {
	a.LastTriggered = nil
	// TriggerCount se mantiene como historial
}

// Validaciones
func (a *Alert) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("alert name is required")
	}

	if a.Type != "above" && a.Type != "below" && a.Type != "change" {
		return fmt.Errorf("alert type must be 'above', 'below', or 'change'")
	}

	if (a.Type == "above" || a.Type == "below") && a.TargetPrice <= 0 {
		return fmt.Errorf("target price must be greater than 0")
	}

	if a.Type == "change" && (a.Percentage <= 0 || a.Percentage > 100) {
		return fmt.Errorf("percentage must be between 0 and 100")
	}

	if a.EnableEmail && a.Email == "" {
		return fmt.Errorf("email is required when email notifications are enabled")
	}

	return nil
}

// Hook para GORM - ejecutar antes de crear
func (a *Alert) BeforeCreate(tx *gorm.DB) error {
	return a.Validate()
}

// Hook para GORM - ejecutar antes de actualizar
func (a *Alert) BeforeUpdate(tx *gorm.DB) error {
	return a.Validate()
}
