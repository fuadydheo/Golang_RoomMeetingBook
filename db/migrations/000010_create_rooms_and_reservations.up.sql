-- Drop existing tables if they exist
DROP TABLE IF EXISTS reservation_history;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS rooms;

-- Drop existing types if they exist
DROP TYPE IF EXISTS reservation_status;
DROP TYPE IF EXISTS room_status;

-- Create room_status enum
CREATE TYPE room_status AS ENUM ('available', 'maintenance', 'occupied');

-- Create rooms table
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    capacity INT NOT NULL,
    price_per_hour DECIMAL(10,2) NOT NULL,
    status room_status DEFAULT 'available',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create reservation_status enum
CREATE TYPE reservation_status AS ENUM ('pending', 'confirmed', 'cancelled', 'completed');

-- Create reservations table
CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL,
    user_id UUID NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    visitor_count INT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    status reservation_status DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_time_range CHECK (end_time > start_time),
    CONSTRAINT reservations_room_id_fkey FOREIGN KEY (room_id) REFERENCES rooms(id),
    CONSTRAINT reservations_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Create indexes
CREATE INDEX idx_reservations_room_id ON reservations(room_id);
CREATE INDEX idx_reservations_user_id ON reservations(user_id);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_reservations_time_range ON reservations(start_time, end_time);

-- Recreate reservation_history table
CREATE TABLE reservation_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reservation_id UUID NOT NULL REFERENCES reservations(id),
    status reservation_status NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 