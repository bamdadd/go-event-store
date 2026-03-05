.DEFAULT_GOAL := help
.PHONY: help build test test-unit test-integration test-all lint db-start db-stop

help: ## Show this help
	@printf '\n\033[1m%-20s %s\033[0m\n\n' 'Target' 'Description'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

build: ## Build all packages
	go build ./...

test: test-unit ## Run unit tests (alias)

test-unit: ## Run unit tests with race detector
	go test -race ./...

test-integration: ## Run integration tests (requires Docker)
	go test -race -tags=integration ./...

test-all: test-unit test-integration ## Run all tests

lint: ## Run golangci-lint
	golangci-lint run ./...

db-start: ## Start PostgreSQL via docker-compose
	docker-compose up -d postgres
	sleep 2

db-stop: ## Stop PostgreSQL
	docker-compose down
