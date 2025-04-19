-- Password is 'password123'
INSERT INTO users (username, email, password, role, status)
VALUES (
    'testuser',
    'test@example.com',
    '$2a$10$zXEQLTI6h3YvFYHzWzWYFOkgkRJw3Y6.eC7giKxZT9sT5FG3dFqCi',
    'user',
    'active'
); 