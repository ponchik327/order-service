#!/bin/sh
set -e  # Выход при ошибках

echo "Running migrations with DB_URL: $DB_URL"
goose -dir migrations postgres "$DB_URL" up

echo "Starting application..."
exec ./main