#!/bin/sh

set -e

until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing migrations"

migrate -path ./migrations -database "$DATABASE_URL" up

>&2 echo "Migrations applied - starting the application"

exec ./main