.PHONY: up down build logs

up:
	docker compose -f deploy/docker-compose.yml up --build

down:
	docker compose -f deploy/docker-compose.yml down

build:
	docker compose -f deploy/docker-compose.yml build

logs:
	docker compose -f deploy/docker-compose.yml logs -f deepforge