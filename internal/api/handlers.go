package api

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cgallonv/btc-alerta-de-precio/internal/interfaces"
	"github.com/cgallonv/btc-alerta-de-precio/internal/storage"

	"github.com/cgallonv/btc-alerta-de-precio/internal/bitcoin"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	alertService   interfaces.AlertService
	configProvider interfaces.ConfigProvider
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// AlertUpdateRequest para la funcionalidad de edición limitada
type AlertUpdateRequest struct {
	TargetPrice *float64 `json:"target_price,omitempty"`
	Percentage  *float64 `json:"percentage,omitempty"`
}

// Add AccountData struct
type AccountData struct {
	TotalBalance     float64   `json:"total_balance"`
	AvailableBalance float64   `json:"available_balance"`
	LastUpdated      time.Time `json:"last_updated"`
	Assets           []Asset   `json:"assets"`
	Orders           []Order   `json:"orders"`
	Budget           Budget    `json:"budget"`

	// Binance API fields
	MakerCommission  int `json:"maker_commission"`
	TakerCommission  int `json:"taker_commission"`
	BuyerCommission  int `json:"buyer_commission"`
	SellerCommission int `json:"seller_commission"`
	CommissionRates  struct {
		Maker  string `json:"maker"`
		Taker  string `json:"taker"`
		Buyer  string `json:"buyer"`
		Seller string `json:"seller"`
	} `json:"commission_rates"`
	CanTrade                   bool     `json:"can_trade"`
	CanWithdraw                bool     `json:"can_withdraw"`
	CanDeposit                 bool     `json:"can_deposit"`
	AccountType                string   `json:"account_type"`
	Permissions                []string `json:"permissions"`
	RequireSelfTradePrevention bool     `json:"require_self_trade_prevention"`
	PreventSor                 bool     `json:"prevent_sor"`
	UpdateTime                 int64    `json:"update_time"`
}

type Asset struct {
	Symbol    string  `json:"symbol"`
	Free      string  `json:"free"`
	Locked    string  `json:"locked"`
	Total     float64 `json:"total"`
	ValueUSD  float64 `json:"value_usd"`
	Change24h float64 `json:"change_24h"`
}

type Order struct {
	ID     string    `json:"id"`
	Date   time.Time `json:"date"`
	Type   string    `json:"type"`
	Amount float64   `json:"amount"`
	Price  float64   `json:"price"`
	Status string    `json:"status"`
}

type Budget struct {
	Used            float64    `json:"used"`
	Limit           float64    `json:"limit"`
	UsagePercentage float64    `json:"usage_percentage"`
	Categories      []Category `json:"categories"`
}

type Category struct {
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

// NewHandler creates a new Handler with the given alert service and config provider.
func NewHandler(alertService interfaces.AlertService, configProvider interfaces.ConfigProvider) *Handler {
	return &Handler{
		alertService:   alertService,
		configProvider: configProvider,
	}
}

// SetupRoutes configures all HTTP routes for the application.
// Example usage:
//
//	router := gin.Default()
//	handler := NewHandler(...)
//	handler.SetupRoutes(router)
func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Static files
	router.Static("/static", "./web/static")

	// Load templates in the correct order
	templ := template.New("")
	var err error
	templ, err = templ.ParseFiles(
		"web/templates/layout.html",
		"web/templates/partials/top_bar.html",
		"web/templates/partials/hamburger_menu.html",
		"web/templates/partials/edit_alert_modal.html",
		"web/templates/partials/alerts_form.html",
		"web/templates/partials/alerts_list.html",
		"web/templates/index.html",
		"web/templates/alerts.html",
		"web/templates/account.html",
		"web/templates/partials/account_balance.html",
		"web/templates/partials/account_assets.html",
		"web/templates/partials/account_orders.html",
		"web/templates/partials/account_budget.html",
	)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		return
	}
	router.SetHTMLTemplate(templ)

	// Routes
	router.GET("/", h.indexPage)
	router.GET("/alerts", h.alertsPage)
	router.GET("/account", h.accountPage)

	// API routes
	api := router.Group("/api/v1")
	{
		// System
		api.GET("/health", h.healthCheck)
		api.HEAD("/health", h.healthCheck) // Add HEAD request support

		// Bitcoin price
		api.GET("/price", h.getCurrentPrice)
		api.GET("/price/history", h.getPriceHistory)
		api.GET("/price/percentage", h.getCurrentPercentage)

		// Account
		api.GET("/account/balance", h.GetAccountBalance)

		// Alerts
		api.GET("/alerts", h.getAlerts)
		api.GET("/alerts/:id", h.getAlert)
		api.POST("/alerts", h.createAlert)
		api.PUT("/alerts/:id", h.updateAlert)
		api.DELETE("/alerts/:id", h.deleteAlert)
		api.POST("/alerts/:id/toggle", h.toggleAlert)
		api.POST("/alerts/:id/test", h.testAlert)
		api.POST("/alerts/:id/reset", h.resetAlert)

		// Stats
		api.GET("/stats", h.getStats)

		// Config
		api.GET("/config", h.getConfig)

		// Development utilities
		api.POST("/preload-alerts", h.preloadAlerts)      // 🆕 Endpoint para precargar alertas
		api.POST("/delete-all-alerts", h.deleteAllAlerts) // 🆕 Endpoint para eliminar todas las alertas
	}
}

