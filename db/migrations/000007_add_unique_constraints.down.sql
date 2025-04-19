ALTER TABLE users 
DROP CONSTRAINT IF EXISTS users_username_unique,
DROP CONSTRAINT IF EXISTS users_email_unique; 