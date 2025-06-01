-- Migration: CreateUsersTable
-- Description: Initial user table with authentication
-- Created: 2025-06-01 12:41:32
-- Version: 1748756492

-- UP Migration
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- DOWN Migration (for rollback)
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;


