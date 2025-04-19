package database

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// SeedUsers creates initial user data in the database
func SeedUsers() error {
	// Default users to be seeded
	defaultUsers := []struct {
		username string
		email    string
		password string
		role     string
		status   string
	}{
		{
			username: "admin",
			email:    "admin@example.com",
			password: "admin123",
			role:     "admin",
			status:   "active",
		},
		{
			username: "user1",
			email:    "user1@example.com",
			password: "password123",
			role:     "user",
			status:   "active",
		},
	}

	db := GetDB()
	if db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete related records first
	_, err = tx.Exec("DELETE FROM password_reset_tokens")
	if err != nil {
		return fmt.Errorf("failed to clean password_reset_tokens table: %v", err)
	}

	// Then delete users
	_, err = tx.Exec("DELETE FROM users")
	if err != nil {
		return fmt.Errorf("failed to clean users table: %v", err)
	}

	// Insert users
	for _, user := range defaultUsers {
		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %v", err)
		}

		// Insert user
		_, err = tx.Exec(`
			INSERT INTO users (
				id, 
				username, 
				email, 
				password, 
				role, 
				status, 
				created_at, 
				updated_at
			) VALUES ($1, $2, $3, $4, $5::user_role, $6, $7, $7)`,
			uuid.New(),
			user.username,
			user.email,
			hashedPassword,
			user.role,
			user.status,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert user %s: %v", user.username, err)
		}
		log.Printf("Seeded user: %s with role: %s", user.username, user.role)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Println("Successfully seeded users table")
	return nil
}

// SeedRooms creates initial room data in the database
func SeedRooms() error {
	// Default rooms to be seeded
	defaultRooms := []struct {
		name         string
		capacity     int
		pricePerHour float64
		status       string
	}{
		{
			name:         "Meeting Room A",
			capacity:     10,
			pricePerHour: 100000,
			status:       "available",
		},
		{
			name:         "Conference Room B",
			capacity:     20,
			pricePerHour: 200000,
			status:       "available",
		},
		{
			name:         "Board Room",
			capacity:     15,
			pricePerHour: 150000,
			status:       "available",
		},
	}

	db := GetDB()
	if db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete existing rooms
	_, err = tx.Exec("DELETE FROM rooms")
	if err != nil {
		return fmt.Errorf("failed to clean rooms table: %v", err)
	}

	// Insert rooms
	for _, room := range defaultRooms {
		_, err = tx.Exec(`
			INSERT INTO rooms (
				id,
				name,
				capacity,
				price_per_hour,
				status,
				created_at,
				updated_at
			) VALUES ($1, $2, $3, $4, $5::room_status, $6, $6)`,
			uuid.New(),
			room.name,
			room.capacity,
			room.pricePerHour,
			room.status,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert room %s: %v", room.name, err)
		}
		log.Printf("Seeded room: %s", room.name)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Println("Successfully seeded rooms table")
	return nil
}
