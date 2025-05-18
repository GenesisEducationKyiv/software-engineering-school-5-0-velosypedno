#!/bin/bash

echo "Running migrations"
migrate -database "${DB_DRIVER}://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" -path db/migrations up