COMPOSE ?= docker compose
CONFIG ?= config.yaml
OUT ?= ./output
WORKERS ?= 4
ITERATIONS ?= 20

.PHONY: help all up down reset generate init-db init-elastic load benchmark smoke clean ps logs

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
	@echo "  make reset        Remove containers and database volumes"
	@echo ""
	@echo "Parameters:"
	@echo "  CONFIG=config.yaml WORKERS=4 ITERATIONS=20 make all"

all: up generate init-db init-elastic load benchmark

smoke:
	$(MAKE) CONFIG=config.smoke.yaml ITERATIONS=3 WORKERS=2 all

up:
	$(COMPOSE) up -d --wait postgres elasticsearch

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

clean:
	$(COMPOSE) run --rm --no-deps benchmark-runner rm -rf $(OUT)

ps:
	$(COMPOSE) ps

logs:
	$(COMPOSE) logs --tail=200 postgres elasticsearch
