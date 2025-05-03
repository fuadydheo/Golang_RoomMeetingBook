package handlers

import (
	"bytes"
	"e-meetingproject/internal/models"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSnackService is a mock implementation of the SnackService
type MockSnackService struct {
	mock.Mock
}

// GetSnacks mocks the GetSnacks method
func (m *MockSnackService) GetSnacks() (*models.SnackListResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SnackListResponse), args.Error(1)
}

// CreateSnack mocks the CreateSnack method
func (m *MockSnackService) CreateSnack(req *models.CreateSnackRequest) (*models.CreateSnackResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CreateSnackResponse), args.Error(1)
}

func TestSnackHandler_GetSnacks(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Test cases
	tests := []struct {
		name           string
		mockResponse   *models.SnackListResponse
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful get snacks",
			mockResponse: &models.SnackListResponse{
				Snacks: []models.Snack{
					{
						ID:        uuid.New(),
						Name:      "Chips",
						Category:  "Savory",
						Price:     5.99,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					{
						ID:        uuid.New(),
						Name:      "Cookies",
						Category:  "Sweet",
						Price:     4.50,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"snacks": []interface{}{},
			},
		},
		{
			name:           "Database error",
			mockResponse:   nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "database error",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock service
			mockService := new(MockSnackService)

			// Set up expectations
			mockService.On("GetSnacks").Return(tc.mockResponse, tc.mockError)

			// Create handler with mock service
			handler := NewSnackHandler(mockService)

			// Create a test router
			router := gin.New()
			router.GET("/snacks", handler.GetSnacks)

			// Create request
			req, _ := http.NewRequest("GET", "/snacks", nil)

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
				assert.Contains(t, response, key)
			}

			// Verify expectations were met
			mockService.AssertExpectations(t)
		})
	}
}

func TestSnackHandler_CreateSnack(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Test cases
	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *models.CreateSnackResponse
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful snack creation",
			requestBody: models.CreateSnackRequest{
				Name:     "New Snack",
				Category: "Test Category",
				Price:    9.99,
			},
			mockResponse: &models.CreateSnackResponse{
				ID:        uuid.New(),
				Name:      "New Snack",
				Category:  "Test Category",
				Price:     9.99,
				CreatedAt: time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"name":     "New Snack",
				"category": "Test Category",
				"price":    9.99,
			},
		},
		{
			name: "Invalid request - missing fields",
			requestBody: map[string]interface{}{
				"name":  "Incomplete Snack",
				"price": 5.99,
				// Missing category
			},
			mockResponse:   nil,
			mockError:      nil, // Not called due to validation failure
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Key: 'CreateSnackRequest.Category' Error:Field validation for 'Category' failed on the 'required' tag",
			},
		},
		{
			name: "Invalid price - zero",
			requestBody: models.CreateSnackRequest{
				Name:     "Zero Price Snack",
				Category: "Test Category",
				Price:    0,
			},
			mockResponse:   nil,
			mockError:      nil, // Not called due to price validation
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "price must be greater than zero",
			},
		},
		{
			name: "Invalid price - negative",
			requestBody: models.CreateSnackRequest{
				Name:     "Negative Price Snack",
				Category: "Test Category",
				Price:    -5.99,
			},
			mockResponse:   nil,
			mockError:      nil, // Not called due to price validation
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "price must be greater than zero",
			},
		},
		{
			name: "Database error",
			requestBody: models.CreateSnackRequest{
				Name:     "Error Snack",
				Category: "Test Category",
				Price:    7.99,
			},
			mockResponse:   nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "database error",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock service
			mockService := new(MockSnackService)

			// Set up expectations for valid requests that should reach the service
			if tc.name == "Successful snack creation" || tc.name == "Database error" {
				mockService.On("CreateSnack", mock.AnythingOfType("*models.CreateSnackRequest")).Return(tc.mockResponse, tc.mockError)
			}

			// Create handler with mock service
			handler := NewSnackHandler(mockService)

			// Create a test router
			router := gin.New()
			router.POST("/snacks", handler.CreateSnack)

			// Create request
			jsonData, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/snacks", bytes.NewBuffer(jsonData))
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

			// Assert response body contains expected keys
			for key, expectedValue := range tc.expectedBody {
				if key == "error" {
					// For error messages, just check that the key exists and contains the expected text
					assert.Contains(t, response, key)
					assert.Contains(t, response[key], expectedValue)
				} else {
					assert.Contains(t, response, key)
				}
			}

			// Verify expectations were met
			mockService.AssertExpectations(t)
		})
	}
}
