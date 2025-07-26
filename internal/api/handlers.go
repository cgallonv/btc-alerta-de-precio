package api

import (
	"net/http"
	"strconv"

	"btc-alerta-de-precio/internal/alerts"
	"btc-alerta-de-precio/internal/storage"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	alertService *alerts.Service
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func NewHandler(alertService *alerts.Service) *Handler {
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

		// Alerts
		api.GET("/alerts", h.getAlerts)
		api.GET("/alerts/:id", h.getAlert)
		api.POST("/alerts", h.createAlert)
		api.PUT("/alerts/:id", h.updateAlert)
		api.DELETE("/alerts/:id", h.deleteAlert)
		api.POST("/alerts/:id/toggle", h.toggleAlert)
		api.POST("/alerts/:id/test", h.testAlert)

		// Stats
		api.GET("/stats", h.getStats)

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

	var alert storage.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	alert.ID = uint(id)
	if err := h.alertService.UpdateAlert(&alert); err != nil {
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
