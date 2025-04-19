-- First, drop foreign key constraints
ALTER TABLE password_reset_tokens DROP CONSTRAINT IF EXISTS password_reset_tokens_user_id_fkey;
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS reservations_user_id_fkey;

-- Modify users table back to SERIAL
ALTER TABLE users 
    DROP CONSTRAINT users_pkey,
    ALTER COLUMN id SET DATA TYPE INTEGER USING (floor(random() * 1000000)::integer),
    ADD PRIMARY KEY (id);

-- Update the foreign key columns in related tables
ALTER TABLE password_reset_tokens 
    ALTER COLUMN user_id SET DATA TYPE INTEGER USING (floor(random() * 1000000)::integer),
    ADD CONSTRAINT password_reset_tokens_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE reservations 
    ALTER COLUMN user_id SET DATA TYPE INTEGER USING (floor(random() * 1000000)::integer),
    ADD CONSTRAINT reservations_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(id);

-- Drop UUID extension if desired
-- DROP EXTENSION IF EXISTS "uuid-ossp"; 