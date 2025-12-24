-- +goose Up
-- +goose StatementBegin

-- =============================
-- Members (Family Tree)
-- =============================
CREATE TABLE IF NOT EXISTS members (
    member_id SERIAL,
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

CREATE INDEX IF NOT EXISTS idx_members_father_id ON members(father_id);
CREATE INDEX IF NOT EXISTS idx_members_mother_id ON members(mother_id);
CREATE INDEX IF NOT EXISTS idx_members_deleted_at ON members(deleted_at);

-- =============================
-- Members Marriages (Spouse Relationships)
-- =============================
CREATE TABLE IF NOT EXISTS members_spouse (
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
    ADD CONSTRAINT fk_marriage_father FOREIGN KEY (father_id) REFERENCES members(member_id),
    ADD CONSTRAINT fk_marriage_mother FOREIGN KEY (mother_id) REFERENCES members(member_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_members_spouse_pair ON members_spouse(father_id, mother_id);

CREATE INDEX IF NOT EXISTS idx_members_spouse_father ON members_spouse(father_id);
CREATE INDEX IF NOT EXISTS idx_members_spouse_mother ON members_spouse(mother_id);
CREATE INDEX IF NOT EXISTS idx_members_spouse_deleted_at ON members_spouse(deleted_at) WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS members_spouse CASCADE;
DROP TABLE IF EXISTS members CASCADE;

-- +goose StatementEnd
