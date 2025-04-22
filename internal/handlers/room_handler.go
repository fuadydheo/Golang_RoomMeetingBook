package handlers

import (
	"e-meetingproject/internal/models"
	"e-meetingproject/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomHandler struct {
	service *services.RoomService
}

func NewRoomHandler(service *services.RoomService) *RoomHandler {
	return &RoomHandler{
		service: service,
	}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req models.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.service.CreateRoom(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, room)
}

func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID format"})
		return
	}

	var req models.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.service.UpdateRoom(id, &req)
	if err != nil {
		if err.Error() == "room not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, room)
}

func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID format"})
		return
	}

	err = h.service.DeleteRoom(id)
	if err != nil {
		switch err.Error() {
		case "room not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "cannot delete room with active reservations":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "room deleted successfully"})
}

func (h *RoomHandler) GetRooms(c *gin.Context) {
	// Parse pagination query parameters
	var pagination models.PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination parameters"})
		return
	}

	// Parse filter from request body (if provided)
	var filter models.RoomFilter
	if c.Request.Body != nil && c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter parameters"})
			return
		}
	}

	// Get rooms with filter and pagination
	response, err := h.service.GetRooms(&filter, &pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *RoomHandler) GetRoomSchedule(c *gin.Context) {
	// Parse room ID from URL
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID format"})
		return
	}

	// Parse and validate query parameters
	var query models.RoomScheduleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid query parameters: %v", err)})
		return
	}

	// Validate time range
	if query.StartDateTime.After(query.EndDateTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_datetime must be before end_datetime"})
		return
	}

	// Get room schedule from service
	response, err := h.service.GetRoomSchedule(roomID, &query)
	if err != nil {
		if err.Error() == "room not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error fetching room schedule: %v", err)})
		return
	}

	c.JSON(http.StatusOK, response)
}
