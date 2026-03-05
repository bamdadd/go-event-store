.PHONY: build test test-unit test-integration test-all lint db-start db-stop

build:
	go build ./...

test: test-unit

test-unit:
	go test -race ./...

test-integration:
	go test -race -tags=integration ./...

test-all: test-unit test-integration

lint:
	golangci-lint run ./...

db-start:
	docker-compose up -d postgres
	sleep 2

db-stop:
	docker-compose down
