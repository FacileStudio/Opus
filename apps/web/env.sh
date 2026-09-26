#!/bin/sh
set -e

client_dir="${CLIENT_DIR:-/client}"
echo "Starting environment variable replacement in $client_dir..."

# Process OPUS_API_URL first (with special handling)
if [ ! -z "$OPUS_API_URL" ]; then
  echo "Found OPUS_API_URL: $OPUS_API_URL"

  # Replace the exact string "OPUS_API_URL" in all JavaScript files
  find "$client_dir" -type f -name "*.js" -exec grep -l "OPUS_API_URL" {} \; | xargs -r sed -i "s#OPUS_API_URL#$OPUS_API_URL#g"

  # Also check for the escaped version which might appear in some files
  find "$client_dir" -type f -name "*.js" -exec grep -l "\"OPUS_API_URL\"" {} \; | xargs -r sed -i "s#\"OPUS_API_URL\"#\"$OPUS_API_URL\"#g"

  echo "Replaced OPUS_API_URL with $OPUS_API_URL"
else
  echo "WARNING: OPUS_API_URL environment variable is not set. API calls may fail."
fi

# Process OPUS_CLIENT_URL efficiently
if [ ! -z "$OPUS_CLIENT_URL" ]; then
  echo "Found OPUS_CLIENT_URL: $OPUS_CLIENT_URL"

  # Only process files that actually contain the string
  find "$client_dir" -type f -name "*.js" -exec grep -l "OPUS_CLIENT_URL" {} \; | xargs -r sed -i "s#OPUS_CLIENT_URL#$OPUS_CLIENT_URL#g"
  find "$client_dir" -type f -name "*.js" -exec grep -l "\"OPUS_CLIENT_URL\"" {} \; | xargs -r sed -i "s#\"OPUS_CLIENT_URL\"#\"$OPUS_CLIENT_URL\"#g"

  echo "Replaced OPUS_CLIENT_URL with $OPUS_CLIENT_URL"
fi

# Process any other OPUS_ prefixed environment variables (for future extensibility)
# Exclude the ones we've already processed
for key in $(env | grep '^OPUS_' | grep -v 'OPUS_API_URL\|OPUS_CLIENT_URL' | cut -d= -f1); do
  value=$(printenv "$key")

  if [ ! -z "$value" ]; then
    echo "Found $key: $value"

    # Only process files that contain this specific key
    find "$client_dir" -type f \( -name "*.js" -o -name "*.css" \) -exec grep -l "$key" {} \; | xargs -r sed -i "s#$key#$value#g"

    echo "Replaced $key with $value"
  fi
done

echo "Environment variable replacement complete"
