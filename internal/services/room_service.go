package services

import (
	"database/sql"
	"e-meetingproject/internal/database"
	"e-meetingproject/internal/models"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RoomService struct {
	db *sql.DB
}

func NewRoomService() *RoomService {
	return &RoomService{
		db: database.GetDB(),
	}
}

func (s *RoomService) CreateRoom(req *models.CreateRoomRequest) (*models.Room, error) {
	room := &models.Room{
		ID:           uuid.New(),
		Name:         req.Name,
		Capacity:     req.Capacity,
		PricePerHour: req.PricePerHour,
		Status:       req.Status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.db.QueryRow(`
		INSERT INTO rooms (id, name, capacity, price_per_hour, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, capacity, price_per_hour, status, created_at, updated_at`,
		room.ID, room.Name, room.Capacity, room.PricePerHour, room.Status, room.CreatedAt, room.UpdatedAt,
	).Scan(&room.ID, &room.Name, &room.Capacity, &room.PricePerHour, &room.Status, &room.CreatedAt, &room.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creating room: %v", err)
	}

	return room, nil
}

func (s *RoomService) UpdateRoom(id uuid.UUID, req *models.UpdateRoomRequest) (*models.Room, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// First, check if room exists
	var room models.Room
	err = tx.QueryRow(`
		SELECT id, name, capacity, price_per_hour, status, created_at, updated_at
		FROM rooms WHERE id = $1`,
		id,
	).Scan(&room.ID, &room.Name, &room.Capacity, &room.PricePerHour, &room.Status, &room.CreatedAt, &room.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found")
		}
		return nil, fmt.Errorf("error fetching room: %v", err)
	}

	// Update only provided fields
	if req.Name != nil {
		room.Name = *req.Name
	}
	if req.Capacity != nil {
		room.Capacity = *req.Capacity
	}
	if req.PricePerHour != nil {
		room.PricePerHour = *req.PricePerHour
	}
	if req.Status != nil {
		room.Status = *req.Status
	}
	room.UpdatedAt = time.Now()

	// Update room
	_, err = tx.Exec(`
		UPDATE rooms 
		SET name = $1, capacity = $2, price_per_hour = $3, status = $4, updated_at = $5
		WHERE id = $6`,
		room.Name, room.Capacity, room.PricePerHour, room.Status, room.UpdatedAt, room.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating room: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &room, nil
}

func (s *RoomService) DeleteRoom(id uuid.UUID) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Check if room has any reservations
	var hasReservations bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM reservations 
			WHERE room_id = $1 
			AND status NOT IN ('cancelled', 'completed')
		)`,
		id,
	).Scan(&hasReservations)

	if err != nil {
		return fmt.Errorf("error checking reservations: %v", err)
	}

	if hasReservations {
		return fmt.Errorf("cannot delete room with active reservations")
	}

	// Delete room
	result, err := tx.Exec(`DELETE FROM rooms WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error deleting room: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("room not found")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (s *RoomService) GetRooms(filter *models.RoomFilter, pagination *models.PaginationQuery) (*models.RoomListResponse, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Build query conditions
	conditions := []string{"1 = 1"} // Always true condition as a starter
	args := []interface{}{}
	argCount := 1

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argCount))
			args = append(args, "%"+*filter.Search+"%")
			argCount++
		}

		if filter.RoomTypeID != nil {
			conditions = append(conditions, fmt.Sprintf("room_type_id = $%d", argCount))
			args = append(args, *filter.RoomTypeID)
			argCount++
		}

		if filter.MinCapacity != nil {
			conditions = append(conditions, fmt.Sprintf("capacity >= $%d", argCount))
			args = append(args, *filter.MinCapacity)
			argCount++
		}

		if filter.MaxCapacity != nil {
			conditions = append(conditions, fmt.Sprintf("capacity <= $%d", argCount))
			args = append(args, *filter.MaxCapacity)
			argCount++
		}

		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
			args = append(args, *filter.Status)
			argCount++
		}
	}

	// Calculate offset
	offset := (pagination.Page - 1) * pagination.PageSize

	// Get total count
	var totalCount int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM rooms 
		WHERE %s`,
		strings.Join(conditions, " AND "),
	)

	err = tx.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("error getting total count: %v", err)
	}

	// Calculate total pages
	totalPages := (totalCount + pagination.PageSize - 1) / pagination.PageSize

	// Get rooms with pagination
	query := fmt.Sprintf(`
		SELECT id, name, capacity, price_per_hour, status, created_at, updated_at
		FROM rooms 
		WHERE %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d`,
		strings.Join(conditions, " AND "),
		argCount,
		argCount+1,
	)

	// Add pagination parameters
	args = append(args, pagination.PageSize, offset)

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying rooms: %v", err)
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Capacity,
			&room.PricePerHour,
			&room.Status,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning room: %v", err)
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rooms: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.RoomListResponse{
		Rooms:      rooms,
		TotalCount: totalCount,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *RoomService) GetRoomSchedule(roomID uuid.UUID, query *models.RoomScheduleQuery) (*models.RoomScheduleResponse, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// First, check if room exists
	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)`, roomID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("error checking room existence: %v", err)
	}
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	// Query reservations within the time range
	rows, err := tx.Query(`
		SELECT id, start_time, end_time, status, visitor_count
		FROM reservations
		WHERE room_id = $1
		AND (
			(start_time >= $2 AND start_time < $3)
			OR (end_time > $2 AND end_time <= $3)
			OR (start_time <= $2 AND end_time >= $3)
		)
		ORDER BY start_time ASC`,
		roomID, query.StartDateTime, query.EndDateTime,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying reservations: %v", err)
	}
	defer rows.Close()

	var schedules []models.RoomScheduleBlock
	for rows.Next() {
		var block models.RoomScheduleBlock
		err := rows.Scan(
			&block.ReservationID,
			&block.StartTime,
			&block.EndTime,
			&block.Status,
			&block.VisitorCount,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning reservation: %v", err)
		}
		schedules = append(schedules, block)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.RoomScheduleResponse{
		RoomID:    roomID,
		Schedules: schedules,
		StartTime: query.StartDateTime,
		EndTime:   query.EndDateTime,
	}, nil
}
