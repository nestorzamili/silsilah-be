-- 000001_init_schema.up.sql
-- Complete database schema for Silsilah Keluarga (Single Tree)
-- Consolidated from all migrations

-- ============================================
-- EXTENSIONS
-- ============================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For fuzzy text search

-- ============================================
-- USERS TABLE
-- ============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    person_id UUID, -- Will add FK after persons table created
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT chk_user_role CHECK (role IN ('member', 'editor', 'developer'))
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_full_name ON users USING gist(full_name gist_trgm_ops);
CREATE INDEX idx_users_role ON users(role);

-- ============================================
-- PERSONS TABLE
-- ============================================
CREATE TYPE gender_type AS ENUM ('MALE', 'FEMALE', 'UNKNOWN');

CREATE TABLE persons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Validation: death_date must be after birth_date
    CONSTRAINT chk_dates CHECK (death_date IS NULL OR birth_date IS NULL OR death_date >= birth_date)
);

CREATE INDEX idx_persons_name ON persons USING gist((first_name || ' ' || COALESCE(last_name, '')) gist_trgm_ops);
CREATE INDEX idx_persons_created_by ON persons(created_by);
CREATE INDEX idx_persons_occupation ON persons(occupation) WHERE deleted_at IS NULL;
CREATE INDEX idx_persons_religion ON persons(religion) WHERE deleted_at IS NULL;
CREATE INDEX idx_persons_alive_gender ON persons(is_alive, gender) WHERE deleted_at IS NULL;

-- Add FK from users.person_id to persons
ALTER TABLE users ADD CONSTRAINT fk_users_person_id FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE SET NULL;
CREATE UNIQUE INDEX idx_users_person_id ON users(person_id) WHERE deleted_at IS NULL AND person_id IS NOT NULL;

-- ============================================
-- RELATIONSHIPS TABLE (GRAPH EDGES)
-- ============================================
CREATE TYPE relationship_type AS ENUM ('PARENT', 'SPOUSE');

CREATE TABLE relationships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    person_a UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    person_b UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    type relationship_type NOT NULL,
    metadata JSONB DEFAULT '{}',
    spouse_order SMALLINT,
    child_order SMALLINT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT chk_no_self_relation CHECK (person_a != person_b),
    CONSTRAINT chk_spouse_order CHECK (spouse_order IS NULL OR spouse_order > 0),
    CONSTRAINT chk_child_order CHECK (child_order IS NULL OR child_order > 0)
);

CREATE UNIQUE INDEX uk_relationship_pair ON relationships (
    LEAST(person_a, person_b), 
    GREATEST(person_a, person_b), 
    type
) WHERE deleted_at IS NULL;

