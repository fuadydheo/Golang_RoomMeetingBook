package handlers

import (
	"e-meetingproject/internal/models"
	"e-meetingproject/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get statistics about reservations, visitors, rooms, and revenue
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.DashboardResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /dashboard [get]
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	var query models.DashboardQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.dashboardService.GetDashboardStats(&query)
	if err != nil {
		if err.Error() == "invalid start_date format" || err.Error() == "invalid end_date format" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching dashboard statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
