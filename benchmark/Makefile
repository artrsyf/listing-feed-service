.PHONY: help generate load benchmark clean docker-up docker-down init-db init-elastic

help:
	@echo "Benchmark Toolkit - PostgreSQL vs Elasticsearch"
	@echo ""
	@echo "Quick Start:"
	@echo "  make all           - Full workflow (docker-up + generate + load + benchmark)"
	@echo ""
	@echo "Usage:"
	@echo "  make generate      - Generate test data"
	@echo "  make load          - Load data into databases"
	@echo "  make benchmark     - Run benchmarks"
	@echo "  make docker-up     - Start PostgreSQL and Elasticsearch"
	@echo "  make docker-down   - Stop containers"
	@echo "  make clean         - Remove generated data"
	@echo "  make init-db       - Initialize PostgreSQL schema"
	@echo "  make init-elastic  - Initialize Elasticsearch index"
	@echo ""

generate:
	@echo "📊 Generating test data..."
	@mkdir -p ./output
	go run generator/cmd/main.go -config config.yaml -out ./output

load:
	@echo "📥 Loading data into databases..."
	go run cmd/main.go -load

benchmark:
	@echo "🏃 Running benchmarks..."
	go run cmd/main.go -benchmark

docker-up:
	@echo "🐳 Starting containers..."
	docker-compose up -d
	@echo ""
	@echo "Waiting for services to be ready..."
	@echo "This may take 1-2 minutes..."
	@count=0; while [ $$count -lt 30 ]; do \
		if docker-compose exec -T postgres pg_isready -U benchmark > /dev/null 2>&1; then \
			echo "✅ PostgreSQL is ready"; \
			break; \
		fi; \
		echo "  Waiting for PostgreSQL... ($$count/30)"; \
		sleep 2; \
		count=$$((count+1)); \
	done
	@count=0; while [ $$count -lt 30 ]; do \
		if curl -s http://localhost:9200 > /dev/null 2>&1; then \
			echo "✅ Elasticsearch is ready"; \
			break; \
		fi; \
		echo "  Waiting for Elasticsearch... ($$count/30)"; \
		sleep 2; \
		count=$$((count+1)); \
	done
	@echo ""
	@echo "✅ All services are up and running"
	@docker-compose ps

docker-down:
	@echo "🛑 Stopping containers..."
	docker-compose down

clean:
	@echo "🧹 Cleaning generated data..."
	rm -rf ./output
	@echo "✅ Clean complete"

# Full workflow: start containers, generate data, load, and run benchmarks
all: docker-up generate init-db init-elastic load benchmark

# Initialize PostgreSQL database schema and indexes
init-db:
	@echo "📋 Initializing PostgreSQL database..."
	@echo "Creating base schema..."
	docker-compose exec -T postgres psql -U benchmark -d benchmark -f /docker-entrypoint-initdb.d/base_ddl.sql
	@echo "Creating benchmark indexes..."
	docker-compose exec -T postgres psql -U benchmark -d benchmark -f /docker-entrypoint-initdb.d/benchmark_indexes.sql
	@echo ""
	@echo "✅ PostgreSQL initialization complete"
	@echo ""
	@echo "📊 Table sizes:"
	docker-compose exec -T postgres psql -U benchmark -d benchmark -c "SELECT relname as table, pg_size_pretty(pg_total_relation_size(relid)) as size FROM pg_catalog.pg_statio_user_tables ORDER BY pg_total_relation_size(relid) DESC;"

# Initialize Elasticsearch index
init-elastic:
	@echo "📋 Initializing Elasticsearch index..."
	@./scripts/init-elastic.sh

# Show help for first-time users
first-run:
	@echo "=============================================="
	@echo "  First-Time Setup Guide"
	@echo "=============================================="
	@echo ""
	@echo "1. Start infrastructure:"
	@echo "   make docker-up"
	@echo ""
	@echo "2. Generate test data:"
	@echo "   make generate"
	@echo ""
	@echo "3. Initialize databases:"
	@echo "   make init-db"
	@echo "   make init-elastic"
	@echo ""
	@echo "4. Load data:"
	@echo "   make load"
	@echo ""
	@echo "5. Run benchmarks:"
	@echo "   make benchmark"
	@echo ""
	@echo "Or run everything at once:"
	@echo "   make all"
	@echo ""
	@echo "=============================================="
