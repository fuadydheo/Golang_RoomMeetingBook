-- Remove the role constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS valid_role;

-- Change role back to varchar
ALTER TABLE users ALTER COLUMN role TYPE varchar USING role::varchar;

-- Drop the role enum type
DROP TYPE IF EXISTS user_role; 