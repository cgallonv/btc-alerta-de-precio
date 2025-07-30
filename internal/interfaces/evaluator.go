package interfaces

import (
	"github.com/cgallonv/btc-alerta-de-precio/internal/bitcoin"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage"
)

// AlertEvaluator defines the interface for evaluating alert conditions
type AlertEvaluator interface {
	ShouldTrigger(alert *storage.Alert, priceData *bitcoin.PriceData) bool
}
