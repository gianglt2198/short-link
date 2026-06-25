.PHONY: run build test test-cover migrate-up migrate-down migrate-add db-up db-down clean

DATABASE_URL ?= postgres://shortlink:shortlink@localhost:5432/shortlink?sslmode=disable
CACHE__REDIS__ADDR ?= localhost:6379

run:
	go run ./cmd/server

build:
	go build -o bin/shortlink ./cmd/server

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out \
		-coverpkg=./internal/handlers/...,./internal/services/...,./internal/utils/...,./internal/helpers/...,./internal/infra/cache/... \
		./...
	grep -v "/mocks/" coverage.out > coverage_filtered.out
	go tool cover -html=coverage_filtered.out

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-add:
	migrate create -ext sql -dir migrations $(name)

infra-up:
	docker compose up -d

infra-down:
	docker compose down
