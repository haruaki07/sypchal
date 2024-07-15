pwd := $(shell pwd)

.PHONY: build dev up down

all: dev

build:
	go build . -o main

dev: up
	air .

migrate-up: up
	migrate \
		-database postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable \
		-path migrations \
		up

migrate-down: up
	migrate \
		-database postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable \
		-path migrations \
		down

migrate-drop: up
	migrate \
		-database postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable \
		-path migrations \
		drop

up:
	docker compose up -d

down:
	docker compose down