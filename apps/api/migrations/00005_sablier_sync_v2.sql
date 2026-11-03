-- +goose Up
ALTER TABLE projects ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;
UPDATE projects SET updated_at = created_at WHERE updated_at IS NULL;
ALTER TABLE projects ALTER COLUMN updated_at SET DEFAULT now();
ALTER TABLE projects ALTER COLUMN updated_at SET NOT NULL;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;

UPDATE projects SET facile_id = md5(random()::text || clock_timestamp()::text || id::text)
WHERE facile_id IS NULL OR btrim(facile_id) = '';
UPDATE tasks SET facile_id = md5(random()::text || clock_timestamp()::text || id::text)
WHERE facile_id IS NULL OR btrim(facile_id) = '';

ALTER TABLE time_entries ADD COLUMN IF NOT EXISTS facile_id TEXT;
ALTER TABLE time_entries ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'opus';
ALTER TABLE time_entries ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE UNIQUE INDEX IF NOT EXISTS idx_time_entries_facile_id ON time_entries (facile_id);

CREATE TABLE IF NOT EXISTS antenne_outbox (
    id BIGSERIAL PRIMARY KEY,
    channel TEXT NOT NULL,
    payload TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS antenne_outbox;
DROP INDEX IF EXISTS idx_time_entries_facile_id;
DELETE FROM time_entries WHERE source = 'sablier';
ALTER TABLE time_entries DROP COLUMN IF EXISTS updated_at;
ALTER TABLE time_entries DROP COLUMN IF EXISTS source;
ALTER TABLE time_entries DROP COLUMN IF EXISTS facile_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS synced_at;
ALTER TABLE projects DROP COLUMN IF EXISTS synced_at;
ALTER TABLE projects DROP COLUMN IF EXISTS updated_at;
