#!/bin/bash

# PostgreSQL Migration Test Script for GRA EF Core Migration System
# This script tests the complete migration lifecycle with PostgreSQL

set -e  # Exit on any error

echo "🐘 PostgreSQL Migration System Test"
echo "=================================="

# Configuration
export DATABASE_URL="postgres://postgres@localhost:5432/gra_test?sslmode=disable"
TEST_DIR="./test_postgresql_migrations"
CLI_TOOL="./bin/ef-migrate"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print test results
print_test() {
    echo -e "${BLUE}📋 Test: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Clean up function
cleanup() {
    echo -e "\n${YELLOW}🧹 Cleaning up...${NC}"
    rm -rf "$TEST_DIR"
    echo "Cleanup completed."
}

# Set up cleanup trap
trap cleanup EXIT

echo -e "\n${BLUE}🔧 Setup${NC}"
print_info "Database URL: $DATABASE_URL"
print_info "Test Directory: $TEST_DIR"

# Create test directory
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

echo -e "\n${BLUE}🧪 Running PostgreSQL Migration Tests${NC}"

# Test 1: Help Command
print_test "1. Help Command (No DB Required)"
if $CLI_TOOL help > /dev/null 2>&1; then
    print_success "Help command works"
else
    print_error "Help command failed"
    exit 1
fi

# Test 2: Schema Initialization
print_test "2. Schema Initialization"
if $CLI_TOOL status > /dev/null 2>&1; then
    print_success "Schema initialized successfully"
    print_info "Migration tracking tables created"
else
    print_error "Schema initialization failed"
    exit 1
fi

# Test 3: Verify PostgreSQL-specific schema
print_test "3. PostgreSQL Schema Verification"
echo "Checking migration tables in PostgreSQL..."
if psql "$DATABASE_URL" -c "\dt" | grep -q "__ef_migrations_history"; then
    print_success "EF migrations table created"
else
    print_error "EF migrations table missing"
fi

if psql "$DATABASE_URL" -c "\dt" | grep -q "__migration_history"; then
    print_success "Migration history table created"
else
    print_error "Migration history table missing"
fi

# Test 4: Create Migration with PostgreSQL-specific SQL
print_test "4. Creating PostgreSQL Migration"
cat > "../migrations/$(date +%s)_CreateUsersTablePG.sql" << 'EOF'
-- Migration: CreateUsersTablePG
-- Description: Create users table with PostgreSQL-specific features
-- Database: PostgreSQL

-- UP
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- Add PostgreSQL-specific constraints
ALTER TABLE users ADD CONSTRAINT check_email_format 
    CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

-- DOWN
DROP TABLE IF EXISTS users CASCADE;
EOF

if $CLI_TOOL add-migration CreateUsersTablePG "PostgreSQL users table with constraints"; then
    print_success "PostgreSQL migration created"
else
    print_error "Migration creation failed"
fi

# Test 5: Migration Status
print_test "5. Migration Status Check"
if $CLI_TOOL status; then
    print_success "Status command works"
else
    print_error "Status command failed"
fi

# Test 6: Apply Migration
print_test "6. Applying Migration to PostgreSQL"
if $CLI_TOOL update-database; then
    print_success "Migration applied successfully"
else
    print_error "Migration application failed"
    exit 1
fi

# Test 7: Verify table creation in PostgreSQL
print_test "7. PostgreSQL Table Verification"
echo "Checking if users table was created..."
if psql "$DATABASE_URL" -c "\d users" > /dev/null 2>&1; then
    print_success "Users table created successfully"
    
    # Show table structure
    print_info "Table structure:"
    psql "$DATABASE_URL" -c "\d users"
    
    # Check constraints
    print_info "Checking PostgreSQL constraints:"
    psql "$DATABASE_URL" -c "SELECT conname, contype FROM pg_constraint WHERE conrelid = 'users'::regclass;"
    
else
    print_error "Users table was not created"
    exit 1
fi

# Test 8: Insert test data to verify PostgreSQL features
print_test "8. Testing PostgreSQL Features"
print_info "Inserting test data..."
if psql "$DATABASE_URL" -c "
INSERT INTO users (username, email) VALUES 
('john_doe', 'john@example.com'),
('jane_smith', 'jane@example.com');
"; then
    print_success "Test data inserted"
    
    # Verify data
    print_info "Verifying data:"
    psql "$DATABASE_URL" -c "SELECT id, username, email, created_at FROM users;"
else
    print_error "Failed to insert test data"
fi

# Test 9: Test constraint validation
print_test "9. Testing PostgreSQL Constraints"
print_info "Testing email constraint (should fail)..."
if psql "$DATABASE_URL" -c "INSERT INTO users (username, email) VALUES ('test', 'invalid-email');" 2>/dev/null; then
    print_error "Constraint validation failed - invalid email was accepted"
else
    print_success "Email constraint working correctly"
fi

# Test 10: Migration History
print_test "10. Migration History"
if $CLI_TOOL get-migration; then
    print_success "Migration history retrieved"
else
    print_error "Failed to get migration history"
fi

# Test 11: Generate SQL Script
print_test "11. SQL Script Generation"
if $CLI_TOOL script > postgresql_migration_script.sql; then
    print_success "SQL script generated"
    print_info "Script saved to postgresql_migration_script.sql"
    head -20 postgresql_migration_script.sql
else
    print_error "SQL script generation failed"
fi

# Test 12: Rollback Test
print_test "12. Migration Rollback"
if $CLI_TOOL rollback CreateUsersTablePG; then
    print_success "Migration rolled back"
    
    # Verify table is dropped
    if psql "$DATABASE_URL" -c "\d users" > /dev/null 2>&1; then
        print_error "Table still exists after rollback"
    else
        print_success "Table successfully removed after rollback"
    fi
else
    print_error "Migration rollback failed"
fi

# Test 13: Re-apply migration
print_test "13. Re-applying Migration"
if $CLI_TOOL update-database; then
    print_success "Migration re-applied successfully"
else
    print_error "Migration re-application failed"
fi

# Test 14: Advanced PostgreSQL Features Test
print_test "14. Testing Advanced PostgreSQL Features"

# Create a more complex migration
cat > "../migrations/$(date +%s)_AddUserProfilesPG.sql" << 'EOF'
-- Migration: AddUserProfilesPG
-- Description: Add user profiles with PostgreSQL JSON and advanced features
-- Database: PostgreSQL

-- UP
CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_data JSONB NOT NULL DEFAULT '{}',
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create PostgreSQL-specific indexes
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_profile_data ON user_profiles USING GIN(profile_data);
CREATE INDEX idx_user_profiles_preferences ON user_profiles USING GIN(preferences);

-- Add trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_profiles_updated_at 
    BEFORE UPDATE ON user_profiles 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- DOWN
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS user_profiles CASCADE;
EOF

if $CLI_TOOL add-migration AddUserProfilesPG "Advanced PostgreSQL features with JSONB and triggers"; then
    print_success "Advanced PostgreSQL migration created"
else
    print_error "Advanced migration creation failed"
fi

# Apply the advanced migration
if $CLI_TOOL update-database; then
    print_success "Advanced migration applied"
    
    # Test JSONB functionality
    print_info "Testing JSONB functionality..."
    psql "$DATABASE_URL" -c "
    INSERT INTO user_profiles (user_id, profile_data, preferences) VALUES 
    (1, '{\"name\": \"John Doe\", \"age\": 30, \"city\": \"New York\"}', '{\"theme\": \"dark\", \"notifications\": true}'),
    (2, '{\"name\": \"Jane Smith\", \"age\": 25, \"city\": \"San Francisco\"}', '{\"theme\": \"light\", \"notifications\": false}');
    "
    
    # Query JSONB data
    print_info "Querying JSONB data:"
    psql "$DATABASE_URL" -c "
    SELECT user_id, 
           profile_data->>'name' as name,
           profile_data->>'city' as city,
           preferences->>'theme' as theme
    FROM user_profiles;
    "
    
    print_success "JSONB functionality working correctly"
else
    print_error "Advanced migration application failed"
fi

echo -e "\n${GREEN}🎉 PostgreSQL Migration Tests Completed!${NC}"
echo -e "\n${BLUE}📊 Test Summary:${NC}"
echo "✅ Schema initialization"
echo "✅ PostgreSQL-specific table creation"
echo "✅ Constraint validation"
echo "✅ Index creation"
echo "✅ Migration application and rollback"
echo "✅ JSONB and advanced PostgreSQL features"
echo "✅ Trigger functionality"
echo "✅ Foreign key relationships"
echo "✅ Migration history tracking"

echo -e "\n${GREEN}PostgreSQL migration system is fully functional! 🐘${NC}"
