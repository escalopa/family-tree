INSERT INTO languages (language_code, is_active, display_order)
VALUES ('en', true, 1)
ON CONFLICT (language_code) DO UPDATE SET
  is_active = EXCLUDED.is_active,
  display_order = EXCLUDED.display_order;

INSERT INTO users (full_name, email, avatar, role_id, is_active)
VALUES
  ('Mock Super Admin', 'superadmin.mock@example.test', NULL, 400, true),
  ('Mock Admin', 'admin.mock@example.test', NULL, 300, true),
  ('Mock Guest', 'guest.mock@example.test', NULL, 200, true)
ON CONFLICT (email) DO UPDATE SET
  full_name = EXCLUDED.full_name,
  avatar = EXCLUDED.avatar,
  role_id = EXCLUDED.role_id,
  is_active = EXCLUDED.is_active;

DO $$
DECLARE
  seed_exists boolean;
  person record;
  new_member_id int;
  actor_user_id int;
  field_points int;
  history_version int;
BEGIN
  SELECT EXISTS (
    SELECT 1
    FROM member_names
    WHERE language_code = 'en'
      AND name LIKE 'Test Family Member %'
  ) INTO seed_exists;

  IF seed_exists THEN
    RETURN;
  END IF;

  CREATE TEMP TABLE seed_people (
    seq int PRIMARY KEY,
    gender char(1) NOT NULL,
    father_seq int,
    mother_seq int,
    born_year int NOT NULL
  ) ON COMMIT DROP;

  CREATE TEMP TABLE seed_member_map (
    seq int PRIMARY KEY,
    member_id int NOT NULL
  ) ON COMMIT DROP;

  INSERT INTO seed_people (seq, gender, father_seq, mother_seq, born_year)
  SELECT gs,
         CASE WHEN gs % 2 = 1 THEN 'M' ELSE 'F' END,
         NULL,
         NULL,
         1938 + gs
  FROM generate_series(1, 8) AS gs;

  INSERT INTO seed_people (seq, gender, father_seq, mother_seq, born_year)
  SELECT gs,
         CASE WHEN gs % 2 = 1 THEN 'M' ELSE 'F' END,
         1 + (((gs - 9) / 6) * 2),
         2 + (((gs - 9) / 6) * 2),
         1960 + ((gs - 9) % 24)
  FROM generate_series(9, 32) AS gs;

  INSERT INTO seed_people (seq, gender, father_seq, mother_seq, born_year)
  SELECT gs,
         CASE WHEN gs % 2 = 1 THEN 'M' ELSE 'F' END,
         9 + (((gs - 33) / 4) * 2),
         10 + (((gs - 33) / 4) * 2),
         1988 + ((gs - 33) % 24)
  FROM generate_series(33, 80) AS gs;

  INSERT INTO seed_people (seq, gender, father_seq, mother_seq, born_year)
  SELECT gs,
         CASE WHEN gs % 2 = 1 THEN 'M' ELSE 'F' END,
         33 + (((gs - 81) / 2) * 2),
         34 + (((gs - 81) / 2) * 2),
         2013 + ((gs - 81) % 10)
  FROM generate_series(81, 100) AS gs;

  FOR person IN
    SELECT *
    FROM seed_people
    ORDER BY seq
  LOOP
    INSERT INTO members (
      gender,
      date_of_birth,
      father_id,
      mother_id,
      nicknames,
      profession,
      version
    )
    VALUES (
      person.gender,
      make_date(person.born_year, ((person.seq - 1) % 12) + 1, ((person.seq - 1) % 27) + 1),
      (SELECT member_id FROM seed_member_map WHERE seq = person.father_seq),
      (SELECT member_id FROM seed_member_map WHERE seq = person.mother_seq),
      ARRAY['seed-' || lpad(person.seq::text, 3, '0')],
      CASE
        WHEN person.seq <= 8 THEN 'Family elder'
        WHEN person.seq <= 32 THEN 'Family organizer'
        WHEN person.seq <= 80 THEN 'Family contributor'
        ELSE 'Young family member'
      END,
      1
    )
    RETURNING member_id INTO new_member_id;

    INSERT INTO seed_member_map (seq, member_id)
    VALUES (person.seq, new_member_id);

    INSERT INTO member_names (member_id, language_code, name, created_at, updated_at)
    VALUES (
      new_member_id,
      'en',
      'Test Family Member ' || lpad(person.seq::text, 3, '0'),
      CURRENT_TIMESTAMP,
      CURRENT_TIMESTAMP
    );

    SELECT user_id INTO actor_user_id
    FROM users
    WHERE email = CASE person.seq % 3
      WHEN 0 THEN 'superadmin.mock@example.test'
      WHEN 1 THEN 'admin.mock@example.test'
      ELSE 'guest.mock@example.test'
    END;

    field_points := CASE person.seq % 3
      WHEN 0 THEN 5
      WHEN 1 THEN 3
      ELSE 1
    END;

    INSERT INTO members_history (
      member_id,
      user_id,
      change_type,
      old_values,
      new_values,
      member_version
    )
    VALUES (
      new_member_id,
      actor_user_id,
      'INSERT',
      NULL,
      jsonb_build_object(
        'names', jsonb_build_object('en', 'Test Family Member ' || lpad(person.seq::text, 3, '0')),
        'gender', person.gender
      ),
      1
    )
    RETURNING member_version INTO history_version;

    INSERT INTO user_scores (
      user_id,
      member_id,
      field_name,
      points,
      member_version
    )
    VALUES (
      actor_user_id,
      new_member_id,
      'seed_member',
      field_points,
      history_version
    );
  END LOOP;
END $$;
