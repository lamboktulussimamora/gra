-- Migration: AddUserProfiles
-- Description: Add user profiles table with additional user information
-- Created: 2025-06-07
-- Author: GRA Framework

-- Up Migration

-- User profiles table
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url VARCHAR(500),
    website_url VARCHAR(255),
    location VARCHAR(100),
    birth_date DATE,
    phone_number VARCHAR(20),
    social_links JSONB DEFAULT '{}',
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add columns to users table
ALTER TABLE users ADD COLUMN profile_completed BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN login_count INTEGER DEFAULT 0;

-- Indexes
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_location ON user_profiles(location);
CREATE INDEX idx_users_last_login_at ON users(last_login_at);
CREATE INDEX idx_users_profile_completed ON users(profile_completed);

-- Trigger for updated_at
CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to check profile completion
CREATE OR REPLACE FUNCTION check_profile_completion()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE users SET profile_completed = (
        NEW.bio IS NOT NULL AND 
        NEW.location IS NOT NULL AND 
        NEW.birth_date IS NOT NULL
    ) WHERE id = NEW.user_id;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to auto-update profile completion status
CREATE TRIGGER trigger_check_profile_completion 
    AFTER INSERT OR UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION check_profile_completion();

-- Insert sample profiles
INSERT INTO user_profiles (user_id, bio, location, social_links) VALUES
((SELECT id FROM users WHERE username = 'admin'), 
 'System administrator and GRA framework maintainer', 
 'San Francisco, CA',
 '{"github": "admin", "twitter": "@admin"}'),
((SELECT id FROM users WHERE username = 'john_doe'), 
 'Software developer passionate about Go and web development', 
 'New York, NY',
 '{"github": "johndoe", "linkedin": "john-doe"}'),
((SELECT id FROM users WHERE username = 'jane_smith'), 
 'Technical writer and developer advocate', 
 'Austin, TX',
 '{"twitter": "@janesmith", "website": "https://janesmith.dev"}');

-- Rollback Migration
-- DROP TRIGGER IF EXISTS trigger_check_profile_completion ON user_profiles;
-- DROP FUNCTION IF EXISTS check_profile_completion();
-- DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
-- ALTER TABLE users DROP COLUMN IF EXISTS profile_completed;
-- ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
-- ALTER TABLE users DROP COLUMN IF EXISTS login_count;
-- DROP TABLE IF EXISTS user_profiles CASCADE;
