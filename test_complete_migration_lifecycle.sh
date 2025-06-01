#!/bin/bash

# Complete EF Core Migration Lifecycle Test
# This script demonstrates the full migration lifecycle using the GRA EF migration system

set -e  # Exit on any error

DB_PATH="./test_migrations/complete_test.db"
export DATABASE_URL="$DB_PATH"

echo "🚀 GRA EF Core Migration Lifecycle Test"
echo "========================================"
echo "Database: $DB_PATH"
echo

# Clean start
rm -f "$DB_PATH"

echo "1️⃣  Testing Initial Status (should be empty)"
./bin/ef-migrate status

echo
echo "2️⃣  Creating Migration 1: Users Table"
./bin/ef-migrate add-migration CreateUsersTable "Create users table with authentication"

# Add real SQL to the migration file
MIGRATION_FILE=$(ls migrations/*CreateUsersTable.sql | head -1)
cat > "$MIGRATION_FILE" << 'EOF'
-- Migration: CreateUsersTable
-- Description: Create users table with authentication
-- Created: 2025-06-01 12:52:00
-- Version: auto-generated

-- UP Migration
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);

-- DOWN Migration (for rollback)
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
EOF

echo "✅ Migration file created and populated"

echo
echo "3️⃣  Checking Status Before Apply"
./bin/ef-migrate status

echo
echo "4️⃣  Applying First Migration"
./bin/ef-migrate update-database

echo
echo "5️⃣  Creating Migration 2: User Profiles"
./bin/ef-migrate add-migration AddUserProfiles "Add user profiles with additional information"

# Add SQL for second migration
MIGRATION_FILE=$(ls migrations/*AddUserProfiles.sql | tail -1)
cat > "$MIGRATION_FILE" << 'EOF'
-- Migration: AddUserProfiles
-- Description: Add user profiles with additional information
-- Created: 2025-06-01 12:52:00
-- Version: auto-generated

-- UP Migration
CREATE TABLE user_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    bio TEXT,
    avatar_url VARCHAR(500),
    date_of_birth DATE,
    location VARCHAR(100),
    website VARCHAR(200),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_profiles_location ON user_profiles(location);

-- DOWN Migration (for rollback)
DROP INDEX IF EXISTS idx_profiles_location;
DROP INDEX IF EXISTS idx_profiles_user_id;
DROP TABLE IF EXISTS user_profiles;
EOF

echo
echo "6️⃣  Creating Migration 3: User Settings"
./bin/ef-migrate add-migration AddUserSettings "Add user preference settings"

# Add SQL for third migration
MIGRATION_FILE=$(ls migrations/*AddUserSettings.sql | tail -1)
cat > "$MIGRATION_FILE" << 'EOF'
-- Migration: AddUserSettings
-- Description: Add user preference settings
-- Created: 2025-06-01 12:52:00
-- Version: auto-generated

-- UP Migration
CREATE TABLE user_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    setting_type VARCHAR(20) DEFAULT 'string',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, setting_key)
);

CREATE INDEX idx_settings_user_key ON user_settings(user_id, setting_key);
CREATE INDEX idx_settings_type ON user_settings(setting_type);

-- DOWN Migration (for rollback)
DROP INDEX IF EXISTS idx_settings_type;
DROP INDEX IF EXISTS idx_settings_user_key;
DROP TABLE IF EXISTS user_settings;
EOF

echo
echo "7️⃣  Checking All Pending Migrations"
./bin/ef-migrate get-migration

echo
echo "8️⃣  Generating SQL Script for Review"
./bin/ef-migrate script

echo
echo "9️⃣  Applying All Migrations"
./bin/ef-migrate update-database

echo
echo "🔟  Final Status Check"
./bin/ef-migrate status

echo
echo "1️⃣1️⃣  Detailed Migration History"
./bin/ef-migrate get-migration

echo
echo "1️⃣2️⃣  Testing Rollback to Second Migration"
./bin/ef-migrate rollback AddUserProfiles

echo
echo "1️⃣3️⃣  Status After Rollback"
./bin/ef-migrate status

echo
echo "1️⃣4️⃣  Re-applying Migrations"
./bin/ef-migrate update-database

echo
echo "1️⃣5️⃣  Final Database Schema Verification"
echo "Tables in database:"
sqlite3 "$DB_PATH" ".tables"

echo
echo "Migration tracking table contents:"
sqlite3 "$DB_PATH" "SELECT migration_id, product_version, applied_at FROM __ef_migrations_history ORDER BY applied_at;"

echo
echo "✅ Complete Migration Lifecycle Test PASSED!"
echo "🎉 GRA EF Core Migration System is working perfectly!"
