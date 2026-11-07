#!/bin/sh
set -eu
file="${1:-package.json}"
version=$(sed -n 's/^[[:space:]]*"version":[[:space:]]*"\([^"]*\)".*/\1/p' "$file" | head -n 1)
if [ -z "$version" ]; then
    echo "version.sh: no version in $file" >&2
    exit 1
fi
printf 'v%s\n' "$version"
