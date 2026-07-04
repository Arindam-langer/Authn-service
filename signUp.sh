#!/usr/bin/env bash

curl -v -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username": "aru", "password": "supersecret", "email": "arindamlanger@gmail.com", "phone_number":"01912561960"}'