// indexPage renders the dashboard page.
// Route: GET /
func (h *Handler) indexPage(c *gin.Context) {
	stats, err := h.alertService.GetStats()
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		c.HTML(http.StatusInternalServerError, "layout", gin.H{
			"error":     "Error loading stats",
			"PageTitle": "Dashboard",
			"Version":   time.Now().Unix(),
			"content":   "index",
		})
		return
	}

	log.Printf("Rendering index page with stats: %+v", stats)
	c.HTML(http.StatusOK, "layout", gin.H{
		"stats":     stats,
		"PageTitle": "Dashboard",
		"Version":   time.Now().Unix(),
		"content":   "index",
	})
}

// alertsPage renders the alerts management page.
// Route: GET /alerts
func (h *Handler) alertsPage(c *gin.Context) {
	alerts, err := h.alertService.GetAlerts()
	if err != nil {
		log.Printf("Error getting alerts: %v", err)
		c.HTML(http.StatusInternalServerError, "layout", gin.H{
			"error":     "Error loading alerts",
			"PageTitle": "Alertas",
			"Version":   time.Now().Unix(),
			"content":   "alerts",
		})
		return
	}

	log.Printf("Rendering alerts page with %d alerts", len(alerts))
	c.HTML(http.StatusOK, "layout", gin.H{
		"alerts":    alerts,
		"PageTitle": "Alertas",
		"Version":   time.Now().Unix(),
		"content":   "alerts",
	})
}

// accountPage renders the account management page.
// Route: GET /account
func (h *Handler) accountPage(c *gin.Context) {
	// Create Binance client
	binanceClient := bitcoin.NewBinanceClient(
		h.configProvider.GetString("binance.api_key"),
		h.configProvider.GetString("binance.api_secret"),
		h.configProvider.GetString("binance.base_url"),
		nil, // No ticker storage needed for account page
	)

	// Get account balance using default symbols from config
	balance, err := binanceClient.GetAccountBalance(h.configProvider.GetDefaultSymbols())
	if err != nil {
		log.Printf("Error getting account balance: %v", err)
		c.HTML(http.StatusInternalServerError, "layout", gin.H{
			"error":     "Failed to get account balance",
			"PageTitle": "Account",
			"Version":   time.Now().Unix(),
			"content":   "account",
		})
		return
	}

	// Prepare account data
	accountData := AccountData{
		TotalBalance:     balance.TotalBalance,
		AvailableBalance: balance.AvailableBalance,
		LastUpdated:      balance.LastUpdated,
		Assets:           make([]Asset, 0, len(balance.Assets)),
		Orders:           []Order{}, // TODO: Implement order history
		Budget:           Budget{},  // TODO: Implement budget history

		// Binance API fields
		MakerCommission:            balance.MakerCommission,
		TakerCommission:            balance.TakerCommission,
		BuyerCommission:            balance.BuyerCommission,
		SellerCommission:           balance.SellerCommission,
		CommissionRates:            balance.CommissionRates,
		CanTrade:                   balance.CanTrade,
		CanWithdraw:                balance.CanWithdraw,
		CanDeposit:                 balance.CanDeposit,
		AccountType:                balance.AccountType,
		Permissions:                balance.Permissions,
		RequireSelfTradePrevention: balance.RequireSelfTradePrevention,
		PreventSor:                 balance.PreventSor,
		UpdateTime:                 balance.UpdateTime,
	}

	// Convert assets
	for _, asset := range balance.Assets {
		accountData.Assets = append(accountData.Assets, Asset{
			Symbol:    asset.Symbol,
			Free:      asset.Free,
			Locked:    asset.Locked,
			Total:     asset.Total,
			ValueUSD:  asset.ValueUSD,
			Change24h: asset.Change24h,
		})
	}

	c.HTML(http.StatusOK, "layout", gin.H{
		"PageTitle": "Account",
		"Version":   time.Now().Unix(),
		"content":   "account",
		"account":   accountData,
	})
}

