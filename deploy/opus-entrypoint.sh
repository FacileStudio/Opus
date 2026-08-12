#!/bin/sh
set -eu

urlencode() {
  bun -e 'const input = process.argv[1]; process.stdout.write(encodeURIComponent(input).replace(/[!\x27()*]/g, (char) => `%${char.charCodeAt(0).toString(16).toUpperCase()}`));' "$1"
}

legacy_pid=""
go_pid=""

cleanup() {
  if [ -n "$legacy_pid" ]; then
    kill "$legacy_pid" 2>/dev/null || true
  fi

  if [ -n "$go_pid" ]; then
    kill "$go_pid" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

# Derive OPUS_API_URL from OPUS_CLIENT_URL if not explicitly set
client_url_trimmed="$(printf '%s' "${OPUS_CLIENT_URL:-}" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
client_url="${client_url_trimmed%/}"
if [ -z "${OPUS_API_URL:-}" ] && [ -n "$client_url" ]; then
  export OPUS_API_URL="${client_url}/api"
  echo "OPUS_API_URL not set — derived from OPUS_CLIENT_URL: $OPUS_API_URL"
fi

# Derive DATABASE_URL from individual postgres vars if not explicitly set
if [ -z "${DATABASE_URL:-}" ]; then
  POSTGRES_DB="${POSTGRES_DB:-opus}"
  POSTGRES_USER="${POSTGRES_USER:-opus}"
  if [ -n "${POSTGRES_PASSWORD:-}" ]; then
    encoded_user="$(urlencode "$POSTGRES_USER")"
    encoded_password="$(urlencode "$POSTGRES_PASSWORD")"
    export DATABASE_URL="postgresql://${encoded_user}:${encoded_password}@${POSTGRES_HOST:-postgres}:${POSTGRES_PORT:-5432}/${POSTGRES_DB}"
    echo "DATABASE_URL not set — derived from POSTGRES_* vars"
  else
    echo "ERROR: DATABASE_URL is not set and POSTGRES_PASSWORD is not set" >&2
    exit 1
  fi
fi

# Auto-generate AUTH_SECRET if not set
if [ -z "${AUTH_SECRET:-}" ]; then
  export AUTH_SECRET="$(bun -e 'process.stdout.write(require("node:crypto").randomBytes(32).toString("hex"))')"
  echo "WARNING: AUTH_SECRET not set — generated a random secret for this session."
  echo "WARNING: Set AUTH_SECRET in your .env to persist sessions across restarts."
fi

# The legacy API keeps its pre-port uploads path even though its source now
# lives under apps/api-legacy.
export UPLOAD_DIR="${UPLOAD_DIR:-/app/apps/api/data/uploads}"

/usr/local/bin/web-env.sh

bun /app/apps/api/dist/index.js &
legacy_pid=$!

echo "Waiting for the legacy API to be ready..."
until wget --spider --quiet http://127.0.0.1:1337/api/health 2>/dev/null; do
  if ! kill -0 "$legacy_pid" 2>/dev/null; then
    echo "Legacy API process exited unexpectedly"
    exit 1
  fi
  sleep 1
done
echo "Legacy API is ready"

export PORT="${PORT:-80}"
export CLIENT_DIR="${CLIENT_DIR:-/client}"
export LEGACY_API_URL="${LEGACY_API_URL:-http://127.0.0.1:1337}"

# The Go backend owns its own schema on its own database: the port adopts the
# suite's int64 identity model, while the legacy Drizzle schema keys everything
# on CUID2 text ids. The two cannot share one database, so the Go process gets
# its own URL scoped to this one process — the DATABASE_URL above still belongs
# to the legacy API.
if [ -z "${GO_DATABASE_URL:-}" ]; then
  go_db="${POSTGRES_DB_GO:-opus_go}"
  if [ -n "${POSTGRES_PASSWORD:-}" ]; then
    encoded_go_user="$(urlencode "$POSTGRES_USER")"
    encoded_go_password="$(urlencode "$POSTGRES_PASSWORD")"
    GO_DATABASE_URL="postgresql://${encoded_go_user}:${encoded_go_password}@${POSTGRES_HOST:-postgres}:${POSTGRES_PORT:-5432}/${go_db}"
  else
    GO_DATABASE_URL="$DATABASE_URL"
    echo "WARNING: no POSTGRES_PASSWORD — the Go backend shares the legacy database"
  fi
  echo "GO_DATABASE_URL not set — derived from POSTGRES_* vars"
fi

DATABASE_URL="$GO_DATABASE_URL" /usr/local/bin/opus-api &
go_pid=$!

echo "Waiting for the Go API to be ready..."
until wget --spider --quiet http://127.0.0.1:80/api/health 2>/dev/null; do
  if ! kill -0 "$go_pid" 2>/dev/null; then
    echo "Go API process exited unexpectedly"
    exit 1
  fi
  sleep 1
done
echo "Go API is ready"

while kill -0 "$legacy_pid" 2>/dev/null && kill -0 "$go_pid" 2>/dev/null; do
  sleep 1
done

first_failed_pid=""
first_failed_status=""
legacy_status=""
go_status=""

if ! kill -0 "$legacy_pid" 2>/dev/null; then
  wait "$legacy_pid" || legacy_status=$?
  legacy_status=${legacy_status:-0}
  first_failed_pid="$legacy_pid"
  first_failed_status="$legacy_status"
fi

if [ -z "$first_failed_pid" ] && ! kill -0 "$go_pid" 2>/dev/null; then
  wait "$go_pid" || go_status=$?
  go_status=${go_status:-0}
  first_failed_pid="$go_pid"
  first_failed_status="$go_status"
fi

cleanup

if [ "$first_failed_pid" != "$legacy_pid" ]; then
  wait "$legacy_pid" || legacy_status=$?
fi

if [ "$first_failed_pid" != "$go_pid" ]; then
  wait "$go_pid" || go_status=$?
fi

legacy_status=${legacy_status:-0}
go_status=${go_status:-0}

if [ -n "$first_failed_status" ]; then
  exit "$first_failed_status"
fi

if [ "$legacy_status" -ne 0 ]; then
  exit "$legacy_status"
fi

exit "$go_status"
