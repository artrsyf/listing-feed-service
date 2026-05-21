#!/bin/bash
set -e

echo "📋 Initializing PostgreSQL database..."

POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-benchmark}"
POSTGRES_USER="${POSTGRES_USER:-benchmark}"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "⚠️  psql not found in PATH, using docker-compose exec..."
    USE_DOCKER=true
fi

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
if [ "$USE_DOCKER" = true ]; then
    count=0
    while [ $count -lt 30 ]; do
        if docker-compose exec -T postgres pg_isready -U benchmark > /dev/null 2>&1; then
            echo "✅ PostgreSQL is ready"
            break
        fi
        echo "  Waiting for PostgreSQL... ($count/30)"
        sleep 2
        count=$((count+1))
    done
    
    if [ $count -eq 30 ]; then
        echo "❌ PostgreSQL is not ready. Try: make docker-up"
        exit 1
    fi
    
    # Run base DDL
    echo "Creating base schema..."
    docker-compose exec -T postgres psql -U benchmark -d benchmark -f /docker-entrypoint-initdb.d/base_ddl.sql
    
    # Run benchmark indexes
    echo "Creating benchmark indexes..."
    docker-compose exec -T postgres psql -U benchmark -d benchmark -f /docker-entrypoint-initdb.d/benchmark_indexes.sql
else
    until pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" > /dev/null 2>&1; do
        echo "  Waiting for PostgreSQL..."
        sleep 2
    done
    
    echo "✅ PostgreSQL is ready"
    
    # Run base DDL
    echo "Creating base schema..."
    PGPASSWORD=benchmark psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f postgres/init/base_ddl.sql
    
    # Run benchmark indexes
    echo "Creating benchmark indexes..."
    PGPASSWORD=benchmark psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f postgres/init/benchmark_indexes.sql
fi

echo ""
echo "✅ PostgreSQL initialization complete"
