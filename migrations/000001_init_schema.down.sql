-- 000001_init_schema.down.sql

DROP TRIGGER IF EXISTS trg_events_updated_at ON events;
DROP TRIGGER IF EXISTS trg_change_requests_updated_at ON change_requests;
DROP TRIGGER IF EXISTS trg_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS trg_relationships_updated_at ON relationships;
DROP TRIGGER IF EXISTS trg_persons_updated_at ON persons;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_type;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS change_requests;
DROP TYPE IF EXISTS request_status;
DROP TYPE IF EXISTS change_action;
DROP TYPE IF EXISTS entity_type;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS media;
DROP TABLE IF EXISTS events;
DROP TYPE IF EXISTS event_type;
DROP TABLE IF EXISTS relationships;
DROP TYPE IF EXISTS relationship_type;
DROP TABLE IF EXISTS users; -- Drop users first/last depending on constraints. Actually users has FK to persons.
-- FK Loop: Users -> Persons, Persons -> Users (created_by)
-- Need to drop constraint first or use CASCADE
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_person_id;
DROP TABLE IF EXISTS persons;
DROP TYPE IF EXISTS gender_type;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "uuid-ossp";
