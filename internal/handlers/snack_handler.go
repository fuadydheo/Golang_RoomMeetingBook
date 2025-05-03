package handlers

import (
	"e-meetingproject/internal/models"
	"e-meetingproject/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SnackHandler struct {
	service *services.SnackService
}

func NewSnackHandler(service *services.SnackService) *SnackHandler {
	return &SnackHandler{
		service: service,
	}
}

func (h *SnackHandler) GetSnacks(c *gin.Context) {
	// Get snacks from service
	response, err := h.service.GetSnacks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SnackHandler) CreateSnack(c *gin.Context) {
	var req models.CreateSnackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate price is positive
	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be greater than zero"})
		return
	}

	// Create snack
	response, err := h.service.CreateSnack(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}
