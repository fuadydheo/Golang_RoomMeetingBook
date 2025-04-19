CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    room_name VARCHAR(255) NOT NULL,
    room_type VARCHAR(100) NOT NULL,
    min_capacity INTEGER NOT NULL,
    max_capacity INTEGER NOT NULL,
    price_per_hour DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 