package services

import (
	"database/sql"
	"e-meetingproject/internal/database"
	"e-meetingproject/internal/models"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db *sql.DB
}

func NewUserService() *UserService {
	return &UserService{
		db: database.GetDB(),
	}
}

func (s *UserService) GetProfile(userID string) (*models.UserProfileResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}

	var profile models.UserProfileResponse
	err = s.db.QueryRow(`
		SELECT id, username, email, role, status, language, profpic, created_at, updated_at
		FROM users
		WHERE id = $1`,
		id,
	).Scan(
		&profile.ID,
		&profile.Username,
		&profile.Email,
		&profile.Role,
		&profile.Status,
		&profile.Language,
		&profile.ProfPic,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("error fetching user profile: %v", err)
	}

	return &profile, nil
}

func (s *UserService) UpdateProfile(userID string, req *models.UpdateProfileRequest) (*models.UserProfileResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Check if username is already taken by another user
	var count int
	err = tx.QueryRow(`
		SELECT COUNT(*) 
		FROM users 
		WHERE username = $1 AND id != $2`,
		req.Username, id,
	).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("error checking username uniqueness: %v", err)
	}
	if count > 0 {
		return nil, errors.New("username already taken")
	}

	// Check if email is already taken by another user
	err = tx.QueryRow(`
		SELECT COUNT(*) 
		FROM users 
		WHERE email = $1 AND id != $2`,
		req.Email, id,
	).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("error checking email uniqueness: %v", err)
	}
	if count > 0 {
		return nil, errors.New("email already taken")
	}

	// Build update query
	query := `
		UPDATE users 
		SET username = $1, 
			email = $2, 
			language = $3, 
			updated_at = $4`
	args := []interface{}{
		req.Username,
		req.Email,
		req.Language,
		time.Now(),
	}
	argCount := 5

	// Add password update if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("error hashing password: %v", err)
		}
		query += fmt.Sprintf(", password = $%d", argCount)
		args = append(args, hashedPassword)
		argCount++
	}

	// Add WHERE clause
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, username, email, role, status, language, profpic, created_at, updated_at", argCount)
	args = append(args, id)

	// Execute update and scan result
	var profile models.UserProfileResponse
	err = tx.QueryRow(query, args...).Scan(
		&profile.ID,
		&profile.Username,
		&profile.Email,
		&profile.Role,
		&profile.Status,
		&profile.Language,
		&profile.ProfPic,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &profile, nil
}
