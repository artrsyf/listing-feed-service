COMPOSE ?= docker compose
CONFIG ?= config.yaml
OUT ?= ./output
WORKERS ?= 4
ITERATIONS ?= 20

LISTING_CONFIG ?= config.listing.yaml
PHOTOS_CONFIG ?= config.photos.yaml
INDEX_CONFIG ?= config.index.large.yaml

.PHONY: help all up down reset generate init-db init-elastic load benchmark smoke clean ps logs listing listing-large listing-generate listing-init listing-load listing-benchmark photos photos-large photos-init photos-load photos-benchmark index-joins-large index-joins-benchmark

help:
	@echo "PostgreSQL vs Elasticsearch benchmark"
	@echo ""
	@echo "Main commands:"
	@echo "  make all          Start containers, generate data, initialize, load, benchmark"
	@echo "  make smoke        Same workflow with config.smoke.yaml"
	@echo "  make up           Start PostgreSQL and Elasticsearch"
	@echo "  make generate     Generate CSV and NDJSON data in containers"
	@echo "  make init-db      Recreate PostgreSQL schema"
	@echo "  make init-elastic Recreate Elasticsearch index"
	@echo "  make load         Load generated data"
	@echo "  make benchmark    Run read benchmarks"
	@echo "  make listing      Run marketplace/listing read-model benchmark"
	@echo "  make listing-large Run large marketplace/listing benchmark"
	@echo "  make photos       Run PostgreSQL bytea vs MinIO photo benchmark"
	@echo "  make photos-large Run large PostgreSQL bytea vs MinIO benchmark"
	@echo "  make index-joins-large Run isolated PostgreSQL index/join benchmark"
	@echo "  make reset        Remove containers and database volumes"
	@echo ""
	@echo "Parameters:"
	@echo "  CONFIG=config.yaml WORKERS=4 ITERATIONS=20 make all"

all: up generate init-db init-elastic load benchmark

smoke:
	$(MAKE) CONFIG=config.smoke.yaml ITERATIONS=3 WORKERS=2 all

up:
	$(COMPOSE) up -d --wait postgres elasticsearch minio

down:
	$(COMPOSE) down

reset:
	$(COMPOSE) down -v --remove-orphans

generate:
	$(COMPOSE) run --rm benchmark-runner go run ./generator/cmd -config $(CONFIG) -out $(OUT)

init-db:
	$(COMPOSE) exec -T postgres psql -U benchmark -d benchmark -v ON_ERROR_STOP=1 -f /docker-entrypoint-initdb.d/base_ddl.sql
	$(COMPOSE) exec -T postgres psql -U benchmark -d benchmark -v ON_ERROR_STOP=1 -f /docker-entrypoint-initdb.d/benchmark_indexes.sql

init-elastic:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -init-elastic

load:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -load

benchmark:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -benchmark -workers $(WORKERS) -iterations $(ITERATIONS)

listing: up listing-generate listing-init listing-load listing-benchmark

listing-large:
	$(MAKE) LISTING_CONFIG=config.listing.large.yaml ITERATIONS=50 WORKERS=4 listing

listing-generate:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -generate-listings -listings-config $(LISTING_CONFIG)

listing-init:
	$(COMPOSE) exec -T postgres psql -U benchmark -d benchmark -v ON_ERROR_STOP=1 -f /docker-entrypoint-initdb.d/listing_ddl.sql
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -init-listings

listing-load:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -load-listings

listing-benchmark:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -benchmark-listings -workers $(WORKERS) -iterations $(ITERATIONS)

photos: up photos-init photos-load photos-benchmark

photos-large:
	$(MAKE) PHOTOS_CONFIG=config.photos.large.yaml ITERATIONS=200 WORKERS=8 photos

photos-init:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -init-photos

photos-load:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -load-photos -photos-config $(PHOTOS_CONFIG)

photos-benchmark:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -benchmark-photos -photos-config $(PHOTOS_CONFIG) -workers $(WORKERS) -iterations $(ITERATIONS)

index-joins-large:
	$(MAKE) CONFIG=$(INDEX_CONFIG) ITERATIONS=20 WORKERS=4 up generate init-db index-load-postgres index-joins-benchmark

index-load-postgres:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -load-postgres

index-joins-benchmark:
	$(COMPOSE) run --rm benchmark-runner go run ./cmd -benchmark-index-joins -workers $(WORKERS) -iterations $(ITERATIONS)

clean:
	$(COMPOSE) run --rm --no-deps benchmark-runner rm -rf $(OUT)

ps:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs --tail=200 postgres elasticsearch
