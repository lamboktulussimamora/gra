-- Migration: CreateUsersTable
-- Description: Initial user table with authentication
-- Created: 2025-06-01 12:41:32
-- Version: 1748756492

-- UP Migration
-- Migration: CreateUsersTable
-- Description: Initial user table with authentication
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- DOWN Migration (for rollback)
-- Rollback for: CreateUsersTable
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;


