CREATE TABLE snacks (
    id SERIAL PRIMARY KEY,
    snack_name VARCHAR(255) NOT NULL,
    snack_category VARCHAR(100) NOT NULL,
    price_per_unit DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 