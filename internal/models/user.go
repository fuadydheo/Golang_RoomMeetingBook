package models

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID      `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Password  string         `json:"-"` // "-" means this field won't be included in JSON
	Role      string         `json:"role"`
	Language  sql.NullString `json:"language"`
	ProfPic   sql.NullString `json:"profpic"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Username        string `json:"username" binding:"required,min=3,max=50"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

type RegisterResponse struct {
	Message string    `json:"message"`
	UserID  uuid.UUID `json:"user_id"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetResponse struct {
	Message   string `json:"message"`
	ResetLink string `json:"reset_link"`
}

type PasswordResetToken struct {
	ID        int       `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PasswordResetConfirmRequest struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type PasswordResetConfirmResponse struct {
	Message string `json:"message"`
}

type UserProfileResponse struct {
	ID        uuid.UUID      `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Role      string         `json:"role"`
	Status    string         `json:"status"`
	Language  sql.NullString `json:"language"`
	ProfPic   sql.NullString `json:"profpic"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type UpdateProfileRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Language string `json:"language"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
