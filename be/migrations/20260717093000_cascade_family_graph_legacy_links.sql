-- +goose Up
-- +goose StatementBegin

ALTER TABLE family_units DROP CONSTRAINT IF EXISTS fk_family_units_source_spouse;
ALTER TABLE family_units DROP CONSTRAINT IF EXISTS fk_family_units_legacy_father;
ALTER TABLE family_units DROP CONSTRAINT IF EXISTS fk_family_units_legacy_mother;

ALTER TABLE family_units
    ADD CONSTRAINT fk_family_units_source_spouse
        FOREIGN KEY (source_spouse_id) REFERENCES members_spouse(spouse_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_units_legacy_father
        FOREIGN KEY (legacy_father_id) REFERENCES members(member_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_units_legacy_mother
        FOREIGN KEY (legacy_mother_id) REFERENCES members(member_id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE family_units DROP CONSTRAINT IF EXISTS fk_family_units_source_spouse;
ALTER TABLE family_units DROP CONSTRAINT IF EXISTS fk_family_units_legacy_father;
ALTER TABLE family_units DROP CONSTRAINT IF EXISTS fk_family_units_legacy_mother;

ALTER TABLE family_units
    ADD CONSTRAINT fk_family_units_source_spouse
        FOREIGN KEY (source_spouse_id) REFERENCES members_spouse(spouse_id),
    ADD CONSTRAINT fk_family_units_legacy_father
        FOREIGN KEY (legacy_father_id) REFERENCES members(member_id),
    ADD CONSTRAINT fk_family_units_legacy_mother
        FOREIGN KEY (legacy_mother_id) REFERENCES members(member_id);

-- +goose StatementEnd
