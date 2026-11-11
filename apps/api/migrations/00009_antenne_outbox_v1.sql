-- +goose Up
ALTER TABLE antenne_outbox ALTER COLUMN payload TYPE JSONB USING payload::jsonb;
ALTER TABLE antenne_outbox ADD COLUMN idempotency_key TEXT;
UPDATE antenne_outbox o SET idempotency_key = k.key
FROM (
    SELECT id, CASE WHEN row_number() OVER (PARTITION BY base ORDER BY id) = 1 THEN base ELSE base || ':' || id END AS key
    FROM (SELECT id, COALESCE(NULLIF(payload->>'idempotency_key', ''), 'opus-outbox:' || id) AS base FROM antenne_outbox) b
) k
WHERE k.id = o.id;
ALTER TABLE antenne_outbox ALTER COLUMN idempotency_key SET NOT NULL;
ALTER TABLE antenne_outbox ADD CONSTRAINT antenne_outbox_idempotency_key_key UNIQUE (idempotency_key);
ALTER TABLE antenne_outbox ADD COLUMN sent_at TIMESTAMPTZ;
ALTER TABLE antenne_outbox DROP COLUMN attempts;
ALTER TABLE antenne_outbox DROP COLUMN last_error;
CREATE INDEX antenne_outbox_unsent ON antenne_outbox (id) WHERE sent_at IS NULL;
CREATE INDEX antenne_outbox_sent ON antenne_outbox (sent_at) WHERE sent_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS antenne_outbox_sent;
DROP INDEX IF EXISTS antenne_outbox_unsent;
DELETE FROM antenne_outbox WHERE sent_at IS NOT NULL;
ALTER TABLE antenne_outbox DROP COLUMN sent_at;
ALTER TABLE antenne_outbox DROP CONSTRAINT antenne_outbox_idempotency_key_key;
ALTER TABLE antenne_outbox DROP COLUMN idempotency_key;
ALTER TABLE antenne_outbox ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE antenne_outbox ADD COLUMN last_error TEXT NOT NULL DEFAULT '';
ALTER TABLE antenne_outbox ALTER COLUMN payload TYPE TEXT USING payload::text;
