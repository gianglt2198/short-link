.PHONY: run build test test-cover migrate-up migrate-down migrate-add db-up db-down clean \
        k6-encode k6-decode k6-all

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
	docker compose up -d postgres redis

infra-down:
	docker compose down

# k6 load tests — requires k6 installed (brew install k6)
# PROFILE=smoke|load|stress   BASE_URL=http://localhost:8080   POOL_SIZE=50
K6_FLAGS ?=
PROFILE  ?= smoke

k6-encode:
	k6 run $(K6_FLAGS) -e PROFILE=$(PROFILE) -e BASE_URL=$(BASE_URL) tests/k6/encode.js

k6-decode:
	k6 run $(K6_FLAGS) -e PROFILE=$(PROFILE) -e BASE_URL=$(BASE_URL) -e POOL_SIZE=$(POOL_SIZE) tests/k6/decode.js

k6-all: k6-encode k6-decode
