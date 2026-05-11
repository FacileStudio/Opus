#!/bin/sh
set -e

echo "Starting environment variable replacement..."

# Process OPUS_API_URL first (with special handling)
if [ ! -z "$OPUS_API_URL" ]; then
  echo "Found OPUS_API_URL: $OPUS_API_URL"

  # First, replace the exact string "OPUS_API_URL" in all JavaScript files
  # Use grep -l to only process files that contain the string
  find /usr/share/nginx/html -type f -name "*.js" -exec grep -l "OPUS_API_URL" {} \; | xargs -r sed -i "s#OPUS_API_URL#$OPUS_API_URL#g"

  # Also check for the escaped version which might appear in some files
  find /usr/share/nginx/html -type f -name "*.js" -exec grep -l "\"OPUS_API_URL\"" {} \; | xargs -r sed -i "s#\"OPUS_API_URL\"#\"$OPUS_API_URL\"#g"

  # Build MCP OAuth discovery JSON for nginx to serve at /.well-known
  BASE_URL=$(echo "$OPUS_API_URL" | sed 's#/api/*$##')
  PRM_JSON="{\"resource\":\"${BASE_URL}/api/mcp\",\"authorization_servers\":[\"${BASE_URL}/api\"]}"
  AS_JSON="{\"issuer\":\"${BASE_URL}/api\",\"authorization_endpoint\":\"${BASE_URL}/api/mcp/authorize\",\"token_endpoint\":\"${BASE_URL}/api/mcp/token\",\"registration_endpoint\":\"${BASE_URL}/api/mcp/register\",\"response_types_supported\":[\"code\"],\"grant_types_supported\":[\"authorization_code\"],\"code_challenge_methods_supported\":[\"S256\"],\"token_endpoint_auth_methods_supported\":[\"none\"]}"
  sed -i "s#MCP_PRM_JSON_PLACEHOLDER#$PRM_JSON#g" /etc/nginx/conf.d/default.conf
  sed -i "s#MCP_AS_JSON_PLACEHOLDER#$AS_JSON#g" /etc/nginx/conf.d/default.conf

  echo "✅ Replaced OPUS_API_URL with $OPUS_API_URL"
else
  echo "WARNING: OPUS_API_URL environment variable is not set. API calls may fail."
  # No API URL — remove MCP placeholders so nginx doesn't serve broken JSON
  sed -i "s#MCP_PRM_JSON_PLACEHOLDER#{}#g" /etc/nginx/conf.d/default.conf
  sed -i "s#MCP_AS_JSON_PLACEHOLDER#{}#g" /etc/nginx/conf.d/default.conf
fi

# Process OPUS_CLIENT_URL efficiently
if [ ! -z "$OPUS_CLIENT_URL" ]; then
  echo "Found OPUS_CLIENT_URL: $OPUS_CLIENT_URL"
  
  # Only process files that actually contain the string
  find /usr/share/nginx/html -type f -name "*.js" -exec grep -l "OPUS_CLIENT_URL" {} \; | xargs -r sed -i "s#OPUS_CLIENT_URL#$OPUS_CLIENT_URL#g"
  find /usr/share/nginx/html -type f -name "*.js" -exec grep -l "\"OPUS_CLIENT_URL\"" {} \; | xargs -r sed -i "s#\"OPUS_CLIENT_URL\"#\"$OPUS_CLIENT_URL\"#g"
  
  echo "✅ Replaced OPUS_CLIENT_URL with $OPUS_CLIENT_URL"
fi

# Process any other OPUS_ prefixed environment variables (for future extensibility)
# Exclude the ones we've already processed
for key in $(env | grep '^OPUS_' | grep -v 'OPUS_API_URL\|OPUS_CLIENT_URL' | cut -d= -f1); do
  value=$(printenv "$key")
  
  if [ ! -z "$value" ]; then
    echo "Found $key: $value"
    
    # Only process files that contain this specific key
    find /usr/share/nginx/html -type f \( -name "*.js" -o -name "*.css" \) -exec grep -l "$key" {} \; | xargs -r sed -i "s#$key#$value#g"
    
    echo "✅ Replaced $key with $value"
  fi
done

echo "✅ Environment variable replacement complete"
