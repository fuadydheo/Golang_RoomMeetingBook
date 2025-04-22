package services

import (
	"database/sql"
	"e-meetingproject/internal/database"
	"e-meetingproject/internal/models"
	"fmt"
)

type SnackService struct {
	db *sql.DB
}

func NewSnackService() *SnackService {
	return &SnackService{
		db: database.GetDB(),
	}
}

func (s *SnackService) GetSnacks() (*models.SnackListResponse, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Query all snacks
	rows, err := tx.Query(`
		SELECT id, name, category, price, created_at, updated_at
		FROM snacks
		ORDER BY category, name
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying snacks: %v", err)
	}
	defer rows.Close()

	var snacks []models.Snack
	for rows.Next() {
		var snack models.Snack
		err := rows.Scan(
			&snack.ID,
			&snack.Name,
			&snack.Category,
			&snack.Price,
			&snack.CreatedAt,
			&snack.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning snack: %v", err)
		}
		snacks = append(snacks, snack)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snacks: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.SnackListResponse{
		Snacks: snacks,
	}, nil
}
