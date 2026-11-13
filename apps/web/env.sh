#!/bin/sh
set -e

client_dir="${CLIENT_DIR:-/client}"
echo "Starting environment variable replacement in $client_dir..."

replace() {
  key="$1"
  value="$2"
  find "$client_dir" -type f -name "*.js" -exec grep -l "$key" {} \; | xargs -r sed -i "s#$key#$value#g"
  echo "Replaced $key"
}

if [ -n "$OPUS_API_URL" ]; then
  replace OPUS_API_URL "$OPUS_API_URL"
else
  echo "WARNING: OPUS_API_URL environment variable is not set. API calls may fail."
fi

if [ -n "$OPUS_CLIENT_URL" ]; then
  replace OPUS_CLIENT_URL "$OPUS_CLIENT_URL"
fi

echo "Environment variable replacement complete"
