DBSTRING := "host=localhost user=postgres password=postgres dbname=ecommerce sslmode=disable"
MIGRATE := goose -dir migrations postgres $(DBSTRING)

.PHONY: all build dev up down psql migrate-create migrate-up migrate-down migrate-reset help

all: dev # watch and run on development environment

build: # build a binary executable
	go build -o ./tmp/main .

dev: up # watch and run on development environment
	air .

test:
	go test -v ./...

psql: # run psql
	docker compose exec -it db psql -U postgres -d ecommerce

migrate-create:
	$(MIGRATE) create $(name) sql

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down

migrate-reset:
	$(MIGRATE) reset

up: # start container
	docker compose up -d

down: # stop container
	docker compose down