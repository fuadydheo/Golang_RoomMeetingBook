package services

import (
	"crypto/rand"
	"database/sql"
	"e-meetingproject/internal/auth"
	"e-meetingproject/internal/database"
	"e-meetingproject/internal/models"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *sql.DB
}

func NewAuthService() *AuthService {
	return &AuthService{
		db: database.GetDB(),
	}
}

func (s *AuthService) Register(req *models.RegisterRequest) (*models.RegisterResponse, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}

	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert new user
	var userID uuid.UUID
	err = tx.QueryRow(`
		INSERT INTO users (id, username, email, password, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id`,
		uuid.New(), // Generate new UUID
		req.Username,
		req.Email,
		hashedPassword,
		"user",   // default role
		"active", // default status
		time.Now(),
	).Scan(&userID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				if pqErr.Constraint == "users_username_unique" {
					return nil, errors.New("username already exists")
				}
				if pqErr.Constraint == "users_email_unique" {
					return nil, errors.New("email already exists")
				}
			}
		}
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.RegisterResponse{
		Message: "User registered successfully",
		UserID:  userID,
	}, nil
}

func (s *AuthService) Login(username, password string) (*models.LoginResponse, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, password, role 
		FROM users 
		WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Create claims
	claims := &auth.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(viper.GetString("JWT_SECRET_KEY")))
	if err != nil {
		return nil, fmt.Errorf("error creating token: %v", err)
	}

	return &models.LoginResponse{
		Token: tokenString,
		User: models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
		},
	}, nil
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *AuthService) RequestPasswordReset(email string) (*models.PasswordResetResponse, error) {
	// Check if user exists
	var userID uuid.UUID
	err := s.db.QueryRow("SELECT id FROM users WHERE email = $1 AND status = 'active'", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// For security reasons, don't reveal if email exists
			return &models.PasswordResetResponse{
				Message: "If your email is registered, you will receive a password reset link",
			}, nil
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Generate reset token
	token, err := generateResetToken()
	if err != nil {
		return nil, fmt.Errorf("error generating token: %v", err)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Invalidate any existing unused tokens for this user
	_, err = tx.Exec(`
		UPDATE password_reset_tokens 
		SET used = true 
		WHERE user_id = $1 AND used = false`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("error invalidating existing tokens: %v", err)
	}

	// Get expiration time from environment variable
	expirationMinutes := viper.GetInt("RESET_LINK_EXPIRATION_MINUTES")
	if expirationMinutes == 0 {
		expirationMinutes = 15 // default to 15 minutes
	}
	expiresAt := time.Now().Add(time.Duration(expirationMinutes) * time.Minute)

	// Insert new reset token
	_, err = tx.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)`,
		userID, token, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("error storing reset token: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	// Generate reset link using base URL from environment
	baseURL := viper.GetString("RESET_LINK_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // default fallback
	}
	resetLink := fmt.Sprintf("%s/password/reset?token=%s", baseURL, token)

	// In a real application, you would send this link via email
	// For now, we'll include it in the response for testing purposes
	return &models.PasswordResetResponse{
		Message:   "Password reset link has been generated",
		ResetLink: resetLink,
	}, nil
}

func (s *AuthService) ResetPassword(req *models.PasswordResetConfirmRequest) (*models.PasswordResetConfirmResponse, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Get and validate token
	var resetToken models.PasswordResetToken
	err = tx.QueryRow(`
		SELECT id, user_id, token, expires_at, used 
		FROM password_reset_tokens 
		WHERE token = $1 AND used = false`,
		req.Token,
	).Scan(&resetToken.ID, &resetToken.UserID, &resetToken.Token, &resetToken.ExpiresAt, &resetToken.Used)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid or expired reset token")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return nil, errors.New("reset token has expired")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}

	// Update user's password
	_, err = tx.Exec(`
		UPDATE users 
		SET password = $1, updated_at = $2 
		WHERE id = $3`,
		hashedPassword,
		time.Now(),
		resetToken.UserID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating password: %v", err)
	}

	// Mark token as used
	_, err = tx.Exec(`
		UPDATE password_reset_tokens 
		SET used = true, updated_at = $1 
		WHERE id = $2`,
		time.Now(),
		resetToken.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("error invalidating reset token: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.PasswordResetConfirmResponse{
		Message: "Password has been successfully reset",
	}, nil
}
