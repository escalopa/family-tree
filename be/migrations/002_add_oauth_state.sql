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


