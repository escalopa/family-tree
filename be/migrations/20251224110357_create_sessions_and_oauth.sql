-- +goose Up
-- +goose StatementBegin

-- =============================
-- User Sessions
-- =============================
CREATE TABLE IF NOT EXISTS user_sessions (
    session_id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id INT NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE NOT NULL
);

ALTER TABLE user_sessions
    ADD CONSTRAINT pk_user_sessions PRIMARY KEY (session_id),
    ADD CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users(user_id);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id_session_id ON user_sessions(user_id, session_id);

-- =============================
-- OAuth State Table
-- =============================
CREATE TABLE IF NOT EXISTS oauth_states (
    state VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE NOT NULL
);

ALTER TABLE oauth_states
    ADD CONSTRAINT pk_oauth_states PRIMARY KEY (state);

CREATE INDEX IF NOT EXISTS idx_oauth_states_expires_at ON oauth_states(expires_at);
CREATE INDEX IF NOT EXISTS idx_oauth_states_provider ON oauth_states(provider);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS oauth_states CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;

-- +goose StatementEnd
