#!/usr/bin/env bash
curl -v -X POST http://localhost:8080/signin \
  -H "Content-Type: application/json" \
  -d '{"phone_number":"1912561960","password":"supersecret"}'
