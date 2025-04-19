package handlers

import (
	"e-meetingproject/internal/models"
	"e-meetingproject/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get user profile information by ID
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} models.UserProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get authenticated user ID from context
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	requestedID := c.Param("id")

	// Check if user is requesting their own profile
	if authUserID.(string) != requestedID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":                 "access denied",
			"message":               "You can only view your own profile. The requested profile ID does not match your authenticated user ID.",
			"authenticated_user_id": authUserID,
			"requested_profile_id":  requestedID,
		})
		return
	}

	profile, err := h.userService.GetProfile(requestedID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case "invalid user ID format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		default:
			fmt.Printf("Error fetching user profile: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile information
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.UpdateProfileRequest true "Profile update details"
// @Security BearerAuth
// @Success 200 {object} models.UserProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /users/{id} [post]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get authenticated user ID from context
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	requestedID := c.Param("id")

	// Check if user is updating their own profile
	if authUserID.(string) != requestedID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":                 "access denied",
			"message":               "You can only update your own profile. The requested profile ID does not match your authenticated user ID.",
			"authenticated_user_id": authUserID,
			"requested_profile_id":  requestedID,
		})
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.userService.UpdateProfile(requestedID, &req)
	if err != nil {
		switch err.Error() {
		case "username already taken", "email already taken":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case "invalid user ID format":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			fmt.Printf("Error updating profile: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, profile)
}
