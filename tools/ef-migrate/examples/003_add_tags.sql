-- Migration: AddTags
-- Description: Add tags system for posts with many-to-many relationship
-- Created: 2025-06-07
-- Author: GRA Framework

-- Up Migration

-- Tags table
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#007bff', -- hex color code
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Post tags junction table (many-to-many)
CREATE TABLE post_tags (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, tag_id)
);

-- Add tags count to posts table
ALTER TABLE posts ADD COLUMN tags_count INTEGER DEFAULT 0;

-- Indexes
CREATE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_tags_slug ON tags(slug);
CREATE INDEX idx_post_tags_post_id ON post_tags(post_id);
CREATE INDEX idx_post_tags_tag_id ON post_tags(tag_id);

-- Trigger for updated_at on tags
CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to update tags count
CREATE OR REPLACE FUNCTION update_post_tags_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE posts SET tags_count = tags_count + 1 WHERE id = NEW.post_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE posts SET tags_count = tags_count - 1 WHERE id = OLD.post_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

-- Triggers to maintain tags count
CREATE TRIGGER trigger_update_post_tags_count_insert 
    AFTER INSERT ON post_tags
    FOR EACH ROW EXECUTE FUNCTION update_post_tags_count();

CREATE TRIGGER trigger_update_post_tags_count_delete 
    AFTER DELETE ON post_tags
    FOR EACH ROW EXECUTE FUNCTION update_post_tags_count();

-- Function to generate slug from name
CREATE OR REPLACE FUNCTION generate_tag_slug()
RETURNS TRIGGER AS $$
BEGIN
    NEW.slug = lower(regexp_replace(NEW.name, '[^a-zA-Z0-9]+', '-', 'g'));
    NEW.slug = trim(both '-' from NEW.slug);
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to auto-generate slug
CREATE TRIGGER trigger_generate_tag_slug 
    BEFORE INSERT OR UPDATE ON tags
    FOR EACH ROW EXECUTE FUNCTION generate_tag_slug();

-- Insert sample tags
INSERT INTO tags (name, description, color) VALUES
('Go', 'Programming language Go', '#00ADD8'),
('Web Development', 'Web development related content', '#61DAFB'),
('Framework', 'Software frameworks and libraries', '#FF6B6B'),
('Tutorial', 'Step-by-step tutorials', '#4ECDC4'),
('Best Practices', 'Coding best practices and guidelines', '#45B7D1'),
('Performance', 'Performance optimization tips', '#96CEB4'),
('Security', 'Security and authentication topics', '#FFEAA7'),
('Database', 'Database and SQL related content', '#DDA0DD');

-- Tag the existing posts
INSERT INTO post_tags (post_id, tag_id) VALUES
((SELECT id FROM posts WHERE slug = 'welcome-to-gra-framework'), 
 (SELECT id FROM tags WHERE name = 'Framework')),
((SELECT id FROM posts WHERE slug = 'welcome-to-gra-framework'), 
 (SELECT id FROM tags WHERE name = 'Go')),
((SELECT id FROM posts WHERE slug = 'getting-started-guide'), 
 (SELECT id FROM tags WHERE name = 'Tutorial')),
((SELECT id FROM posts WHERE slug = 'getting-started-guide'), 
 (SELECT id FROM tags WHERE name = 'Framework'));

-- Update tags count for existing posts
UPDATE posts SET tags_count = (
    SELECT COUNT(*) FROM post_tags WHERE post_id = posts.id
);

-- Rollback Migration
-- DROP TRIGGER IF EXISTS trigger_generate_tag_slug ON tags;
-- DROP FUNCTION IF EXISTS generate_tag_slug();
-- DROP TRIGGER IF EXISTS trigger_update_post_tags_count_delete ON post_tags;
-- DROP TRIGGER IF EXISTS trigger_update_post_tags_count_insert ON post_tags;
-- DROP FUNCTION IF EXISTS update_post_tags_count();
-- DROP TRIGGER IF EXISTS update_tags_updated_at ON tags;
-- ALTER TABLE posts DROP COLUMN IF EXISTS tags_count;
-- DROP TABLE IF EXISTS post_tags CASCADE;
-- DROP TABLE IF EXISTS tags CASCADE;
