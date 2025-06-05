-- Migration: create_initial_schema
-- Created: 2025-06-05T23:37:24+07:00
-- Checksum: 4f8033f36270529bf4f139c38c2daf01
-- Mode: GenerateOnly
-- Has Destructive: false
-- Requires Review: false

-- +migrate Up
-- Migration Up Script
-- Generated at: 2025-06-05T23:37:24+07:00
-- Changes: 3

-- Create Tables (3)

CREATE TABLE "categories" (
    children TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    description TEXT NOT NULL,
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    parent TEXT,
    parent_id INTEGER,
    products TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE "products" (
    category TEXT,
    category_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    description TEXT NOT NULL,
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    in_stock INTEGER NOT NULL,
    name VARCHAR(200) NOT NULL,
    order_items TEXT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    reviews TEXT NOT NULL,
    sku VARCHAR(100) NOT NULL,
    stock_count INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE "users" (
    created_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    is_active INTEGER NOT NULL,
    last_login TIMESTAMP,
    last_name VARCHAR(50) NOT NULL,
    orders TEXT NOT NULL,
    password VARCHAR(255) NOT NULL,
    reviews TEXT NOT NULL,
    roles TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +migrate Down
-- Migration Down Script
-- Generated at: 2025-06-05T23:37:24+07:00
-- Reverses changes from up script

-- Drop Tables (3)

DROP TABLE IF EXISTS "categories";

DROP TABLE IF EXISTS "products";

DROP TABLE IF EXISTS "users";

