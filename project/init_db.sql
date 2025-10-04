-- ============================================
-- UIT-Go Backend - Database Initialization
-- ============================================

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    password VARCHAR(255) NOT NULL,
    user_active INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to auto-update updated_at
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert test users
-- NOTE: Password is 'password123' hashed with bcrypt
-- You should generate your own hash using: https://bcrypt-generator.com/

-- Test User 1
INSERT INTO users (email, first_name, last_name, password, user_active)
VALUES (
    'admin@example.com',
    'Admin',
    'User',
    '$2a$12$jTyjFOoRBzkkndUdrx678uUEPXrtPw235PlWZ2l8jwdaI4PWBFUwS',  -- Replace with actual bcrypt hash
    1
) ON CONFLICT (email) DO NOTHING;

-- Test User 2
INSERT INTO users (email, first_name, last_name, password, user_active)
VALUES (
    'john.doe@example.com',
    'John',
    'Doe',
    '$2a$12$jTyjFOoRBzkkndUdrx678uUEPXrtPw235PlWZ2l8jwdaI4PWBFUwS',  -- Replace with actual bcrypt hash
    1
) ON CONFLICT (email) DO NOTHING;

-- Test User 3
INSERT INTO users (email, first_name, last_name, password, user_active)
VALUES (
    'jane.smith@example.com',
    'Jane',
    'Smith',
    '$2a$12$jTyjFOoRBzkkndUdrx678uUEPXrtPw235PlWZ2l8jwdaI4PWBFUwS',  -- Replace with actual bcrypt hash
    1
) ON CONFLICT (email) DO NOTHING;

-- Verify users
SELECT id, email, first_name, last_name, user_active, created_at 
FROM users;

-- ============================================
-- How to use this file:
-- ============================================
-- 1. Start the Docker containers:
--    docker-compose up -d
--
-- 2. Run this SQL file:
--    docker exec -i uit-go-postgres-1 psql -U postgres -d users < init_db.sql
--
-- Or manually:
--    docker exec -it uit-go-postgres-1 psql -U postgres -d users
--    Then paste the contents of this file
--
-- ============================================
-- Generate bcrypt hash for your password:
-- ============================================
-- Online: https://bcrypt-generator.com/
-- Or use Go:
--
-- package main
-- import (
--     "fmt"
--     "golang.org/x/crypto/bcrypt"
-- )
-- func main() {
--     password := "password123"
--     hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
--     fmt.Println(string(hash))
-- }
-- ============================================
