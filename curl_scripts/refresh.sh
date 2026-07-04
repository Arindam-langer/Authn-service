#!/usr/bin/env bash

REFRESH_TOKEN="${1}"

if [ -z "$REFRESH_TOKEN" ]; then
  echo "Usage: $0 <refresh_token>"
  echo "Example: $0 8d5a86d... (the value of the RefreshToken cookie from sign-in)"
  exit 1
fi

curl -v -X POST http://localhost:8080/refresh \
  --cookie "RefreshToken=${REFRESH_TOKEN}"
