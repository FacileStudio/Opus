-- +goose Up
ALTER TABLE projects ADD COLUMN IF NOT EXISTS facile_id TEXT;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS facile_id TEXT;

UPDATE projects SET facile_id = md5(random()::text || clock_timestamp()::text || id::text) WHERE facile_id IS NULL OR facile_id = '';
UPDATE tasks SET facile_id = md5(random()::text || clock_timestamp()::text || id::text) WHERE facile_id IS NULL OR facile_id = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_facile_id ON projects (facile_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_facile_id ON tasks (facile_id);

CREATE TABLE IF NOT EXISTS pool_processed_events (
    idempotency_key TEXT PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pool_processed_events_processed_at ON pool_processed_events (processed_at);

-- +goose Down
DROP TABLE IF EXISTS pool_processed_events;
DROP INDEX IF EXISTS idx_tasks_facile_id;
DROP INDEX IF EXISTS idx_projects_facile_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS facile_id;
ALTER TABLE projects DROP COLUMN IF EXISTS facile_id;
