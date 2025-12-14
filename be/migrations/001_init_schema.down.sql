-- Drop tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS score_history CASCADE;
DROP TABLE IF EXISTS member_history CASCADE;
DROP TABLE IF EXISTS spouses CASCADE;
DROP TABLE IF EXISTS members CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";
