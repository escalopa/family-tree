-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS family_trees (
    tree_id SERIAL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_user_id INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE family_trees
    ADD CONSTRAINT pk_family_trees PRIMARY KEY (tree_id),
    ADD CONSTRAINT fk_family_trees_owner FOREIGN KEY (owner_user_id) REFERENCES users(user_id);

CREATE INDEX IF NOT EXISTS idx_family_trees_owner_user_id ON family_trees(owner_user_id);

CREATE TABLE IF NOT EXISTS family_tree_memberships (
    tree_id INT NOT NULL,
    user_id INT NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'editor',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE family_tree_memberships
    ADD CONSTRAINT pk_family_tree_memberships PRIMARY KEY (tree_id, user_id),
    ADD CONSTRAINT fk_family_tree_memberships_tree FOREIGN KEY (tree_id) REFERENCES family_trees(tree_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_tree_memberships_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    ADD CONSTRAINT chk_family_tree_memberships_role CHECK (role IN ('owner', 'editor', 'viewer'));

CREATE INDEX IF NOT EXISTS idx_family_tree_memberships_user_id ON family_tree_memberships(user_id);

CREATE TABLE IF NOT EXISTS family_tree_invitations (
    invitation_id SERIAL,
    tree_id INT NOT NULL,
    inviter_user_id INT NOT NULL,
    invitee_user_id INT,
    invitee_email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    responded_at TIMESTAMP
);

ALTER TABLE family_tree_invitations
    ADD CONSTRAINT pk_family_tree_invitations PRIMARY KEY (invitation_id),
    ADD CONSTRAINT fk_family_tree_invitations_tree FOREIGN KEY (tree_id) REFERENCES family_trees(tree_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_tree_invitations_inviter FOREIGN KEY (inviter_user_id) REFERENCES users(user_id),
    ADD CONSTRAINT fk_family_tree_invitations_invitee FOREIGN KEY (invitee_user_id) REFERENCES users(user_id),
    ADD CONSTRAINT chk_family_tree_invitations_status CHECK (status IN ('pending', 'accepted', 'declined', 'revoked'));

CREATE INDEX IF NOT EXISTS idx_family_tree_invitations_invitee_user_id ON family_tree_invitations(invitee_user_id);
CREATE INDEX IF NOT EXISTS idx_family_tree_invitations_email ON family_tree_invitations(LOWER(invitee_email));
CREATE INDEX IF NOT EXISTS idx_family_tree_invitations_tree_status ON family_tree_invitations(tree_id, status);

CREATE TABLE IF NOT EXISTS family_tree_share_links (
    share_id SERIAL,
    tree_id INT NOT NULL,
    token VARCHAR(128) NOT NULL,
    created_by INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    max_visits INT,
    visit_count INT NOT NULL DEFAULT 0,
    revoked_at TIMESTAMP
);

ALTER TABLE family_tree_share_links
    ADD CONSTRAINT pk_family_tree_share_links PRIMARY KEY (share_id),
    ADD CONSTRAINT fk_family_tree_share_links_tree FOREIGN KEY (tree_id) REFERENCES family_trees(tree_id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_family_tree_share_links_created_by FOREIGN KEY (created_by) REFERENCES users(user_id),
    ADD CONSTRAINT uq_family_tree_share_links_token UNIQUE (token),
    ADD CONSTRAINT chk_family_tree_share_links_max_visits CHECK (max_visits IS NULL OR max_visits > 0),
    ADD CONSTRAINT chk_family_tree_share_links_visit_count CHECK (visit_count >= 0);

CREATE INDEX IF NOT EXISTS idx_family_tree_share_links_tree_id ON family_tree_share_links(tree_id);
CREATE INDEX IF NOT EXISTS idx_family_tree_share_links_token ON family_tree_share_links(token);

ALTER TABLE members
    ADD COLUMN IF NOT EXISTS tree_id INT;

ALTER TABLE members
    ADD CONSTRAINT fk_members_tree FOREIGN KEY (tree_id) REFERENCES family_trees(tree_id);

CREATE INDEX IF NOT EXISTS idx_members_tree_id ON members(tree_id);
CREATE INDEX IF NOT EXISTS idx_members_tree_deleted ON members(tree_id, deleted_at);

WITH owner AS (
    SELECT user_id
    FROM users
    ORDER BY CASE WHEN role_id = 400 THEN 0 ELSE 1 END, user_id
    LIMIT 1
),
default_tree AS (
    INSERT INTO family_trees (name, description, owner_user_id)
    SELECT 'Default Family Tree', 'Migrated from the original single-tree workspace.', owner.user_id
    FROM owner
    WHERE EXISTS (SELECT 1 FROM members WHERE tree_id IS NULL)
      AND NOT EXISTS (SELECT 1 FROM family_trees)
    RETURNING tree_id, owner_user_id
),
fallback_tree AS (
    SELECT tree_id, owner_user_id FROM default_tree
    UNION ALL
    SELECT tree_id, owner_user_id FROM family_trees ORDER BY tree_id LIMIT 1
)
UPDATE members
SET tree_id = (SELECT tree_id FROM fallback_tree LIMIT 1)
WHERE tree_id IS NULL
  AND EXISTS (SELECT 1 FROM fallback_tree);

ALTER TABLE members
    ALTER COLUMN tree_id SET NOT NULL;

INSERT INTO family_tree_memberships (tree_id, user_id, role)
SELECT ft.tree_id, ft.owner_user_id, 'owner'
FROM family_trees ft
ON CONFLICT (tree_id, user_id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE members DROP CONSTRAINT IF EXISTS fk_members_tree;
DROP INDEX IF EXISTS idx_members_tree_deleted;
DROP INDEX IF EXISTS idx_members_tree_id;
ALTER TABLE members DROP COLUMN IF EXISTS tree_id;

DROP TABLE IF EXISTS family_tree_share_links CASCADE;
DROP TABLE IF EXISTS family_tree_invitations CASCADE;
DROP TABLE IF EXISTS family_tree_memberships CASCADE;
DROP TABLE IF EXISTS family_trees CASCADE;

-- +goose StatementEnd
