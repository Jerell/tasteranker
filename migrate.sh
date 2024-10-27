#!/bin/sh
echo "Running migrations..."
migrate -database "${DATABASE_URL}" -path /app/migrations up

