#!/bin/sh

set -e

check_env_var() {
  if [ -z "$1" ]; then
    echo "Error: $2 environment variable is not set." >&2
    exit 1
  fi
}

check_env_var "$POSTGRES_HOST" "POSTGRES_HOST"
check_env_var "$POSTGRES_PORT" "POSTGRES_PORT"
check_env_var "$POSTGRES_USER" "POSTGRES_USER"
check_env_var "$POSTGRES_DB" "POSTGRES_DB"
check_env_var "$POSTGRES_PASSWORD" "POSTGRES_PASSWORD"

if ! command -v pg_isready >/dev/null 2>&1; then
  echo "Error: pg_isready command not found. Please install PostgreSQL client utilities." >&2
  exit 1
fi

if ! command -v migrate >/dev/null 2>&1; then
  echo "Error: migrate command not found. Please install the migrate tool." >&2
  exit 1
fi

TIMEOUT=60
INTERVAL=1
elapsed=0

until pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep "$INTERVAL"
  elapsed=$((elapsed + INTERVAL))
  if [ "$elapsed" -ge "$TIMEOUT" ]; then
    >&2 echo "Postgres did not become ready within $TIMEOUT seconds."
    exit 1
  fi
done

>&2 echo "Postgres is up - executing migrations"

DATABASE_URL="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable"

if migrate -path ./migrations -database "$DATABASE_URL" up; then
  >&2 echo "Migrations applied successfully."
else
  >&2 echo "Error: Failed to apply migrations." >&2
  exit 1
fi

>&2 echo "Starting the application..."

exec ./main
