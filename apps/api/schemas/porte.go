package schemas

import "gorm.io/gorm"

// AdoptPorte creates porte's storage. The Go backend keeps its own users table
// (schemas.User) and reads identities, sessions and login codes through
// portepg, whose stores only ever touch the three tables below — porte's own
// porte_users is left out and the foreign keys point at users(id) instead.
//
// The one UPDATE is porte v0.3.0's re-keying of password identities onto the
// account id. It is idempotent: once every local subject holds the account id
// the predicate is false for every row. Unlike porte, which tolerates this
// statement failing, the whole script is one Exec here: the state that could
// make it fail (two password identities on one account) is unreachable on a
// schema this port created, and if it is ever reached, refusing to boot is
// the right answer to ambiguous credentials.
//
// Sablier's AdoptPorte also backfills OIDC subjects from a pre-port users
// table; Opus has none — the port created the schema it migrates — so there
// is nothing to backfill and no issuer parameter.
//
// Every statement is PostgreSQL. The tests run against in-memory SQLite and
// porte/pg is Postgres-only, so on any other dialect there is nothing to
// create and the call returns early.
func AdoptPorte(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	return db.Exec(porteSchema).Error
}

// porteSchema is porte/pg's Schema with its porte_users table left out and the
// foreign keys repointed at the app's own users table.
const porteSchema = `
CREATE TABLE IF NOT EXISTS porte_identities (
	user_id         bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	provider        text NOT NULL,
	subject         text NOT NULL,
	password_hash   text NOT NULL DEFAULT '',
	access_token    text NOT NULL DEFAULT '',
	refresh_token   text NOT NULL DEFAULT '',
	token_expiry    timestamptz,
	roles           jsonb,
	roles_synced_at timestamptz,
	synced_at       timestamptz,
	created_at      timestamptz DEFAULT now(),
	PRIMARY KEY (provider, subject)
);
CREATE INDEX IF NOT EXISTS porte_identities_user_idx ON porte_identities (user_id);
ALTER TABLE porte_identities ADD COLUMN IF NOT EXISTS created_at timestamptz;
ALTER TABLE porte_identities ALTER COLUMN created_at SET DEFAULT now();

UPDATE porte_identities SET subject = user_id::text
 WHERE provider = 'local' AND subject <> user_id::text;

CREATE TABLE IF NOT EXISTS porte_sessions (
	id           bigserial PRIMARY KEY,
	token_hash   text NOT NULL UNIQUE,
	user_id      bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	label        text NOT NULL DEFAULT '',
	created_at   timestamptz NOT NULL DEFAULT now(),
	last_used_at timestamptz NOT NULL DEFAULT now(),
	expires_at   timestamptz
);
CREATE INDEX IF NOT EXISTS porte_sessions_user_idx ON porte_sessions (user_id);
CREATE INDEX IF NOT EXISTS porte_sessions_expiry_idx ON porte_sessions (expires_at)
	WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS porte_login_codes (
	code_hash   text PRIMARY KEY,
	user_id     bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at  timestamptz NOT NULL,
	consumed_at timestamptz
);
ALTER TABLE porte_login_codes ADD COLUMN IF NOT EXISTS consumed_at timestamptz;
`
