export GOTOOLCHAIN := local

# Load .env if present and export its vars to recipe commands (go run reads os.Getenv).
-include .env
export

GOOSE_DRIVER := postgres
DATABASE_URL ?= postgres://monify:monify@localhost:5432/monify?sslmode=disable

.PHONY: up down run build migrate migrate-down seed-user api-key sqlc test lint tidy

up:
	docker compose up -d

down:
	docker compose down

run:
	go run ./cmd/monify

build:
	go build -o bin/monify ./cmd/monify

migrate:
	go run ./cmd/monify migrate

# down needs the goose CLI: go install github.com/pressly/goose/v3/cmd/goose@v3.22.1
migrate-down:
	goose -dir migrations $(GOOSE_DRIVER) "$(DATABASE_URL)" down

seed-user:
	go run ./cmd/monify seed-user

# Usage: make api-key EMAIL=me@example.com NAME="my integration"
api-key:
	go run ./cmd/monify apikey create "$(EMAIL)" "$(NAME)"

sqlc:
	sqlc generate

test:
	go test ./...

lint:
	golangci-lint run

tidy:
	go mod tidy
