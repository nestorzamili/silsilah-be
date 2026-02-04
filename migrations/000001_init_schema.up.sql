-- 000001_init_schema.up.sql
-- Silsilah Keluarga (Genealogy Platform) - Consolidated Schema
-- Features: Users, Persons, Relationships (Graph), Media, Comments, Change Requests, Audit Logs, Events
-- Naming Convention: Primary keys are named {singular_table_name}_id (e.g., user_id)

-- ============================================
-- EXTENSIONS
-- ============================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For fuzzy text search

-- ============================================
-- 1. USERS TABLE
-- ============================================
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    bio TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    password_reset_token TEXT,
    password_reset_expires_at TIMESTAMP WITH TIME ZONE,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verification_token TEXT,
    email_verification_sent_at TIMESTAMP WITH TIME ZONE,
    linked_person_id UUID, -- Will add FK after persons table created
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT chk_user_role CHECK (role IN ('member', 'editor', 'developer'))
);

COMMENT ON TABLE users IS 'Registered users of the platform';
COMMENT ON COLUMN users.user_id IS 'Unique identifier for the user';
COMMENT ON COLUMN users.email IS 'User email address (login credential)';
COMMENT ON COLUMN users.linked_person_id IS 'Link to the person record in the family tree that represents this user';

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_full_name ON users USING gist(full_name gist_trgm_ops);
CREATE INDEX idx_users_role ON users(role);

-- ============================================
-- 2. PERSONS TABLE
-- ============================================
CREATE TYPE gender_type AS ENUM ('MALE', 'FEMALE', 'UNKNOWN');

