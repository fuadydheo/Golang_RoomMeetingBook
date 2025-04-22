-- Drop indexes
DROP INDEX IF EXISTS idx_reservation_snacks_snack_id;
DROP INDEX IF EXISTS idx_reservation_snacks_reservation_id;

-- Drop table
DROP TABLE IF EXISTS reservation_snacks; 