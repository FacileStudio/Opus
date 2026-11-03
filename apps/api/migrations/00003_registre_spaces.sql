-- +goose Up
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS registre_space_id TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_workspaces_registre_space_id ON workspaces (registre_space_id);

-- +goose Down
DROP INDEX IF EXISTS idx_workspaces_registre_space_id;
ALTER TABLE workspaces DROP COLUMN IF EXISTS registre_space_id;
