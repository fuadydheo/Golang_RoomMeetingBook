-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- First, drop existing foreign key constraints
ALTER TABLE password_reset_tokens DROP CONSTRAINT IF EXISTS password_reset_tokens_user_id_fkey;
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS reservations_user_id_fkey;

-- Modify users table
ALTER TABLE users 
    DROP CONSTRAINT users_pkey,
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id SET DATA TYPE UUID USING (uuid_generate_v4()),
    ADD PRIMARY KEY (id);

-- Update the foreign key columns in related tables
ALTER TABLE password_reset_tokens 
    ALTER COLUMN user_id SET DATA TYPE UUID USING (uuid_generate_v4()),
    ADD CONSTRAINT password_reset_tokens_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE reservations 
    ALTER COLUMN user_id SET DATA TYPE UUID USING (uuid_generate_v4()),
    ADD CONSTRAINT reservations_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(id); 