package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomInfo struct {
	Capacity     int     `json:"capacity"`
	PricePerHour float64 `json:"price_per_hour"`
}

type ReservationEvent struct {
	ID            uuid.UUID `json:"id"`
	RoomID        uuid.UUID `json:"room_id"`
	RoomName      string    `json:"room_name"`
	RoomDetails   RoomInfo  `json:"room_details"`
	UserID        uuid.UUID `json:"user_id"`
	Username      string    `json:"username"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	DurationHours float64   `json:"duration_hours"`
	VisitorCount  int       `json:"visitor_count"`
	Price         float64   `json:"price"`
	Status        string    `json:"status"`
}

type ReservationHistoryQuery struct {
	StartDatetime string    `form:"start_datetime"`
	EndDatetime   string    `form:"end_datetime"`
	RoomTypeID    uuid.UUID `form:"room_type_id"`
	Status        string    `form:"status"`
	Page          int       `form:"page" binding:"min=1"`
	PageSize      int       `form:"page_size" binding:"min=1,max=100"`
}

type ReservationHistoryResponse struct {
	StartDatetime time.Time          `json:"start_datetime"`
	EndDatetime   time.Time          `json:"end_datetime"`
	Page          int                `json:"page"`
	PageSize      int                `json:"page_size"`
	TotalItems    int                `json:"total_items"`
	TotalPages    int                `json:"total_pages"`
	Events        []ReservationEvent `json:"events"`
}

type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
	ReservationStatusCompleted ReservationStatus = "completed"
)

func (s ReservationStatus) IsValid() bool {
	switch s {
	case ReservationStatusPending, ReservationStatusConfirmed,
		ReservationStatusCancelled, ReservationStatusCompleted:
		return true
	}
	return false
}

type UpdateReservationStatusRequest struct {
	ReservationID uuid.UUID         `json:"reservation_id" binding:"required"`
	Status        ReservationStatus `json:"status" binding:"required"`
}

type ReservationCalculationRequest struct {
	RoomID uuid.UUID `json:"room_id" binding:"required"`
	Snacks []struct {
		SnackID  uuid.UUID `json:"snack_id" binding:"required"`
		Quantity int       `json:"quantity" binding:"required,min=1"`
	} `json:"snacks" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

type ReservationCalculationResponse struct {
	Room struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		PricePerHour float64   `json:"price_per_hour"`
		TotalHours   float64   `json:"total_hours"`
		TotalCost    float64   `json:"total_cost"`
	} `json:"room"`
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

type CreateReservationRequest struct {
	RoomID       uuid.UUID `json:"room_id" binding:"required"`
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	StartTime    time.Time `json:"start_time" binding:"required"`
	EndTime      time.Time `json:"end_time" binding:"required"`
	VisitorCount int       `json:"visitor_count" binding:"required,min=1"`
	Snacks       []struct {
		SnackID  uuid.UUID `json:"snack_id" binding:"required"`
		Quantity int       `json:"quantity" binding:"required,min=1"`
	} `json:"snacks" binding:"required"`
}

type CreateReservationResponse struct {
	ReservationID uuid.UUID `json:"reservation_id"`
	Status        string    `json:"status"`
	TotalCost     float64   `json:"total_cost"`
	CreatedAt     time.Time `json:"created_at"`
}
