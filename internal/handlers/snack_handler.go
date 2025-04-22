package handlers

import (
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
