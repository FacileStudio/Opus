-- +goose Up
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS personal_owner_id BIGINT REFERENCES users (id) ON DELETE CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_workspaces_personal_owner_id ON workspaces (personal_owner_id);

INSERT INTO workspaces (name, slug, created_at, personal_owner_id)
SELECT 'Personal', 'personal-' || u.id, now(), u.id
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM workspaces w WHERE w.personal_owner_id = u.id)
ON CONFLICT DO NOTHING;

INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
SELECT w.id, w.personal_owner_id, 'owner', now()
FROM workspaces w
WHERE w.personal_owner_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM workspace_members m WHERE m.workspace_id = w.id AND m.user_id = w.personal_owner_id);

-- +goose Down
DELETE FROM workspaces WHERE personal_owner_id IS NOT NULL;
DROP INDEX IF EXISTS idx_workspaces_personal_owner_id;
ALTER TABLE workspaces DROP COLUMN IF EXISTS personal_owner_id;