CREATE TABLE persons (
    person_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    nickname VARCHAR(50),
    gender gender_type NOT NULL DEFAULT 'UNKNOWN',
    birth_date DATE,
    birth_place VARCHAR(200),
    death_date DATE,
    death_place VARCHAR(200),
    bio TEXT,
    avatar_url TEXT,
    is_alive BOOLEAN NOT NULL DEFAULT true,
    occupation VARCHAR(200),
    religion VARCHAR(50),
    nationality VARCHAR(100),
    education VARCHAR(200),
    phone VARCHAR(20),
    email VARCHAR(255),
    address TEXT,
    created_by UUID NOT NULL REFERENCES users(user_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT chk_dates CHECK (death_date IS NULL OR birth_date IS NULL OR death_date >= birth_date)
);

COMMENT ON TABLE persons IS 'Individuals in the family tree';
COMMENT ON COLUMN persons.person_id IS 'Unique identifier for the person';
COMMENT ON COLUMN persons.created_by IS 'User who created this person record';

CREATE INDEX idx_persons_name ON persons USING gist((first_name || ' ' || COALESCE(last_name, '')) gist_trgm_ops);
CREATE INDEX idx_persons_created_by ON persons(created_by);
CREATE INDEX idx_persons_occupation ON persons(occupation) WHERE deleted_at IS NULL;
CREATE INDEX idx_persons_alive_gender ON persons(is_alive, gender) WHERE deleted_at IS NULL;

-- Add FK from users.linked_person_id to persons
ALTER TABLE users ADD CONSTRAINT fk_users_person_id FOREIGN KEY (linked_person_id) REFERENCES persons(person_id) ON DELETE SET NULL;
CREATE UNIQUE INDEX idx_users_linked_person_id ON users(linked_person_id) WHERE deleted_at IS NULL AND linked_person_id IS NOT NULL;

-- ============================================
-- 3. RELATIONSHIPS TABLE (GRAPH EDGES)
-- ============================================
CREATE TYPE relationship_type AS ENUM ('PARENT', 'SPOUSE');

CREATE TABLE relationships (
    relationship_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    person_a UUID NOT NULL REFERENCES persons(person_id) ON DELETE CASCADE,
    person_b UUID NOT NULL REFERENCES persons(person_id) ON DELETE CASCADE,
    type relationship_type NOT NULL,
    role VARCHAR(50), -- e.g. 'biological', 'adopted', 'step'
    start_date DATE,
    end_date DATE,
    metadata JSONB DEFAULT '{}',
    spouse_order SMALLINT,
    child_order SMALLINT,
    created_by UUID NOT NULL REFERENCES users(user_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT chk_no_self_relation CHECK (person_a != person_b),
    CONSTRAINT chk_spouse_order CHECK (spouse_order IS NULL OR spouse_order > 0),
    CONSTRAINT chk_child_order CHECK (child_order IS NULL OR child_order > 0)
);

COMMENT ON TABLE relationships IS 'Edges connecting persons in the graph';
COMMENT ON COLUMN relationships.relationship_id IS 'Unique identifier for the relationship';
COMMENT ON COLUMN relationships.role IS 'Specific role description (e.g. Biological, Adopted)';
COMMENT ON COLUMN relationships.spouse_order IS 'For SPOUSE type: marriage order (1=first marriage, etc.)';
COMMENT ON COLUMN relationships.child_order IS 'For PARENT type: birth order of the child from person_b';

CREATE INDEX idx_relationships_pair ON relationships(person_a, person_b) WHERE deleted_at IS NULL;
CREATE INDEX idx_relationships_person_a ON relationships(person_a) WHERE deleted_at IS NULL;
CREATE INDEX idx_relationships_person_b ON relationships(person_b) WHERE deleted_at IS NULL;
CREATE INDEX idx_relationships_type ON relationships(type) WHERE deleted_at IS NULL;

-- ============================================
-- 4. EVENTS TABLE
-- ============================================
CREATE TYPE event_type AS ENUM ('BIRTH', 'DEATH', 'MARRIAGE', 'DIVORCE', 'OTHER');

CREATE TABLE events (
    event_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    person_id UUID REFERENCES persons(person_id) ON DELETE CASCADE,
    relationship_id UUID REFERENCES relationships(relationship_id) ON DELETE SET NULL,
    type event_type NOT NULL,
    title VARCHAR(100) NOT NULL,
    date DATE,
    place VARCHAR(200),
    description TEXT,
    metadata JSONB DEFAULT '{}',
    created_by UUID NOT NULL REFERENCES users(user_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

COMMENT ON TABLE events IS 'Life events associated with persons or relationships';
COMMENT ON COLUMN events.event_id IS 'Unique identifier for the event';
COMMENT ON COLUMN events.person_id IS 'The person this event belongs to (for birth, death, etc.)';
COMMENT ON COLUMN events.relationship_id IS 'The relationship this event belongs to (for marriage, divorce)';

CREATE INDEX idx_events_person ON events(person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_relationship ON events(relationship_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_date ON events(date);

-- ============================================
-- 5. MEDIA TABLE
-- ============================================
CREATE TABLE media (
    media_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    person_id UUID REFERENCES persons(person_id) ON DELETE SET NULL,
    uploaded_by UUID NOT NULL REFERENCES users(user_id),
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    storage_path TEXT NOT NULL,
    caption TEXT,
    taken_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT chk_media_status CHECK (status IN ('pending', 'active'))
);

COMMENT ON TABLE media IS 'Photos and documents';
COMMENT ON COLUMN media.media_id IS 'Unique identifier for the media file';

CREATE INDEX idx_media_person_id ON media(person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_status ON media(status);

-- ============================================
-- 6. COMMENTS TABLE
-- ============================================
CREATE TABLE comments (
    comment_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    person_id UUID NOT NULL REFERENCES persons(person_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

COMMENT ON TABLE comments IS 'User comments on person profiles';
COMMENT ON COLUMN comments.comment_id IS 'Unique identifier for the comment';

CREATE INDEX idx_comments_person_id ON comments(person_id) WHERE deleted_at IS NULL;

-- ============================================
-- 7. CHANGE REQUESTS TABLE
-- ============================================
CREATE TYPE entity_type AS ENUM ('PERSON', 'RELATIONSHIP', 'MEDIA', 'EVENT');
CREATE TYPE change_action AS ENUM ('CREATE', 'UPDATE', 'DELETE');
CREATE TYPE request_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED');

CREATE TABLE change_requests (
    request_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requested_by UUID NOT NULL REFERENCES users(user_id),
    entity_type entity_type NOT NULL,
    entity_id UUID, -- NULL for CREATE actions
    action change_action NOT NULL,
    payload JSONB NOT NULL,
    status request_status NOT NULL DEFAULT 'PENDING',
    reviewed_by UUID REFERENCES users(user_id),
    reviewed_at TIMESTAMPTZ,
    review_note TEXT,
    requester_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE change_requests IS 'Approval workflow for data changes';
COMMENT ON COLUMN change_requests.request_id IS 'Unique identifier for the change request';

CREATE INDEX idx_change_requests_status ON change_requests(status) WHERE status = 'PENDING';
CREATE INDEX idx_change_requests_requested_by ON change_requests(requested_by);

-- ============================================
-- 8. AUDIT LOGS TABLE
-- ============================================
CREATE TABLE audit_logs (
    log_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(user_id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE audit_logs IS 'Immutable history of all actions';
COMMENT ON COLUMN audit_logs.log_id IS 'Unique identifier for the log entry';

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- ============================================
-- 9. NOTIFICATIONS TABLE
-- ============================================
CREATE TYPE notification_type AS ENUM (
    'CHANGE_REQUEST',
    'CHANGE_APPROVED',
    'CHANGE_REJECTED',
    'NEW_COMMENT',
    'PERSON_ADDED',
    'RELATIONSHIP_ADDED'
);

CREATE TABLE notifications (
    notification_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    data JSONB DEFAULT '{}',
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE notifications IS 'User notifications';
COMMENT ON COLUMN notifications.notification_id IS 'Unique identifier for the notification';

CREATE INDEX idx_notifications_user_id ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- ============================================
-- 10. SESSIONS TABLE
-- ============================================
CREATE TABLE sessions (
    session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address INET,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

COMMENT ON TABLE sessions IS 'User login sessions';
COMMENT ON COLUMN sessions.session_id IS 'Unique identifier for the session';

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash) WHERE revoked_at IS NULL;

-- ============================================
-- FUNCTIONS & TRIGGERS
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER trg_persons_updated_at BEFORE UPDATE ON persons FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER trg_relationships_updated_at BEFORE UPDATE ON relationships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER trg_comments_updated_at BEFORE UPDATE ON comments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER trg_change_requests_updated_at BEFORE UPDATE ON change_requests FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER trg_events_updated_at BEFORE UPDATE ON events FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
