#!/bin/sh

set -e

until pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing migrations"

migrate -path ./migrations -database "$DATABASE_URL" up

>&2 echo "Migrations applied - starting the application"

exec ./main