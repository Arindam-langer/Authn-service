#!/usr/bin/env bash

ACCESS_TOKEN="${1}"

if [ -z "$ACCESS_TOKEN" ]; then
  echo "Usage: $0 <access_token>"
  echo "Example: $0 eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... (the Bearer token from sign-in)"
  exit 1
fi

curl -v -X POST http://localhost:8080/verify/token \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
