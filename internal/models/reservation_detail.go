package models

import (
	"time"

	"github.com/google/uuid"
)

// ReservationDetailResponse represents the detailed information of a reservation
// including room, user, and snack details
type ReservationDetailResponse struct {
	ID           uuid.UUID `json:"id"`
	Status       string    `json:"status"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	VisitorCount int       `json:"visitor_count"`
	Price        float64   `json:"price"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Room struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		Capacity     int       `json:"capacity"`
		PricePerHour float64   `json:"price_per_hour"`
	} `json:"room"`

	User struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
	} `json:"user"`

	Snacks []struct {
		ID       uuid.UUID `json:"id"`
		Name     string    `json:"name"`
		Category string    `json:"category"`
		Price    float64   `json:"price"`
		Quantity int       `json:"quantity"`
		Subtotal float64   `json:"subtotal"`
	} `json:"snacks"`

	TotalCost float64 `json:"total_cost"`
}