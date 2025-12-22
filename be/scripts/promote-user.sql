-- Promote user to super_admin and activate
-- Usage: psql -U familytree -d familytree -f scripts/promote-user.sql -v email='user@example.com'

UPDATE users
SET
    role_id = 400,
    is_active = true
WHERE email = :'email'
RETURNING *;
