package services

import (
	"database/sql"
	"e-meetingproject/internal/database"
	"e-meetingproject/internal/models"
	"fmt"
	"time"
)

type DashboardService struct {
	db *sql.DB
}

func NewDashboardService() *DashboardService {
	return &DashboardService{
		db: database.GetDB(),
	}
}

func (s *DashboardService) GetDashboardStats(query *models.DashboardQuery) (*models.DashboardResponse, error) {
	// Parse dates
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Default to last 30 days
	var err error

	if query != nil {
		if query.StartDate != "" {
			startDate, err = time.Parse("2006-01-02", query.StartDate)
			if err != nil {
				return nil, fmt.Errorf("invalid start_date format: %v", err)
			}
		}

		if query.EndDate != "" {
			endDate, err = time.Parse("2006-01-02", query.EndDate)
			if err != nil {
				return nil, fmt.Errorf("invalid end_date format: %v", err)
			}
		}
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Get total statistics
	var totalOmzet float64
	var totalReservations, totalVisitors, totalRooms int

	err = tx.QueryRow(`
		SELECT 
			COALESCE(SUM(COALESCE(r.price, 0)), 0) as total_omzet,
			COUNT(DISTINCT r.id) as total_reservations,
			COALESCE(SUM(COALESCE(r.visitor_count, 0)), 0) as total_visitors,
			COUNT(DISTINCT rm.id) as total_rooms
		FROM rooms rm
		LEFT JOIN reservations r ON r.room_id = rm.id
			AND r.start_time >= $1 
			AND r.end_time <= $2
			AND r.status = 'confirmed'`,
		startDate, endDate,
	).Scan(&totalOmzet, &totalReservations, &totalVisitors, &totalRooms)

	if err != nil {
		return nil, fmt.Errorf("error getting total statistics: %v", err)
	}

	// Get per-room statistics
	rows, err := tx.Query(`
		WITH room_bookings AS (
			SELECT 
				rm.id as room_id,
				rm.name as room_name,
				COUNT(r.id) as total_bookings,
				COALESCE(SUM(
					EXTRACT(EPOCH FROM (
						LEAST($2, r.end_time) - 
						GREATEST($1, r.start_time)
					)) / 3600
				), 0) as total_hours,
				COALESCE(SUM(COALESCE(r.price, 0)), 0) as revenue
			FROM rooms rm
			LEFT JOIN reservations r ON r.room_id = rm.id
				AND r.start_time < $2 
				AND r.end_time > $1
				AND r.status = 'confirmed'
			GROUP BY rm.id, rm.name
		)
		SELECT 
			room_id,
			room_name,
			total_bookings,
			total_hours,
			CASE 
				WHEN $3 = 0 THEN 0
				ELSE (total_hours / ($3 * 24) * 100)
			END as occupancy_rate,
			revenue
		FROM room_bookings
		ORDER BY revenue DESC`,
		startDate, endDate,
		endDate.Sub(startDate).Hours()/24, // Total days in period
	)
	if err != nil {
		return nil, fmt.Errorf("error getting room statistics: %v", err)
	}
	defer rows.Close()

	var roomStats []models.RoomStats
	for rows.Next() {
		var stat models.RoomStats
		err := rows.Scan(
			&stat.RoomID,
			&stat.RoomName,
			&stat.TotalBookings,
			&stat.TotalHours,
			&stat.Occupancy,
			&stat.Revenue,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning room statistics: %v", err)
		}
		roomStats = append(roomStats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating room statistics: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.DashboardResponse{
		StartDate:    startDate,
		EndDate:      endDate,
		TotalOmzet:   totalOmzet,
		Reservations: totalReservations,
		Visitors:     totalVisitors,
		TotalRooms:   totalRooms,
		RoomStats:    roomStats,
	}, nil
}
