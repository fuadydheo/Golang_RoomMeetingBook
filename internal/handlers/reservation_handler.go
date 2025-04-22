package handlers

import (
	"e-meetingproject/internal/models"
	"e-meetingproject/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReservationHandler struct {
	service *services.ReservationService
}

func NewReservationHandler(service *services.ReservationService) *ReservationHandler {
	return &ReservationHandler{
		service: service,
	}
}

func (h *ReservationHandler) GetReservationHistory(c *gin.Context) {
	var query models.ReservationHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in token"})
		return
	}

	// Convert userID to UUID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
		return
	}

	response, err := h.service.GetReservationHistory(&query, userUUID)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReservationHandler) UpdateReservationStatus(c *gin.Context) {
	var req models.UpdateReservationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedReservation, err := h.service.UpdateReservationStatus(&req)
	if err != nil {
		if err.Error() == "reservation not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "invalid status") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedReservation)
}

func (h *ReservationHandler) CalculateReservationCost(c *gin.Context) {
	var req models.ReservationCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate time range
	if req.EndTime.Before(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end time must be after start time"})
		return
	}

	// Calculate costs
	response, err := h.service.CalculateReservationCost(&req)
	if err != nil {
		if err.Error() == "room not found or inactive" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ReservationHandler) CreateReservation(c *gin.Context) {
	var req models.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate time range
	if req.EndTime.Before(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end time must be after start time"})
		return
	}

	// Create reservation
	response, err := h.service.CreateReservation(&req)
	if err != nil {
		if err.Error() == "room not found or inactive" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "visitor count exceeds room capacity" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "room is already booked for the selected time period" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}
