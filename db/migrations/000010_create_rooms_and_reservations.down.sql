-- Drop indexes if they exist
DROP INDEX IF EXISTS idx_reservations_time_range;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_user_id;
DROP INDEX IF EXISTS idx_reservations_room_id;

-- Drop tables if they exist
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS rooms;

-- Drop custom types if they exist
DROP TYPE IF EXISTS reservation_status;
DROP TYPE IF EXISTS room_status; 