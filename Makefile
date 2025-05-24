# Colors for pretty output
YELLOW = \033[1;33m
GREEN = \033[1;32m
RED = \033[1;31m
BLUE = \033[1;34m
NC = \033[0m

.PHONY: dev-reader
dev-reader: ## Run the flight-reader locally with .env variables.
	@export $$(grep -v '^#' .env | xargs) && go run cmd/reader/main.go

.PHONY: dev-processor
dev-processor: ## Run the flight-processor locally with .env variables.
	@export $$(grep -v '^#' .env | xargs) && go run cmd/processor/main.go

.PHONY: dev-poster
dev-poster: ## Run the flight-poster locally with .env variables.
	@export $$(grep -v '^#' .env | xargs) && go run cmd/poster/main.go

.PHONY: docker-reader
docker-reader: ## Build the flight-reader Docker image.
	docker build -t flight-reader -f docker/reader.Dockerfile .

.PHONY: docker-processor
docker-processor: ## Build the flight-processor Docker image.
	docker build -t flight-processor -f docker/processor.Dockerfile .

.PHONY: docker-poster
docker-poster: ## Build the flight-poster Docker image.
	docker build -t flight-poster -f docker/poster.Dockerfile .

.PHONY: compose-up
compose-up: ## Start all services with build in foreground.
	docker compose up --build -d

.PHONY: compose-down
compose-down: ## Stop all services, remove volumes and orphans.
	docker compose down --volumes --remove-orphans

.PHONY: compose-logs
compose-logs: ## Show logs from all services.
	docker compose logs -f

.PHONY: lint
lint: ## Run linter.
	@echo "$(YELLOW)Running linter...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not installed$(NC)" && exit 1)
	golangci-lint run -c .golangci.yml
	@echo "$(GREEN)Linting completed!$(NC)"

.PHONY: help
help: ## Display this help screen.
	@echo "$(BLUE)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help
