COMPOSE_FILE=./docker/docker-compose.yml

.PHONY: up down clean

up:
	docker compose -f $(COMPOSE_FILE) up --build

down:
	docker compose -f $(COMPOSE_FILE) down

clean:
	docker compose -f $(COMPOSE_FILE) down -v --remove-orphans

	docker system prune
	docker volume prune
	docker network prune
	docker image prune