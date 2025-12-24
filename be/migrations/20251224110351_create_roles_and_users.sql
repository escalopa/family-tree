-- +goose Up
-- +goose StatementBegin

-- =============================
-- Extensions
-- =============================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================
-- Roles
-- =============================
CREATE TABLE IF NOT EXISTS roles (
    role_id INT NOT NULL,
    name TEXT NOT NULL
);

ALTER TABLE roles
    ADD CONSTRAINT pk_roles PRIMARY KEY (role_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_roles_name ON roles(name);

-- Insert default roles
INSERT INTO roles (role_id, name)
VALUES
    (100, 'none'),
    (200, 'guest'),
    (300, 'admin'),
    (400, 'super_admin');

-- =============================
-- Users
-- =============================
CREATE TABLE IF NOT EXISTS users (
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
    ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(role_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email ON users(email);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP EXTENSION IF EXISTS "uuid-ossp";

-- +goose StatementEnd
