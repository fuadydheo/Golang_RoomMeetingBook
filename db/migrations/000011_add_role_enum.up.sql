-- Drop existing constraint if it exists
ALTER TABLE users DROP CONSTRAINT IF EXISTS valid_role;

-- Create role enum type if it doesn't exist
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('admin', 'user');
    END IF;
END $$;

-- Drop the existing default
ALTER TABLE users ALTER COLUMN role DROP DEFAULT;

-- Alter users table to use the new role enum
ALTER TABLE users 
    ALTER COLUMN role TYPE user_role USING CASE 
        WHEN role = 'admin' THEN 'admin'::user_role 
        ELSE 'user'::user_role 
    END;

-- Set the default value
ALTER TABLE users 
    ALTER COLUMN role SET DEFAULT 'user'::user_role;

-- Add constraint to ensure role is either 'admin' or 'user'
ALTER TABLE users 
    ADD CONSTRAINT valid_role CHECK (role IN ('admin', 'user')); 