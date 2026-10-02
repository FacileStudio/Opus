#!/bin/sh
set -eu

urlencode() {
  if command -v bun >/dev/null 2>&1; then
    bun -e 'const input = process.argv[1]; process.stdout.write(encodeURIComponent(input).replace(/[!\x27()*]/g, (char) => `%${char.charCodeAt(0).toString(16).toUpperCase()}`));' "$1"
  elif command -v node >/dev/null 2>&1; then
    node -e 'const input = process.argv[1]; process.stdout.write(encodeURIComponent(input).replace(/[!\x27()*]/g, (char) => `%${char.charCodeAt(0).toString(16).toUpperCase()}`));' "$1"
  else
    printf '%s' "$1"
  fi
}

# Derive OPUS_API_URL from OPUS_CLIENT_URL if not explicitly set
client_url_trimmed="$(printf '%s' "${OPUS_CLIENT_URL:-}" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
client_url="${client_url_trimmed%/}"
if [ -z "${OPUS_API_URL:-}" ] && [ -n "$client_url" ]; then
  export OPUS_API_URL="${client_url}/api"
  echo "OPUS_API_URL not set — derived from OPUS_CLIENT_URL: $OPUS_API_URL"
fi

# Derive DATABASE_URL from individual postgres vars if not explicitly set
if [ -z "${DATABASE_URL:-}" ]; then
  POSTGRES_DB="${POSTGRES_DB:-opus_go}"
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

if [ -n "${GO_DATABASE_URL:-}" ]; then
  export DATABASE_URL="$GO_DATABASE_URL"
fi

export PORT="${PORT:-27538}"
export CLIENT_DIR="${CLIENT_DIR:-/client}"
export UPLOAD_DIR="${UPLOAD_DIR:-/app/data/uploads}"

if [ -f /usr/local/bin/web-env.sh ]; then
  /usr/local/bin/web-env.sh
fi

echo "Starting Opus Go backend..."
exec /usr/local/bin/opus-api
