package api

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/storage"

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

func NewHandler(alertService interfaces.AlertService, configProvider interfaces.ConfigProvider) *Handler {
	return &Handler{
		alertService:   alertService,
		configProvider: configProvider,
	}
}

func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Servir archivos estáticos
	router.Static("/static", "./web/static")
	router.StaticFile("/sw.js", "./web/sw.js") // Serve service worker at root

	// Load templates programmatically
	templ := template.New("")
	templ = template.Must(templ.ParseFiles(
		"web/templates/index.html",
		"web/templates/alerts.html",
		"web/templates/partials/edit_alert_modal.html",
		"web/templates/partials/hamburger_menu.html",
		"web/templates/partials/top_bar.html",
	))
	router.SetHTMLTemplate(templ)

	// Página principal
	router.GET("/", h.indexPage)
	router.GET("/alerts", h.alertsPage)

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

		// Web Push Subscriptions
		api.POST("/webpush/subscribe", h.subscribeWebPush)
		api.DELETE("/webpush/unsubscribe", h.unsubscribeWebPush)
		api.GET("/webpush/vapid-public-key", h.getVAPIDPublicKey)
		api.GET("/config", h.getConfig)
		api.POST("/preload-alerts", h.preloadAlerts)      // 🆕 Endpoint para precargar alertas
		api.POST("/delete-all-alerts", h.deleteAllAlerts) // 🆕 Endpoint para eliminar todas las alertas
	}
}

// Páginas web
func (h *Handler) indexPage(c *gin.Context) {
	stats, _ := h.alertService.GetStats()
	c.HTML(http.StatusOK, "index.html", gin.H{
		"stats":       stats,
		"CurrentPage": "dashboard",
		"Version":     time.Now().Unix(), // Add timestamp as version
	})
}

// Alerts page
func (h *Handler) alertsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "alerts.html", gin.H{
		"CurrentPage": "alerts",
		"Version":     time.Now().Unix(), // Add timestamp as version
	})
}

// Bitcoin Price endpoints
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

// Web Push handlers
func (h *Handler) subscribeWebPush(c *gin.Context) {
	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
		P256dh   string `json:"p256dh" binding:"required"`
		Auth     string `json:"auth" binding:"required"`
		UserID   string `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	subscription := &storage.WebPushSubscription{
		Endpoint: req.Endpoint,
		P256dh:   req.P256dh,
		Auth:     req.Auth,
		UserID:   req.UserID,
		IsActive: true,
	}

	db, err := storage.NewDatabase("alerts.db") // TODO: Usar configuración
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Database connection failed",
		})
		return
	}

	if err := db.SaveWebPushSubscription(subscription); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Web Push subscription saved successfully",
	})
}

func (h *Handler) unsubscribeWebPush(c *gin.Context) {
	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	db, err := storage.NewDatabase("alerts.db") // TODO: Usar configuración
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Database connection failed",
		})
		return
	}

	if err := db.RemoveWebPushSubscription(req.Endpoint); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Web Push subscription removed successfully",
	})
}

func (h *Handler) getVAPIDPublicKey(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"publicKey": h.configProvider.GetVAPIDPublicKey(),
		},
	})
}

// GET /api/config
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
