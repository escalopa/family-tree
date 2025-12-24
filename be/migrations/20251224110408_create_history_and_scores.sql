-- +goose Up
-- +goose StatementBegin

-- =============================
-- Members History (Edit Tracking)
-- =============================
CREATE TABLE IF NOT EXISTS members_history (
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

CREATE INDEX IF NOT EXISTS idx_members_history_member_id ON members_history(member_id);
CREATE INDEX IF NOT EXISTS idx_members_history_user_id ON members_history(user_id);
CREATE INDEX IF NOT EXISTS idx_members_history_changed_at ON members_history(changed_at);

-- =============================
-- User Scores (Contribution Tracking)
-- =============================
CREATE TABLE IF NOT EXISTS user_scores (
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

CREATE INDEX IF NOT EXISTS idx_user_scores_user_id ON user_scores(user_id);
CREATE INDEX IF NOT EXISTS idx_user_scores_member_id ON user_scores(member_id);
CREATE INDEX IF NOT EXISTS idx_user_scores_user_id_member_id_field_name ON user_scores(user_id, member_id, field_name);

-- =============================
-- User Role History (Grant/Revoke Tracking)
-- =============================
CREATE TABLE IF NOT EXISTS user_role_history (
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

CREATE INDEX IF NOT EXISTS idx_user_role_history_user_id ON user_role_history(user_id);
CREATE INDEX IF NOT EXISTS idx_user_role_history_changed_by ON user_role_history(changed_by);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_role_history CASCADE;
DROP TABLE IF EXISTS user_scores CASCADE;
DROP TABLE IF EXISTS members_history CASCADE;

-- +goose StatementEnd
