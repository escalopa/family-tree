-- =============================
-- Extensions
-- =============================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================
-- Roles
-- =============================
CREATE TABLE roles (
    role_id INT NOT NULL,
    name TEXT NOT NULL
);

ALTER TABLE roles
    ADD CONSTRAINT pk_roles PRIMARY KEY (role_id),
    ADD CONSTRAINT uq_roles_name UNIQUE (name);

-- Insert default roles (fixed: removed SERIAL, using explicit IDs)
INSERT INTO roles (role_id, name)
VALUES
    (100, 'none'),
    (200, 'guest'),
    (300, 'admin'),
    (400, 'super_admin');

-- =============================
-- Users
-- =============================
CREATE TABLE users (
    user_id SERIAL,
    full_name VARCHAR(255) NOT NULL,
    email TEXT NOT NULL,
    avatar TEXT,
    role_id INT,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE users
    ADD CONSTRAINT pk_users PRIMARY KEY (user_id),
    ADD CONSTRAINT uq_users_email UNIQUE (email),
    ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(role_id);

CREATE INDEX idx_users_email ON users(email);

-- =============================
-- User Sessions
-- =============================
CREATE TABLE user_sessions (
    session_id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id INT NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE NOT NULL
);

ALTER TABLE user_sessions
    ADD CONSTRAINT pk_user_sessions PRIMARY KEY (session_id),
    ADD CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users(user_id);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_user_id_session_id ON user_sessions(user_id, session_id);

-- =============================
-- OAuth State Table
-- =============================
CREATE TABLE oauth_states (
    state VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE NOT NULL
);

ALTER TABLE oauth_states
    ADD CONSTRAINT pk_oauth_states PRIMARY KEY (state);

CREATE INDEX idx_oauth_states_expires_at ON oauth_states(expires_at);
CREATE INDEX idx_oauth_states_provider ON oauth_states(provider);

-- =============================
-- Members (Family Tree)
-- =============================
CREATE TABLE members (
    member_id SERIAL,
    arabic_name VARCHAR(255) NOT NULL,
    english_name VARCHAR(255) NOT NULL,
    gender CHAR(1) NOT NULL,               -- 'M', 'F', 'N'
    picture TEXT,
    date_of_birth DATE,
    date_of_death DATE,
    father_id INT,
    mother_id INT,
    nicknames TEXT[],
    profession VARCHAR(255),
    version INT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMP                    -- Soft delete timestamp
);

ALTER TABLE members
    ADD CONSTRAINT pk_members PRIMARY KEY (member_id),
    ADD CONSTRAINT fk_members_father FOREIGN KEY (father_id) REFERENCES members(member_id),
    ADD CONSTRAINT fk_members_mother FOREIGN KEY (mother_id) REFERENCES members(member_id),
    ADD CONSTRAINT chk_members_gender CHECK (gender IN ('M', 'F', 'N'));

CREATE INDEX idx_members_father_id ON members(father_id);
CREATE INDEX idx_members_mother_id ON members(mother_id);
CREATE INDEX idx_members_arabic_name ON members(arabic_name);
CREATE INDEX idx_members_english_name ON members(english_name);
CREATE INDEX idx_members_deleted_at ON members(deleted_at);

-- =============================
-- Members Marriages (Spouse Relationships)
-- =============================
CREATE TABLE members_spouse (
    spouse_id SERIAL,
    father_id INT NOT NULL,
    mother_id INT NOT NULL,
    marriage_date DATE,
    divorce_date DATE,
    deleted_at TIMESTAMP,
    CONSTRAINT chk_marriage_dates CHECK (divorce_date IS NULL OR marriage_date IS NULL OR divorce_date >= marriage_date)
);

ALTER TABLE members_spouse
    ADD CONSTRAINT pk_members_spouse PRIMARY KEY (spouse_id),
    ADD CONSTRAINT uq_members_spouse_pair UNIQUE (father_id, mother_id),
    ADD CONSTRAINT fk_marriage_father FOREIGN KEY (father_id) REFERENCES members(member_id),
    ADD CONSTRAINT fk_marriage_mother FOREIGN KEY (mother_id) REFERENCES members(member_id);

CREATE INDEX idx_members_spouse_father ON members_spouse(father_id);
CREATE INDEX idx_members_spouse_mother ON members_spouse(mother_id);
CREATE INDEX idx_members_spouse_deleted_at ON members_spouse(deleted_at) WHERE deleted_at IS NULL;

-- =============================
-- Members History (Edit Tracking)
-- =============================
CREATE TABLE members_history (
    history_id SERIAL,
    member_id INT NOT NULL,
    user_id INT NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_type VARCHAR(50) NOT NULL,      -- e.g. INSERT, UPDATE, DELETE, ADD_SPOUSE, REMOVE_SPOUSE
    old_values JSONB,
    new_values JSONB,
    member_version INT NOT NULL DEFAULT 0
);

ALTER TABLE members_history
    ADD CONSTRAINT pk_members_history PRIMARY KEY (history_id),
    ADD CONSTRAINT fk_members_history_member FOREIGN KEY (member_id) REFERENCES members(member_id),
    ADD CONSTRAINT fk_members_history_user FOREIGN KEY (user_id) REFERENCES users(user_id);

CREATE INDEX idx_members_history_member_id ON members_history(member_id);
CREATE INDEX idx_members_history_user_id ON members_history(user_id);
CREATE INDEX idx_members_history_changed_at ON members_history(changed_at);

-- =============================
-- User Scores (Contribution Tracking)
-- =============================
CREATE TABLE user_scores (
    score_id SERIAL,
    user_id INT NOT NULL,
    member_id INT NOT NULL,
    field_name TEXT NOT NULL,
    points INT NOT NULL,
    member_version INT NOT NULL,            -- Version from members_history
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE user_scores
    ADD CONSTRAINT pk_user_scores PRIMARY KEY (score_id),
    ADD CONSTRAINT fk_user_scores_user FOREIGN KEY (user_id) REFERENCES users(user_id),
    ADD CONSTRAINT fk_user_scores_member FOREIGN KEY (member_id) REFERENCES members(member_id);

CREATE INDEX idx_user_scores_user_id ON user_scores(user_id);
CREATE INDEX idx_user_scores_member_id ON user_scores(member_id);
CREATE INDEX idx_user_scores_user_id_member_id_field_name ON user_scores(user_id, member_id, field_name);

-- =============================
-- User Role History (Grant/Revoke Tracking)
-- =============================
CREATE TABLE user_role_history (
    history_id SERIAL,
    user_id INT NOT NULL,                   -- Target user
    old_role_id INT,                        -- Previous role
    new_role_id INT,                        -- New role
    changed_by INT NOT NULL,                -- Acting admin
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    action_type VARCHAR(20) NOT NULL        -- 'GRANT' or 'REVOKE'
);

ALTER TABLE user_role_history
    ADD CONSTRAINT pk_user_role_history PRIMARY KEY (history_id),
    ADD CONSTRAINT fk_role_history_user FOREIGN KEY (user_id) REFERENCES users(user_id),
    ADD CONSTRAINT fk_role_history_old_role FOREIGN KEY (old_role_id) REFERENCES roles(role_id),
    ADD CONSTRAINT fk_role_history_new_role FOREIGN KEY (new_role_id) REFERENCES roles(role_id),
    ADD CONSTRAINT fk_role_history_changed_by FOREIGN KEY (changed_by) REFERENCES users(user_id);

CREATE INDEX idx_user_role_history_user_id ON user_role_history(user_id);
CREATE INDEX idx_user_role_history_changed_by ON user_role_history(changed_by);
