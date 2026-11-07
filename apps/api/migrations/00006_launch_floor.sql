-- +goose Up
ALTER TABLE porte_login_codes ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '',
                              ADD COLUMN IF NOT EXISTS subject TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS directory_deletion_state (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    state BYTEA NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE comments ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE comments DROP CONSTRAINT IF EXISTS comments_user_id_fkey;
ALTER TABLE comments ADD CONSTRAINT comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL;

ALTER TABLE share_links ALTER COLUMN created_by DROP NOT NULL;
ALTER TABLE share_links DROP CONSTRAINT IF EXISTS share_links_created_by_fkey;
ALTER TABLE share_links ADD CONSTRAINT share_links_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE share_links DROP CONSTRAINT IF EXISTS share_links_created_by_fkey;
ALTER TABLE share_links ADD CONSTRAINT share_links_created_by_fkey FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE CASCADE;
ALTER TABLE comments DROP CONSTRAINT IF EXISTS comments_user_id_fkey;
ALTER TABLE comments ADD CONSTRAINT comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
DROP TABLE IF EXISTS directory_deletion_state;
ALTER TABLE porte_login_codes DROP COLUMN IF EXISTS subject, DROP COLUMN IF EXISTS provider;