// getCurrentPrice handles GET /api/v1/price and returns the current Bitcoin price as JSON.
// Example usage:
//
//	GET /api/v1/price
func (h *Handler) getCurrentPrice(c *gin.Context) {
	price, err := h.alertService.GetCurrentPrice()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    price,
	})
}

// getPriceHistory handles GET /api/v1/price/history and returns the price history.
// Example usage:
//
//	GET /api/v1/price/history?limit=24
func (h *Handler) getPriceHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid limit parameter",
		})
		return
	}

	history, err := h.alertService.GetPriceHistory(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    history,
	})
}

// getCurrentPercentage handles GET /api/v1/price/percentage and returns the current price change percentage.
func (h *Handler) getCurrentPercentage(c *gin.Context) {
	percentage := h.alertService.GetCurrentPercentage()

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"percentage": percentage,
			"formatted":  fmt.Sprintf("%+.2f%%", percentage),
		},
	})
}

// Alert endpoints
// getAlerts handles GET /api/v1/alerts and returns all alerts.
func (h *Handler) getAlerts(c *gin.Context) {
	alerts, err := h.alertService.GetAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    alerts,
	})
}

// getAlert handles GET /api/v1/alerts/:id and returns a specific alert by ID.
func (h *Handler) getAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid alert ID",
		})
		return
	}

	alert, err := h.alertService.GetAlert(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error:   "Alert not found",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    alert,
	})
}

// createAlert handles POST /api/v1/alerts and creates a new alert.
func (h *Handler) createAlert(c *gin.Context) {
	var alert storage.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := h.alertService.CreateAlert(&alert); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    alert,
		Message: "Alert created successfully",
	})
}

// updateAlert handles PUT /api/v1/alerts/:id and updates an existing alert.
func (h *Handler) updateAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid alert ID",
		})
		return
	}

	var updateReq AlertUpdateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Obtener la alerta actual
	alert, err := h.alertService.GetAlert(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error:   "Alert not found",
		})
		return
	}

	// Solo actualizar el campo correspondiente según el tipo de alerta
	switch alert.Type {
	case "above", "below":
		if updateReq.TargetPrice != nil {
			alert.TargetPrice = *updateReq.TargetPrice
		} else {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error:   "target_price is required for price-based alerts",
			})
			return
		}
	case "change":
		if updateReq.Percentage != nil {
			alert.Percentage = *updateReq.Percentage
		} else {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error:   "percentage is required for percentage-based alerts",
			})
			return
		}
	}

	// Si la alerta estaba disparada, resetearla para que pueda activarse de nuevo
	if alert.LastTriggered != nil {
		alert.Reset()
	}

	if err := h.alertService.UpdateAlert(alert); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    alert,
		Message: "Alert updated successfully",
	})
}

// deleteAlert handles DELETE /api/v1/alerts/:id and deletes an alert by ID.
func (h *Handler) deleteAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid alert ID",
		})
		return
	}

	if err := h.alertService.DeleteAlert(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Alert deleted successfully",
	})
}

// toggleAlert handles POST /api/v1/alerts/:id/toggle and toggles the active state of an alert.
func (h *Handler) toggleAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid alert ID",
		})
		return
	}

	if err := h.alertService.ToggleAlert(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Alert toggled successfully",
	})
}

// testAlert handles POST /api/v1/alerts/:id/test and sends a test notification for an alert.
func (h *Handler) testAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid alert ID",
		})
		return
	}

	if err := h.alertService.TestAlert(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Test notification sent successfully",
	})
}

// resetAlert handles POST /api/v1/alerts/:id/reset and resets an alert so it can be triggered again.
func (h *Handler) resetAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid alert ID",
		})
		return
	}

	if err := h.alertService.ResetAlert(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Alert reset successfully",
	})
}

// System endpoints
// getStats handles GET /api/v1/stats and returns system statistics.
func (h *Handler) getStats(c *gin.Context) {
	stats, err := h.alertService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// healthCheck handles GET/HEAD /api/v1/health and returns a health status for the service.
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Service is healthy",
		Data: gin.H{
			"status":     "ok",
			"monitoring": h.alertService.IsMonitoring(),
		},
	})
}

// GET /api/config
// getConfig handles GET /api/v1/config and returns configuration values.
func (h *Handler) getConfig(c *gin.Context) {
	checkInterval := h.configProvider.GetCheckInterval()
	checkIntervalMs := int(checkInterval.Milliseconds())

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Configuration retrieved successfully",
		Data: gin.H{
			"check_interval_ms": checkIntervalMs,
		},
	})
}

