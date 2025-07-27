package api

import (
	"fmt"
	"net/http"
	"strconv"

	"btc-alerta-de-precio/internal/interfaces"
	"btc-alerta-de-precio/internal/storage"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	alertService interfaces.AlertService
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

func NewHandler(alertService interfaces.AlertService) *Handler {
	return &Handler{
		alertService: alertService,
	}
}

func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Servir archivos estáticos
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// Página principal
	router.GET("/", h.indexPage)

	// API routes
	api := router.Group("/api/v1")
	{
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

		// System
		api.GET("/health", h.healthCheck)
	}
}

// Páginas web
func (h *Handler) indexPage(c *gin.Context) {
	stats, _ := h.alertService.GetStats()
	c.HTML(http.StatusOK, "index.html", gin.H{
		"stats": stats,
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
	if alert.Type == "above" || alert.Type == "below" {
		if updateReq.TargetPrice != nil {
			alert.TargetPrice = *updateReq.TargetPrice
		} else {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error:   "target_price is required for price-based alerts",
			})
			return
		}
	} else if alert.Type == "change" {
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
	// TODO: Obtener la clave desde la configuración
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"publicKey": "BI3enbbkk8hud4SXGXriy9wPEBovCg210LDckVrM5ldTzkbXCwEGZLGegjhwkTrOd9z152h4iLtTCrqOP_UzV-M",
		},
	})
}

func (h *Handler) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Configuration retrieved successfully",
		Data: gin.H{
			"check_interval_ms": 30000, // 30 segundos en milisegundos para JavaScript
		},
	})
}
