-- +goose Up
-- +goose StatementBegin

-- =============================
-- Languages Table
-- =============================
CREATE TABLE IF NOT EXISTS languages (
    language_code VARCHAR(10) PRIMARY KEY,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    display_order INT DEFAULT 0 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_languages_display_order ON languages(display_order);
CREATE INDEX IF NOT EXISTS idx_languages_is_active ON languages(is_active);

-- =============================
-- Member Names Table
-- =============================
CREATE TABLE IF NOT EXISTS member_names (
    member_name_id SERIAL PRIMARY KEY,
    member_id INT NOT NULL,
    language_code VARCHAR(10) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE member_names
    ADD CONSTRAINT fk_member_names_member FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_member_names_language FOREIGN KEY (language_code) REFERENCES languages(language_code);

CREATE UNIQUE INDEX IF NOT EXISTS uq_member_names_member_language ON member_names(member_id, language_code);

CREATE INDEX IF NOT EXISTS idx_member_names_member_id ON member_names(member_id);
CREATE INDEX IF NOT EXISTS idx_member_names_language_code ON member_names(language_code);
CREATE INDEX IF NOT EXISTS idx_member_names_name ON member_names(name);

-- =============================
-- User Language Preferences Table
-- =============================
CREATE TABLE IF NOT EXISTS user_language_preferences (
    user_id INT PRIMARY KEY,
    preferred_language VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE user_language_preferences
    ADD CONSTRAINT fk_user_lang_pref_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_user_lang_pref_preferred FOREIGN KEY (preferred_language) REFERENCES languages(language_code);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_language_preferences CASCADE;
DROP TABLE IF EXISTS member_names CASCADE;
DROP TABLE IF EXISTS languages CASCADE;

-- +goose StatementEnd
