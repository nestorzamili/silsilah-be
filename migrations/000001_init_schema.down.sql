-- 000001_init_schema.down.sql
-- Complete rollback of Silsilah Keluarga database schema
-- Consolidated from all migrations

-- ============================================
-- DROP CONSANGUINITY FUNCTIONS
-- ============================================
DROP FUNCTION IF EXISTS get_children(UUID);
DROP FUNCTION IF EXISTS get_parents(UUID);
DROP FUNCTION IF EXISTS get_spouses(UUID);
DROP FUNCTION IF EXISTS get_siblings(UUID);
DROP FUNCTION IF EXISTS calculate_consanguinity(UUID, UUID);
DROP FUNCTION IF EXISTS find_common_ancestors(UUID, UUID);
DROP FUNCTION IF EXISTS get_descendants(UUID, INT);
DROP FUNCTION IF EXISTS get_ancestors(UUID, INT);

-- ============================================
-- DROP USER PERMISSION FUNCTION
-- ============================================
DROP FUNCTION IF EXISTS check_user_permission(UUID, VARCHAR);

-- ============================================
-- DROP TRIGGERS
-- ============================================
DROP TRIGGER IF EXISTS trg_change_requests_updated_at ON change_requests;
DROP TRIGGER IF EXISTS trg_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS trg_relationships_updated_at ON relationships;
DROP TRIGGER IF EXISTS trg_persons_updated_at ON persons;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;

-- Drop updated_at function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- ============================================
-- DROP TABLES (reverse order of dependencies)
-- ============================================
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS change_requests;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS media;
DROP TABLE IF EXISTS relationships;

-- Drop users FK constraint before dropping persons
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_person_id;
DROP INDEX IF EXISTS idx_users_person_id;

DROP TABLE IF EXISTS persons;
DROP TABLE IF EXISTS users;

-- ============================================
-- DROP CUSTOM TYPES
-- ============================================
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS request_status;
DROP TYPE IF EXISTS change_action;
DROP TYPE IF EXISTS entity_type;
DROP TYPE IF EXISTS relationship_type;
DROP TYPE IF EXISTS gender_type;

-- ============================================
-- DROP EXTENSIONS
-- ============================================
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS "uuid-ossp";