// POST /api/v1/preload-alerts
// preloadAlerts handles POST /api/v1/preload-alerts and creates a set of sample alerts for testing.
func (h *Handler) preloadAlerts(c *gin.Context) {
	alerts := []storage.Alert{
		{
			Name:           "Precio por debajo de 117000",
			Type:           "below",
			TargetPrice:    117000,
			Percentage:     0,
			IsActive:       true,
			Email:          "cgallonv@gmail.com",
			EnableEmail:    true,
			EnableTelegram: true,
		},
		{
			Name:           "Precio por debajo de 116000",
			Type:           "below",
			TargetPrice:    116000,
			Percentage:     0,
			IsActive:       true,
			Email:          "cgallonv@gmail.com",
			EnableEmail:    true,
			EnableTelegram: true,
		},
		{
			Name:           "Precio por debajo de 115000",
			Type:           "below",
			TargetPrice:    115000,
			Percentage:     0,
			IsActive:       true,
			Email:          "cgallonv@gmail.com",
			EnableEmail:    true,
			EnableTelegram: true,
		},
		{
			Name:           "Bajó 3%",
			Type:           "change",
			TargetPrice:    0,
			Percentage:     -3,
			IsActive:       true,
			Email:          "cgallonv@gmail.com",
			EnableEmail:    true,
			EnableTelegram: true,
		},
		{
			Name:           "Bajó 4%",
			Type:           "change",
			TargetPrice:    0,
			Percentage:     -4,
			IsActive:       true,
			Email:          "cgallonv@gmail.com",
			EnableEmail:    true,
			EnableTelegram: true,
		},
		{
			Name:           "Bajó 5%",
			Type:           "change",
			TargetPrice:    0,
			Percentage:     -5,
			IsActive:       true,
			Email:          "cgallonv@gmail.com",
			EnableEmail:    true,
			EnableTelegram: true,
		},
		{
			Name:           "PRECIO BAJÓ 6%!!!!!!",
			Type:           "change",
			TargetPrice:    0,
			Percentage:     -6,
			IsActive:       true,
			Email:          "cgallonv@gmail.com",
			EnableEmail:    true,
			EnableTelegram: true,
		},
		{
			Name:           "PRECIO BAJÓ 7%!!!!!!",
			Type:           "change",
			TargetPrice:    0,
			Percentage:     -7,
			IsActive:       true,
			Email:          "cgallonv@gmail.com",
			EnableEmail:    true,
			EnableTelegram: true,
		},
	}

	success := 0
	errors := 0
	for _, alert := range alerts {
		err := h.alertService.CreateAlert(&alert)
		if err != nil {
			errors++
		} else {
			success++
		}
	}

	c.JSON(http.StatusOK, Response{
		Success: errors == 0,
		Message: fmt.Sprintf("%d alertas precargadas, %d errores", success, errors),
		Data: gin.H{
			"success": success,
			"errors":  errors,
		},
	})
}

// POST /api/v1/delete-all-alerts
// deleteAllAlerts handles POST /api/v1/delete-all-alerts and deletes all alerts from the system.
func (h *Handler) deleteAllAlerts(c *gin.Context) {
	alerts, err := h.alertService.GetAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Error obteniendo alertas: " + err.Error(),
		})
		return
	}
	count := 0
	errors := 0
	for _, alert := range alerts {
		err := h.alertService.DeleteAlert(alert.ID)
		if err != nil {
			errors++
		} else {
			count++
		}
	}
	c.JSON(http.StatusOK, Response{
		Success: errors == 0,
		Message: fmt.Sprintf("%d alertas eliminadas, %d errores", count, errors),
		Data: gin.H{
			"deleted": count,
			"errors":  errors,
		},
	})
}

// GetAccountBalance returns the current account balance
// Example: GET /api/v1/account/balance?symbols=BTC,USDT,COP
func (h *Handler) GetAccountBalance(c *gin.Context) {
	// Get symbols from query parameter or use defaults
	symbolsParam := c.Query("symbols")
	var symbols []string
	if symbolsParam != "" {
		symbols = strings.Split(symbolsParam, ",")
	} else {
		// Use default symbols from config if none provided
		symbols = h.configProvider.GetDefaultSymbols()
	}
	// Create Binance client
	binanceClient := bitcoin.NewBinanceClient(
		h.configProvider.GetString("binance.api_key"),
		h.configProvider.GetString("binance.api_secret"),
		h.configProvider.GetString("binance.base_url"),
		nil, // No ticker storage needed for account balance
	)

	// Get account balance
	balance, err := binanceClient.GetAccountBalance(symbols)
	if err != nil {
		log.Printf("Error getting account balance: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    balance,
	})
}
