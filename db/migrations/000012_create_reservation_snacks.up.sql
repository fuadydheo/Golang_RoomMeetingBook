-- Create reservation_snacks table
CREATE TABLE IF NOT EXISTS reservation_snacks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    snack_id UUID NOT NULL REFERENCES snacks(id) ON DELETE RESTRICT,
    quantity INT NOT NULL CHECK (quantity > 0),
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_reservation_snacks_reservation_id ON reservation_snacks(reservation_id);
CREATE INDEX IF NOT EXISTS idx_reservation_snacks_snack_id ON reservation_snacks(snack_id); 