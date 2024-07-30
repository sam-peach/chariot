#! /bin/sh
set -e

docker-entrypoint.sh postgres &

until pg_isready -h localhost -p 5432; do
  echo "Waiting for Postgres to be ready..."
  sleep 1
done

for f in /migrations/*.sql; do
  echo "Running migration $f"
  psql -U $PGUSER -d $POSTGRES_DB -f "$f"
done

wait