-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS family_units (
    family_unit_id SERIAL,
    tree_id INT NOT NULL,
    relationship_type VARCHAR(32) NOT NULL DEFAULT 'marriage',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    start_date DATE,
    end_date DATE,
    source_spouse_id INT,
    legacy_father_id INT,
    legacy_mother_id INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT chk_family_units_relationship_type CHECK (relationship_type IN ('marriage', 'partnership', 'unknown')),
    CONSTRAINT chk_family_units_status CHECK (status IN ('active', 'divorced', 'separated', 'widowed', 'unknown'))
);

ALTER TABLE family_units
    ADD CONSTRAINT pk_family_units PRIMARY KEY (family_unit_id),
    ADD CONSTRAINT fk_family_units_tree FOREIGN KEY (tree_id) REFERENCES family_trees(tree_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_units_source_spouse FOREIGN KEY (source_spouse_id) REFERENCES members_spouse(spouse_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_units_legacy_father FOREIGN KEY (legacy_father_id) REFERENCES members(member_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_units_legacy_mother FOREIGN KEY (legacy_mother_id) REFERENCES members(member_id) ON DELETE CASCADE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_family_units_source_spouse_id
    ON family_units(source_spouse_id)
    WHERE source_spouse_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_family_units_legacy_parent_pair
    ON family_units(tree_id, COALESCE(legacy_father_id, 0), COALESCE(legacy_mother_id, 0))
    WHERE source_spouse_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_family_units_tree_deleted
    ON family_units(tree_id, deleted_at);

CREATE TABLE IF NOT EXISTS family_unit_partners (
    family_unit_id INT NOT NULL,
    person_id INT NOT NULL,
    partner_order INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE family_unit_partners
    ADD CONSTRAINT pk_family_unit_partners PRIMARY KEY (family_unit_id, person_id),
    ADD CONSTRAINT fk_family_unit_partners_unit FOREIGN KEY (family_unit_id) REFERENCES family_units(family_unit_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_unit_partners_person FOREIGN KEY (person_id) REFERENCES members(member_id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_family_unit_partners_person
    ON family_unit_partners(person_id);

CREATE TABLE IF NOT EXISTS family_unit_children (
    family_unit_id INT NOT NULL,
    child_person_id INT NOT NULL,
    relation_type VARCHAR(32) NOT NULL DEFAULT 'biological',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_family_unit_children_relation_type CHECK (relation_type IN ('biological', 'adopted', 'step', 'foster', 'unknown'))
);

ALTER TABLE family_unit_children
    ADD CONSTRAINT pk_family_unit_children PRIMARY KEY (family_unit_id, child_person_id),
    ADD CONSTRAINT fk_family_unit_children_unit FOREIGN KEY (family_unit_id) REFERENCES family_units(family_unit_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_unit_children_child FOREIGN KEY (child_person_id) REFERENCES members(member_id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_family_unit_children_child
    ON family_unit_children(child_person_id);

CREATE OR REPLACE FUNCTION sync_family_unit_from_spouse()
RETURNS TRIGGER AS $$
DECLARE
    unit_id INT;
    unit_tree_id INT;
BEGIN
    SELECT tree_id INTO unit_tree_id
    FROM members
    WHERE member_id = NEW.father_id;

    IF unit_tree_id IS NULL THEN
        RETURN NEW;
    END IF;

    INSERT INTO family_units (
        tree_id,
        relationship_type,
        status,
        start_date,
        end_date,
        source_spouse_id,
        legacy_father_id,
        legacy_mother_id,
        deleted_at,
        updated_at
    )
    VALUES (
        unit_tree_id,
        'marriage',
        CASE WHEN NEW.divorce_date IS NULL THEN 'active' ELSE 'divorced' END,
        NEW.marriage_date,
        NEW.divorce_date,
        NEW.spouse_id,
        NEW.father_id,
        NEW.mother_id,
        NEW.deleted_at,
        CURRENT_TIMESTAMP
    )
    ON CONFLICT (source_spouse_id)
    WHERE source_spouse_id IS NOT NULL
    DO UPDATE SET
        tree_id = EXCLUDED.tree_id,
        relationship_type = EXCLUDED.relationship_type,
        status = EXCLUDED.status,
        start_date = EXCLUDED.start_date,
        end_date = EXCLUDED.end_date,
        legacy_father_id = EXCLUDED.legacy_father_id,
        legacy_mother_id = EXCLUDED.legacy_mother_id,
        deleted_at = EXCLUDED.deleted_at,
        updated_at = CURRENT_TIMESTAMP
    RETURNING family_unit_id INTO unit_id;

    DELETE FROM family_unit_partners WHERE family_unit_id = unit_id;
    INSERT INTO family_unit_partners (family_unit_id, person_id, partner_order)
    VALUES
        (unit_id, NEW.father_id, 1),
        (unit_id, NEW.mother_id, 2)
    ON CONFLICT (family_unit_id, person_id) DO NOTHING;

    UPDATE family_unit_children
    SET family_unit_id = unit_id
    WHERE child_person_id IN (
        SELECT member_id
        FROM members
        WHERE father_id = NEW.father_id
          AND mother_id = NEW.mother_id
          AND deleted_at IS NULL
    );

    INSERT INTO family_unit_children (family_unit_id, child_person_id, relation_type)
    SELECT unit_id, member_id, 'biological'
    FROM members
    WHERE father_id = NEW.father_id
      AND mother_id = NEW.mother_id
      AND deleted_at IS NULL
    ON CONFLICT (family_unit_id, child_person_id) DO UPDATE SET
        relation_type = EXCLUDED.relation_type;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sync_family_unit_from_member_parentage()
RETURNS TRIGGER AS $$
DECLARE
    unit_id INT;
    parent_tree_id INT;
BEGIN
    IF TG_OP = 'UPDATE' THEN
        DELETE FROM family_unit_children WHERE child_person_id = OLD.member_id;
    END IF;

    IF NEW.deleted_at IS NOT NULL OR (NEW.father_id IS NULL AND NEW.mother_id IS NULL) THEN
        RETURN NEW;
    END IF;

    IF NEW.father_id IS NOT NULL AND NEW.mother_id IS NOT NULL THEN
        SELECT fu.family_unit_id INTO unit_id
        FROM family_units fu
        WHERE fu.legacy_father_id = NEW.father_id
          AND fu.legacy_mother_id = NEW.mother_id
          AND fu.deleted_at IS NULL
        ORDER BY CASE WHEN fu.source_spouse_id IS NULL THEN 1 ELSE 0 END, fu.family_unit_id
        LIMIT 1;
    END IF;

    IF unit_id IS NULL THEN
        parent_tree_id := NEW.tree_id;

        INSERT INTO family_units (
            tree_id,
            relationship_type,
            status,
            legacy_father_id,
            legacy_mother_id,
            updated_at
        )
        VALUES (
            parent_tree_id,
            'unknown',
            'unknown',
            NEW.father_id,
            NEW.mother_id,
            CURRENT_TIMESTAMP
        )
        ON CONFLICT (tree_id, COALESCE(legacy_father_id, 0), COALESCE(legacy_mother_id, 0))
        WHERE source_spouse_id IS NULL
        DO UPDATE SET
            updated_at = CURRENT_TIMESTAMP,
            deleted_at = NULL
        RETURNING family_unit_id INTO unit_id;

        IF NEW.father_id IS NOT NULL THEN
            INSERT INTO family_unit_partners (family_unit_id, person_id, partner_order)
            VALUES (unit_id, NEW.father_id, 1)
            ON CONFLICT (family_unit_id, person_id) DO NOTHING;
        END IF;

        IF NEW.mother_id IS NOT NULL THEN
            INSERT INTO family_unit_partners (family_unit_id, person_id, partner_order)
            VALUES (unit_id, NEW.mother_id, 2)
            ON CONFLICT (family_unit_id, person_id) DO NOTHING;
        END IF;
    END IF;

    INSERT INTO family_unit_children (family_unit_id, child_person_id, relation_type)
    VALUES (unit_id, NEW.member_id, 'biological')
    ON CONFLICT (family_unit_id, child_person_id) DO UPDATE SET
        relation_type = EXCLUDED.relation_type;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_family_unit_from_spouse ON members_spouse;
CREATE TRIGGER trg_sync_family_unit_from_spouse
AFTER INSERT OR UPDATE ON members_spouse
FOR EACH ROW
EXECUTE FUNCTION sync_family_unit_from_spouse();

DROP TRIGGER IF EXISTS trg_sync_family_unit_from_member_parentage ON members;
CREATE TRIGGER trg_sync_family_unit_from_member_parentage
AFTER INSERT OR UPDATE OF father_id, mother_id, tree_id, deleted_at ON members
FOR EACH ROW
EXECUTE FUNCTION sync_family_unit_from_member_parentage();

INSERT INTO family_units (
    tree_id,
    relationship_type,
    status,
    start_date,
    end_date,
    source_spouse_id,
    legacy_father_id,
    legacy_mother_id,
    deleted_at
)
SELECT father.tree_id,
       'marriage',
       CASE WHEN ms.divorce_date IS NULL THEN 'active' ELSE 'divorced' END,
       ms.marriage_date,
       ms.divorce_date,
       ms.spouse_id,
       ms.father_id,
       ms.mother_id,
       ms.deleted_at
FROM members_spouse ms
JOIN members father ON father.member_id = ms.father_id
JOIN members mother ON mother.member_id = ms.mother_id
WHERE father.deleted_at IS NULL
  AND mother.deleted_at IS NULL
ON CONFLICT (source_spouse_id)
WHERE source_spouse_id IS NOT NULL
DO UPDATE SET
    tree_id = EXCLUDED.tree_id,
    relationship_type = EXCLUDED.relationship_type,
    status = EXCLUDED.status,
    start_date = EXCLUDED.start_date,
    end_date = EXCLUDED.end_date,
    legacy_father_id = EXCLUDED.legacy_father_id,
    legacy_mother_id = EXCLUDED.legacy_mother_id,
    deleted_at = EXCLUDED.deleted_at,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO family_unit_partners (family_unit_id, person_id, partner_order)
SELECT fu.family_unit_id, fu.legacy_father_id, 1
FROM family_units fu
WHERE fu.legacy_father_id IS NOT NULL
ON CONFLICT (family_unit_id, person_id) DO NOTHING;

INSERT INTO family_unit_partners (family_unit_id, person_id, partner_order)
SELECT fu.family_unit_id, fu.legacy_mother_id, 2
FROM family_units fu
WHERE fu.legacy_mother_id IS NOT NULL
ON CONFLICT (family_unit_id, person_id) DO NOTHING;

INSERT INTO family_units (
    tree_id,
    relationship_type,
    status,
    legacy_father_id,
    legacy_mother_id
)
SELECT DISTINCT m.tree_id,
       'unknown',
       'unknown',
       m.father_id,
       m.mother_id
FROM members m
LEFT JOIN family_units fu
  ON fu.legacy_father_id IS NOT DISTINCT FROM m.father_id
 AND fu.legacy_mother_id IS NOT DISTINCT FROM m.mother_id
 AND fu.tree_id = m.tree_id
 AND fu.deleted_at IS NULL
WHERE m.deleted_at IS NULL
  AND (m.father_id IS NOT NULL OR m.mother_id IS NOT NULL)
  AND fu.family_unit_id IS NULL
ON CONFLICT (tree_id, COALESCE(legacy_father_id, 0), COALESCE(legacy_mother_id, 0))
WHERE source_spouse_id IS NULL
DO NOTHING;

INSERT INTO family_unit_partners (family_unit_id, person_id, partner_order)
SELECT fu.family_unit_id, fu.legacy_father_id, 1
FROM family_units fu
WHERE fu.legacy_father_id IS NOT NULL
ON CONFLICT (family_unit_id, person_id) DO NOTHING;

INSERT INTO family_unit_partners (family_unit_id, person_id, partner_order)
SELECT fu.family_unit_id, fu.legacy_mother_id, 2
FROM family_units fu
WHERE fu.legacy_mother_id IS NOT NULL
ON CONFLICT (family_unit_id, person_id) DO NOTHING;

INSERT INTO family_unit_children (family_unit_id, child_person_id, relation_type)
SELECT fu.family_unit_id, m.member_id, 'biological'
FROM members m
JOIN family_units fu
  ON fu.tree_id = m.tree_id
 AND fu.legacy_father_id IS NOT DISTINCT FROM m.father_id
 AND fu.legacy_mother_id IS NOT DISTINCT FROM m.mother_id
 AND fu.deleted_at IS NULL
WHERE m.deleted_at IS NULL
  AND (m.father_id IS NOT NULL OR m.mother_id IS NOT NULL)
ON CONFLICT (family_unit_id, child_person_id) DO UPDATE SET
    relation_type = EXCLUDED.relation_type;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_sync_family_unit_from_member_parentage ON members;
DROP TRIGGER IF EXISTS trg_sync_family_unit_from_spouse ON members_spouse;
DROP FUNCTION IF EXISTS sync_family_unit_from_member_parentage();
DROP FUNCTION IF EXISTS sync_family_unit_from_spouse();

DROP TABLE IF EXISTS family_unit_children CASCADE;
DROP TABLE IF EXISTS family_unit_partners CASCADE;
DROP TABLE IF EXISTS family_units CASCADE;

-- +goose StatementEnd
