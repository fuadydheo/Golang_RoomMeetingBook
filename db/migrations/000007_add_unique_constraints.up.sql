-- Add unique constraints if they don't exist
ALTER TABLE users 
ADD CONSTRAINT users_username_unique UNIQUE (username),
ADD CONSTRAINT users_email_unique UNIQUE (email); 