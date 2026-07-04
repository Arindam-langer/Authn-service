#!/usr/bin/env bash

ACCESS_TOKEN="${1}"
REFRESH_TOKEN="${2}"

if [ -z "$ACCESS_TOKEN" ] || [ -z "$REFRESH_TOKEN" ]; then
  echo "Usage: $0 <access_token> <refresh_token>"
  echo "Example: $0 eyJhbGciOiJIUzI1Ni... 8d5a86d..."
  exit 1
fi

curl -v -X POST http://localhost:8080/signout \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  --cookie "RefreshToken=${REFRESH_TOKEN}"
