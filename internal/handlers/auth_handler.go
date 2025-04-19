package handlers

import (
	"e-meetingproject/internal/models"
	"e-meetingproject/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
// @Summary Register new user
// @Description Register a new user with username, email, and password
// @Accept json
// @Produce json
// @Param register body models.RegisterRequest true "Registration details"
// @Success 201 {object} models.RegisterResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Additional validation
	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		switch err.Error() {
		case "username already exists", "email already exists":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Accept json
// @Produce json
// @Param login body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq models.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		fmt.Printf("Invalid request format: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	fmt.Printf("Login attempt for username: %s\n", loginReq.Username)

	// Validate input
	if loginReq.Username == "" || loginReq.Password == "" {
		fmt.Println("Empty username or password")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	// Attempt login
	response, err := h.authService.Login(loginReq.Username, loginReq.Password)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Send a password reset link to the user's email
// @Accept json
// @Produce json
// @Param request body models.PasswordResetRequest true "Password reset request"
// @Success 200 {object} models.PasswordResetResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /password/reset_request [post]
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req models.PasswordResetRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	response, err := h.authService.RequestPasswordReset(req.Email)
	if err != nil {
		fmt.Printf("Error processing password reset request: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ResetPassword godoc
// @Summary Reset password using token
// @Description Reset user's password using a valid reset token
// @Accept json
// @Produce json
// @Param token query string false "Reset token (can also be provided in request body)"
// @Param request body models.PasswordResetConfirmRequest true "Password reset confirmation"
// @Success 200 {object} models.PasswordResetConfirmResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /password/reset [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.PasswordResetConfirmRequest

	// Try to get token from query parameter first
	token := c.Query("token")
	if token != "" {
		req.Token = token
	}

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate token presence
	if req.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reset token is required"})
		return
	}

	// Validate password match
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
		return
	}

	response, err := h.authService.ResetPassword(&req)
	if err != nil {
		switch err.Error() {
		case "invalid or expired reset token", "reset token has expired":
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			fmt.Printf("Error resetting password: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}
