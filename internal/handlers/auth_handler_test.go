package handlers

import (
	"bytes"
	"e-meetingproject/internal/models"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of the AuthService
type MockAuthService struct {
	mock.Mock
}

// Register mocks the Register method
func (m *MockAuthService) Register(req *models.RegisterRequest) (*models.RegisterResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RegisterResponse), args.Error(1)
}

// Login mocks the Login method
func (m *MockAuthService) Login(username, password string) (*models.LoginResponse, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

// RequestPasswordReset mocks the RequestPasswordReset method
func (m *MockAuthService) RequestPasswordReset(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

// ResetPassword mocks the ResetPassword method
func (m *MockAuthService) ResetPassword(token, newPassword string) error {
	args := m.Called(token, newPassword)
	return args.Error(0)
}

func TestAuthHandler_Register(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Test cases
	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *models.RegisterResponse
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful registration",
			requestBody: models.RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			mockResponse: &models.RegisterResponse{
				Message: "User registered successfully",
				UserID:  uuid.New(),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"message": "User registered successfully",
			},
		},
		{
			name: "Password mismatch",
			requestBody: models.RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password456",
			},
			mockResponse:   nil,
			mockError:      nil, // Not called due to password mismatch
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "passwords do not match",
			},
		},
		{
			name: "Username already exists",
			requestBody: models.RegisterRequest{
				Username:        "existinguser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			mockResponse:   nil,
			mockError:      errors.New("username already exists"),
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"error": "username already exists",
			},
		},
		{
			name: "Email already exists",
			requestBody: models.RegisterRequest{
				Username:        "testuser",
				Email:           "existing@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			mockResponse:   nil,
			mockError:      errors.New("email already exists"),
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"error": "email already exists",
			},
		},
		{
			name: "Internal server error",
			requestBody: models.RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			mockResponse:   nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "internal server error",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock service
			mockService := new(MockAuthService)

			// Set up expectations
			if tc.name != "Password mismatch" {
				mockService.On("Register", mock.AnythingOfType("*models.RegisterRequest")).Return(tc.mockResponse, tc.mockError)
			}

			// Create handler with mock service
			handler := NewAuthHandler(mockService)

			// Create a test router
			router := gin.New()
			router.POST("/register", handler.Register)

			// Create request
			jsonData, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse response body
			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)

			// Assert response body
			for key, expectedValue := range tc.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}

			// Verify expectations were met
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Test cases
	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *models.LoginResponse
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful login",
			requestBody: models.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockResponse: &models.LoginResponse{
				Token: "jwt-token",
				User: models.UserInfo{
					ID:       uuid.New(),
					Username: "testuser",
					Role:     "user",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"token": "jwt-token",
			},
		},
		{
			name: "Empty credentials",
			requestBody: models.LoginRequest{
				Username: "",
				Password: "",
			},
			mockResponse:   nil,
			mockError:      nil, // Not called due to empty credentials
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Username and password are required",
			},
		},
		{
			name: "Invalid credentials",
			requestBody: models.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockResponse:   nil,
			mockError:      errors.New("invalid credentials"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid credentials",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock service
			mockService := new(MockAuthService)

			// Set up expectations
			if tc.name != "Empty credentials" {
				req := tc.requestBody.(models.LoginRequest)
				mockService.On("Login", req.Username, req.Password).Return(tc.mockResponse, tc.mockError)
			}

			// Create handler with mock service
			handler := NewAuthHandler(mockService)

			// Create a test router
			router := gin.New()
			router.POST("/login", handler.Login)

			// Create request
			jsonData, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse response body
			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)

			// Assert response body
			for key, expectedValue := range tc.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}

			// Verify expectations were met
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_RequestPasswordReset(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful password reset request",
			requestBody: map[string]string{
				"email": "test@example.com",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "If the email exists, a password reset link has been sent",
			},
		},
		{
			name: "Email not found",
			requestBody: map[string]string{
				"email": "nonexistent@example.com",
			},
			mockError:      errors.New("user not found"),
			expectedStatus: http.StatusOK, // Still return OK for security reasons
			expectedBody: map[string]interface{}{
				"message": "If the email exists, a password reset link has been sent",
			},
		},
		{
			name: "Missing email",
			requestBody: map[string]string{
				"email": "",
			},
			mockError:      nil, // Not called due to missing email
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Email is required",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock service
			mockService := new(MockAuthService)

			// Set up expectations
			if tc.requestBody["email"] != "" {
				mockService.On("RequestPasswordReset", tc.requestBody["email"]).Return(tc.mockError)
			}

			// Create handler with mock service
			handler := NewAuthHandler(mockService)

			// Create a test router
			router := gin.New()
			router.POST("/password/reset_request", handler.RequestPasswordReset)

			// Create request
			jsonData, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/password/reset_request", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse response body
			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)

			// Assert response body
			for key, expectedValue := range tc.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}

			// Verify expectations were met
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_ResetPassword(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful password reset",
			requestBody: map[string]string{
				"token":            "valid-token",
				"new_password":     "newpassword123",
				"confirm_password": "newpassword123",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Password has been reset successfully",
			},
		},
		{
			name: "Password mismatch",
			requestBody: map[string]string{
				"token":            "valid-token",
				"new_password":     "newpassword123",
				"confirm_password": "differentpassword",
			},
			mockError:      nil, // Not called due to password mismatch
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Passwords do not match",
			},
		},
		{
			name: "Invalid token",
			requestBody: map[string]string{
				"token":            "invalid-token",
				"new_password":     "newpassword123",
				"confirm_password": "newpassword123",
			},
			mockError:      errors.New("invalid or expired token"),
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid or expired token",
			},
		},
		{
			name: "Missing fields",
			requestBody: map[string]string{
				"token":            "",
				"new_password":     "",
				"confirm_password": "",
			},
			mockError:      nil, // Not called due to missing fields
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Token and new password are required",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock service
			mockService := new(MockAuthService)

			// Set up expectations
			if tc.requestBody["token"] != "" && tc.requestBody["new_password"] != "" &&
				tc.requestBody["new_password"] == tc.requestBody["confirm_password"] {
				mockService.On("ResetPassword", tc.requestBody["token"], tc.requestBody["new_password"]).Return(tc.mockError)
			}

			// Create handler with mock service
			handler := NewAuthHandler(mockService)

			// Create a test router
			router := gin.New()
			router.POST("/password/reset", handler.ResetPassword)

			// Create request
			jsonData, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/password/reset", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Parse response body
			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)

			// Assert response body
			for key, expectedValue := range tc.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}

			// Verify expectations were met
			mockService.AssertExpectations(t)
		})
	}
}
