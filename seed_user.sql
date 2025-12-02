DELETE FROM users WHERE email = 'admin@example.com';
INSERT INTO users (email, password_hash, role) VALUES ('admin@example.com', '$2a$10$9ZDGHipbM.6EuRAlAdLO8uGDMBVLKQZ8OzJw4rWw9WJBzn47APtDO', 'admin');
