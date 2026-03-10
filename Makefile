.PHONY: up down build logs

up:
	docker compose --env-file .env.docker -f deploy/docker-compose.yml up --build

down:
	docker compose --env-file .env.docker -f deploy/docker-compose.yml down

build:
	docker compose --env-file .env.docker -f deploy/docker-compose.yml build

logs:
	docker compose --env-file .env.docker -f deploy/docker-compose.yml logs -f deepforge-app