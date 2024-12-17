include .env

MIGRATIONS_PATH = ./migration

build:
	@go build -o bin/tcp-fs

run: build
	@./bin/tcp-fs

.PHONY: migration
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@, $(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down $(filter-out $@, $(MAKECMDGOALS))