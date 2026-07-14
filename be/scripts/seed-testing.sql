INSERT INTO languages (language_code, is_active, display_order)
VALUES
  ('en', true, 1),
  ('ar', true, 2)
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
  person record;
  new_member_id int;
  actor_user_id int;
  field_points int;
  history_version int;
BEGIN
  CREATE TEMP TABLE seed_existing_members (
    member_id int PRIMARY KEY
  ) ON COMMIT DROP;

  INSERT INTO seed_existing_members (member_id)
  SELECT member_id
  FROM member_names
  WHERE language_code = 'en'
    AND name LIKE 'Test Family Member %';

  DELETE FROM user_scores
  WHERE member_id IN (SELECT member_id FROM seed_existing_members);

  DELETE FROM members_history
  WHERE member_id IN (SELECT member_id FROM seed_existing_members);

  DELETE FROM members_spouse
  WHERE father_id IN (SELECT member_id FROM seed_existing_members)
     OR mother_id IN (SELECT member_id FROM seed_existing_members);

  DELETE FROM member_names
  WHERE member_id IN (SELECT member_id FROM seed_existing_members);

  UPDATE members
  SET father_id = NULL
  WHERE father_id IN (SELECT member_id FROM seed_existing_members);

  UPDATE members
  SET mother_id = NULL
  WHERE mother_id IN (SELECT member_id FROM seed_existing_members);

  DELETE FROM members
  WHERE member_id IN (SELECT member_id FROM seed_existing_members);

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

  CREATE TEMP TABLE seed_spouses (
    father_seq int NOT NULL,
    mother_seq int NOT NULL,
    marriage_date date,
    divorce_date date
  ) ON COMMIT DROP;

  INSERT INTO seed_people (seq, gender, father_seq, mother_seq, born_year)
  VALUES
    (1, 'M', NULL, NULL, 1938),
    (2, 'F', NULL, NULL, 1940);

  INSERT INTO seed_people (seq, gender, father_seq, mother_seq, born_year)
  SELECT gs,
         CASE WHEN gs % 2 = 1 THEN 'M' ELSE 'F' END,
         1,
         2,
         1962 + ((gs - 3) % 18)
  FROM generate_series(3, 20) AS gs;

  INSERT INTO seed_people (seq, gender, father_seq, mother_seq, born_year)
  SELECT gs,
         CASE WHEN gs % 2 = 1 THEN 'M' ELSE 'F' END,
         CASE WHEN gs BETWEEN 66 AND 70 THEN 3 ELSE 3 + (((gs - 21) / 5) % 9) * 2 END,
         CASE WHEN gs BETWEEN 66 AND 70 THEN 6 ELSE 4 + (((gs - 21) / 5) % 9) * 2 END,
         1988 + ((gs - 21) % 22)
  FROM generate_series(21, 70) AS gs;

  INSERT INTO seed_people (seq, gender, father_seq, mother_seq, born_year)
  SELECT gs,
         CASE WHEN gs % 2 = 1 THEN 'M' ELSE 'F' END,
         21 + (((gs - 71) / 3) % 25) * 2,
         22 + (((gs - 71) / 3) % 25) * 2,
         2012 + ((gs - 71) % 10)
  FROM generate_series(71, 100) AS gs;

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
        WHEN person.seq <= 2 THEN 'Family founder'
        WHEN person.seq <= 20 THEN 'Family elder'
        WHEN person.seq <= 70 THEN 'Family contributor'
        ELSE 'Young family member'
      END,
      1
    )
    RETURNING member_id INTO new_member_id;

    INSERT INTO seed_member_map (seq, member_id)
    VALUES (person.seq, new_member_id);

    INSERT INTO member_names (member_id, language_code, name, created_at, updated_at)
    VALUES
      (
        new_member_id,
        'en',
        'Test Family Member ' || lpad(person.seq::text, 3, '0'),
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
      ),
      (
        new_member_id,
        'ar',
        'عضو عائلة تجريبي ' || lpad(person.seq::text, 3, '0'),
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
        'names', jsonb_build_object(
          'en', 'Test Family Member ' || lpad(person.seq::text, 3, '0'),
          'ar', 'عضو عائلة تجريبي ' || lpad(person.seq::text, 3, '0')
        ),
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

  INSERT INTO seed_spouses (father_seq, mother_seq, marriage_date, divorce_date)
  VALUES
    (1, 2, DATE '1960-06-01', NULL),
    (3, 4, DATE '1985-04-12', DATE '2001-09-15'),
    (3, 6, DATE '2003-03-10', NULL),
    (5, 6, DATE '1986-05-20', DATE '2002-11-01');

  INSERT INTO seed_spouses (father_seq, mother_seq, marriage_date, divorce_date)
  SELECT father_seq,
         father_seq + 1,
         make_date(1984 + ((father_seq - 3) / 2), 6, 15),
         NULL
  FROM generate_series(7, 19, 2) AS gs(father_seq);

  INSERT INTO seed_spouses (father_seq, mother_seq, marriage_date, divorce_date)
  SELECT father_seq,
         father_seq + 1,
         make_date(2010 + ((father_seq - 21) / 2), 5, 18),
         CASE WHEN father_seq IN (21, 25, 29) THEN make_date(2018 + ((father_seq - 21) / 2), 8, 20) ELSE NULL END
  FROM generate_series(21, 69, 2) AS gs(father_seq);

  INSERT INTO members_spouse (father_id, mother_id, marriage_date, divorce_date)
  SELECT father_map.member_id,
         mother_map.member_id,
         seed_spouses.marriage_date,
         seed_spouses.divorce_date
  FROM seed_spouses
  JOIN seed_member_map father_map ON father_map.seq = seed_spouses.father_seq
  JOIN seed_member_map mother_map ON mother_map.seq = seed_spouses.mother_seq
  ON CONFLICT (father_id, mother_id) DO UPDATE SET
    marriage_date = EXCLUDED.marriage_date,
    divorce_date = EXCLUDED.divorce_date,
    deleted_at = NULL;
END $$;
