package models

import (
	"time"

	"github.com/google/uuid"
)

type Snack struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SnackListResponse struct {
	Snacks []Snack `json:"snacks"`
}
