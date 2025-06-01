-- Migration: AddUserProfiles
-- Description: Add user profiles table with PostgreSQL features
-- Created: 2025-06-01 13:18:21
-- Version: 1748758701

-- UP Migration
CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url VARCHAR(500),
    date_of_birth DATE,
    location VARCHAR(100),
    website VARCHAR(200),
    social_links JSONB DEFAULT '{}'::jsonb,
    preferences JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_profile UNIQUE(user_id)
);

-- Create indexes
CREATE INDEX idx_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_profiles_location ON user_profiles(location);
CREATE INDEX idx_profiles_social_links ON user_profiles USING GIN(social_links);
CREATE INDEX idx_profiles_preferences ON user_profiles USING GIN(preferences);

-- Add trigger for automatic timestamp updates
CREATE TRIGGER update_user_profiles_updated_at 
    BEFORE UPDATE ON user_profiles 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- DOWN Migration (for rollback)
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP TABLE IF EXISTS user_profiles CASCADE;