CREATE INDEX idx_relationships_person_a ON relationships(person_a) WHERE deleted_at IS NULL;
CREATE INDEX idx_relationships_person_b ON relationships(person_b) WHERE deleted_at IS NULL;
CREATE INDEX idx_relationships_type ON relationships(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_relationships_spouse_order ON relationships(person_a, spouse_order) WHERE type = 'SPOUSE' AND deleted_at IS NULL;
CREATE INDEX idx_relationships_child_order ON relationships(person_b, child_order) WHERE type = 'PARENT' AND deleted_at IS NULL;

COMMENT ON COLUMN relationships.spouse_order IS 'For SPOUSE type: marriage order (1=first marriage, 2=second, etc.)';
COMMENT ON COLUMN relationships.child_order IS 'For PARENT type: birth order of the child from parent person_b (1=firstborn, 2=second, etc.)';

-- ============================================
-- MEDIA TABLE
-- ============================================
CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    person_id UUID REFERENCES persons(id) ON DELETE SET NULL,
    uploaded_by UUID NOT NULL REFERENCES users(id),
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

CREATE INDEX idx_media_person_id ON media(person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_status ON media(status);

COMMENT ON COLUMN media.status IS 'Media approval status: pending (awaiting approval) or active (approved/visible)';

-- ============================================
-- COMMENTS TABLE
-- ============================================
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_comments_person_id ON comments(person_id) WHERE deleted_at IS NULL;

-- ============================================
-- CHANGE REQUESTS TABLE (Approval Workflow)
-- ============================================
CREATE TYPE entity_type AS ENUM ('PERSON', 'RELATIONSHIP', 'MEDIA');
CREATE TYPE change_action AS ENUM ('CREATE', 'UPDATE', 'DELETE');
CREATE TYPE request_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED');

CREATE TABLE change_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requested_by UUID NOT NULL REFERENCES users(id),
    entity_type entity_type NOT NULL,
    entity_id UUID, -- NULL for CREATE actions
    action change_action NOT NULL,
    payload JSONB NOT NULL, -- Contains the proposed changes
    status request_status NOT NULL DEFAULT 'PENDING',
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    review_note TEXT,
    requester_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_change_requests_status ON change_requests(status) WHERE status = 'PENDING';
CREATE INDEX idx_change_requests_requested_by ON change_requests(requested_by);

-- ============================================
-- AUDIT LOGS TABLE (Append-Only)
-- ============================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- ============================================
-- NOTIFICATIONS TABLE
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    data JSONB DEFAULT '{}',
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- ============================================
-- SESSIONS TABLE
-- ============================================
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address INET,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash) WHERE revoked_at IS NULL;

-- ============================================
-- FUNCTIONS & TRIGGERS
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to relevant tables
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_persons_updated_at
    BEFORE UPDATE ON persons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_relationships_updated_at
    BEFORE UPDATE ON relationships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_comments_updated_at
    BEFORE UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_change_requests_updated_at
    BEFORE UPDATE ON change_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- USER PERMISSION FUNCTION
-- ============================================
CREATE OR REPLACE FUNCTION check_user_permission(user_id UUID, required_role VARCHAR)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM users 
        WHERE id = user_id 
        AND role IN (
            CASE required_role
                WHEN 'developer' THEN ARRAY['developer']
                WHEN 'editor' THEN ARRAY['editor', 'developer']
                WHEN 'member' THEN ARRAY['member', 'editor', 'developer']
                ELSE ARRAY[]::VARCHAR[]
            END
        )
        AND is_active = true
        AND deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_user_permission(UUID, VARCHAR) IS 
'Check if user has required role permission. Returns true if user has sufficient privileges.';

-- ============================================
-- CONSANGUINITY FUNCTIONS
-- ============================================

-- Function to find all ancestors of a person
CREATE OR REPLACE FUNCTION get_ancestors(person_uuid UUID, max_depth INT DEFAULT 20)
RETURNS TABLE(ancestor_id UUID, depth INT) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE ancestor_tree AS (
        -- Base case: direct parents
        SELECT 
            r.person_b AS ancestor,
            1 AS level
        FROM relationships r
        WHERE r.person_a = person_uuid 
            AND r.type = 'PARENT'
            AND r.deleted_at IS NULL
        
        UNION ALL
        
        -- Recursive case: parents of parents
        SELECT 
            r.person_b AS ancestor,
            at.level + 1 AS level
        FROM relationships r
        INNER JOIN ancestor_tree at ON r.person_a = at.ancestor
        WHERE r.type = 'PARENT'
            AND r.deleted_at IS NULL
            AND at.level < max_depth
    )
    SELECT DISTINCT ancestor, level FROM ancestor_tree;
END;
$$ LANGUAGE plpgsql;

-- Function to find all descendants of a person
CREATE OR REPLACE FUNCTION get_descendants(person_uuid UUID, max_depth INT DEFAULT 20)
RETURNS TABLE(descendant_id UUID, depth INT) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE descendant_tree AS (
        -- Base case: direct children
        SELECT 
            r.person_a AS descendant,
            1 AS level
        FROM relationships r
        WHERE r.person_b = person_uuid 
            AND r.type = 'PARENT'
            AND r.deleted_at IS NULL
        
        UNION ALL
        
        -- Recursive case: children of children
        SELECT 
            r.person_a AS descendant,
            dt.level + 1 AS level
        FROM relationships r
        INNER JOIN descendant_tree dt ON r.person_b = dt.descendant
        WHERE r.type = 'PARENT'
            AND r.deleted_at IS NULL
            AND dt.level < max_depth
    )
    SELECT DISTINCT descendant, level FROM descendant_tree;
END;
$$ LANGUAGE plpgsql;

-- Function to find common ancestors between two persons
CREATE OR REPLACE FUNCTION find_common_ancestors(person_a_uuid UUID, person_b_uuid UUID)
RETURNS TABLE(
    common_ancestor_id UUID, 
    depth_from_a INT, 
    depth_from_b INT,
    total_degree INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a1.ancestor_id AS common_ancestor_id,
        a1.depth AS depth_from_a,
        a2.depth AS depth_from_b,
        (a1.depth + a2.depth) AS total_degree
    FROM get_ancestors(person_a_uuid) a1
    INNER JOIN get_ancestors(person_b_uuid) a2 ON a1.ancestor_id = a2.ancestor_id
    ORDER BY total_degree ASC;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate consanguinity degree between two persons
-- Returns NULL if not blood related, otherwise returns the degree
CREATE OR REPLACE FUNCTION calculate_consanguinity(person_a_uuid UUID, person_b_uuid UUID)
RETURNS TABLE(
    is_consanguineous BOOLEAN,
    degree INT,
    closest_common_ancestors UUID[],
    relationship_description TEXT
) AS $$
DECLARE
    min_degree INT;
    ancestors UUID[];
    rel_desc TEXT;
BEGIN
    -- Find the minimum degree (closest relationship)
    SELECT MIN(ca.total_degree), ARRAY_AGG(ca.common_ancestor_id)
    INTO min_degree, ancestors
    FROM find_common_ancestors(person_a_uuid, person_b_uuid) ca
    WHERE ca.total_degree = (
        SELECT MIN(total_degree) FROM find_common_ancestors(person_a_uuid, person_b_uuid)
    );
    
    IF min_degree IS NULL THEN
        RETURN QUERY SELECT false, NULL::INT, NULL::UUID[], 'Not blood related'::TEXT;
        RETURN;
    END IF;
    
    -- Determine relationship description based on degree
    rel_desc := CASE 
        WHEN min_degree = 2 THEN 'Siblings'
        WHEN min_degree = 3 THEN 'Uncle/Aunt - Nephew/Niece'
        WHEN min_degree = 4 THEN 'First Cousins'
        WHEN min_degree = 5 THEN 'First Cousins Once Removed'
        WHEN min_degree = 6 THEN 'Second Cousins'
        ELSE 'Distant relatives (degree: ' || min_degree || ')'
    END;
    
    RETURN QUERY SELECT true, min_degree, ancestors, rel_desc;
END;
$$ LANGUAGE plpgsql;

-- Function to get siblings of a person
CREATE OR REPLACE FUNCTION get_siblings(person_uuid UUID)
RETURNS TABLE(sibling_id UUID) AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT r2.person_a AS sibling_id
    FROM relationships r1
    INNER JOIN relationships r2 ON r1.person_b = r2.person_b
    WHERE r1.person_a = person_uuid
        AND r1.type = 'PARENT'
        AND r2.type = 'PARENT'
        AND r2.person_a != person_uuid
        AND r1.deleted_at IS NULL
        AND r2.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql;

-- Function to get spouses of a person
CREATE OR REPLACE FUNCTION get_spouses(person_uuid UUID)
RETURNS TABLE(spouse_id UUID, metadata JSONB) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        CASE 
            WHEN r.person_a = person_uuid THEN r.person_b
            ELSE r.person_a
        END AS spouse_id,
        r.metadata
    FROM relationships r
    WHERE (r.person_a = person_uuid OR r.person_b = person_uuid)
        AND r.type = 'SPOUSE'
        AND r.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql;

-- Function to get parents of a person
CREATE OR REPLACE FUNCTION get_parents(person_uuid UUID)
RETURNS TABLE(parent_id UUID) AS $$
BEGIN
    RETURN QUERY
    SELECT r.person_b AS parent_id
    FROM relationships r
    WHERE r.person_a = person_uuid
        AND r.type = 'PARENT'
        AND r.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql;

-- Function to get children of a person
CREATE OR REPLACE FUNCTION get_children(person_uuid UUID)
RETURNS TABLE(child_id UUID) AS $$
BEGIN
    RETURN QUERY
    SELECT r.person_a AS child_id
    FROM relationships r
    WHERE r.person_b = person_uuid
        AND r.type = 'PARENT'
        AND r.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql;
