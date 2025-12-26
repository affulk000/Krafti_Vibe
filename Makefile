.PHONY: help build run run-race dev test test-unit test-integration test-coverage clean lint fmt check docker-build docker-run migrate-up migrate-down migrate-create migrate-status deps install-tools k8s-validate k8s-preview-dev k8s-preview-prod k8s-dev k8s-prod k8s-status-dev k8s-status-prod k8s-logs-dev k8s-logs-prod k8s-delete-dev k8s-delete-prod k8s-shell-dev k8s-shell-prod k8s-port-forward-dev k8s-port-forward-prod deploy-dev deploy-prod

# Variables
APP_NAME=kraftivibe
BINARY_NAME=kraftivibe
MAIN_PATH=./cmd/api/main.go
DOCKER_IMAGE=$(APP_NAME):latest
DOCKER_REGISTRY=your-registry.com
GO_VERSION=1.24

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "$(GREEN)Krafti Vibe - Available Commands$(NC)"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed -e 's/^##//' | awk 'BEGIN {FS = ":"}; {printf "$(YELLOW)%-30s$(NC) %s\n", $$1, $$2}'
	@echo ""

## build: Build the application binary
build:
	@echo "$(GREEN)Building application...$(NC)"
	@go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: bin/$(BINARY_NAME)$(NC)"

## run: Run the application
run: build
	@echo "$(GREEN)Running application...$(NC)"
	@./bin/$(BINARY_NAME)

## dev: Run with hot reload (requires air)
dev:
	@echo "$(GREEN)Starting development server with hot reload...$(NC)"
	@air -c air.toml

## run-race: Run with race detector
run-race:
	@echo "$(GREEN)Running application with race detector...$(NC)"
	@go run -race $(MAIN_PATH)

## test: Run all tests
test:
	@echo "$(GREEN)Running all tests...$(NC)"
	@go test -v -race ./...

## test-unit: Run unit tests only
test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	@go test -v -race -short ./...

## test-integration: Run integration tests only
test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	@go test -v -race -run Integration ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

## test-coverage-func: Show coverage by function
test-coverage-func:
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out

## bench: Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...

## lint: Run linters
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not installed. Run: make install-tools$(NC)" && exit 1)
	@golangci-lint run --timeout=5m ./...

## fmt: Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@goimports -w .

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(GREEN)All checks passed!$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@rm -rf tmp/
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache
	@echo "$(GREEN)Clean complete$(NC)"

## deps: Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify

## tidy: Tidy dependencies
tidy:
	@echo "$(GREEN)Tidying dependencies...$(NC)"
	@go mod tidy

## install-tools: Install development tools
install-tools:
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)Tools installed$(NC)"

## docker-build: Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE)$(NC)"

## docker-run: Run Docker container
docker-run:
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run -p 3000:3000 --env-file .env $(DOCKER_IMAGE)

## docker-push: Push Docker image to registry
docker-push: docker-build
	@echo "$(GREEN)Pushing Docker image to registry...$(NC)"
	@docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)

## docker-compose-up: Start all services with docker-compose
docker-compose-up:
	@echo "$(GREEN)Starting services with docker-compose...$(NC)"
	@docker-compose up -d

## docker-compose-down: Stop all services
docker-compose-down:
	@echo "$(GREEN)Stopping services...$(NC)"
	@docker-compose down

## docker-compose-logs: View logs from all services
docker-compose-logs:
	@docker-compose logs -f

## docker-compose-build: Build and start services
docker-compose-build:
	@echo "$(GREEN)Building and starting services...$(NC)"
	@docker-compose up -d --build

## migrate-create: Create a new migration (usage: make migrate-create name=create_users_table)
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: name is required. Usage: make migrate-create name=create_users_table$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating migration: $(name)$(NC)"
	@mkdir -p migrations
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	up_file="migrations/$${timestamp}_$(name).up.sql"; \
	down_file="migrations/$${timestamp}_$(name).down.sql"; \
	echo "-- Migration: $(name)" > $$up_file; \
	echo "-- Created at: $$(date)" >> $$up_file; \
	echo "" >> $$up_file; \
	echo "-- Add your UP migration SQL here" >> $$up_file; \
	echo "" >> $$up_file; \
	echo "-- Migration: $(name)" > $$down_file; \
	echo "-- Created at: $$(date)" >> $$down_file; \
	echo "" >> $$down_file; \
	echo "-- Add your DOWN migration SQL here" >> $$down_file; \
	echo "" >> $$down_file; \
	echo "$(GREEN)Created migration files:$(NC)"; \
	echo "  - $$up_file"; \
	echo "  - $$down_file"

## migrate-build: Build migration tool
migrate-build:
	@echo "$(GREEN)Building migration tool...$(NC)"
	@go build -o bin/migrate ./cmd/migrate

## migrate-up: Run database migrations up
migrate-up: migrate-build
	@echo "$(GREEN)Running migrations up...$(NC)"
	@./bin/migrate -action=up

## migrate-down: Rollback last migration
migrate-down: migrate-build
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	@./bin/migrate -action=down

## migrate-status: Show migration status
migrate-status: migrate-build
	@echo "$(GREEN)Checking migration status...$(NC)"
	@./bin/migrate -action=status

## migrate-seed: Seed initial data
migrate-seed: migrate-build
	@echo "$(GREEN)Seeding database...$(NC)"
	@./bin/migrate -action=seed

## migrate-up-dry: Run migrations in dry-run mode
migrate-up-dry: migrate-build
	@echo "$(GREEN)Running migrations (dry-run)...$(NC)"
	@./bin/migrate -action=up -dry-run=true

## migrate-up-force: Force run migrations
migrate-up-force: migrate-build
	@echo "$(YELLOW)Force running migrations...$(NC)"
	@./bin/migrate -action=up -force=true

## db-reset: Drop and recreate database (DANGEROUS!)
db-reset:
	@echo "$(RED)WARNING: This will drop and recreate the database!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "$(YELLOW)Dropping database...$(NC)"; \
		docker-compose exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS krafti_vibe;"; \
		docker-compose exec postgres psql -U postgres -c "CREATE DATABASE krafti_vibe;"; \
		echo "$(GREEN)Database recreated. Running migrations...$(NC)"; \
		$(MAKE) migrate-up; \
	fi

## db-seed: Seed database with sample data
db-seed:
	@echo "$(GREEN)Seeding database...$(NC)"
	@go run scripts/seed.go

## db-backup: Backup database
db-backup:
	@echo "$(GREEN)Backing up database...$(NC)"
	@mkdir -p backups
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	docker-compose exec -T postgres pg_dump -U postgres krafti_vibe > backups/backup_$$timestamp.sql; \
	echo "$(GREEN)Backup created: backups/backup_$$timestamp.sql$(NC)"

## db-restore: Restore database from backup (usage: make db-restore file=backups/backup_20240115_120000.sql)
db-restore:
	@if [ -z "$(file)" ]; then \
		echo "$(RED)Error: file is required. Usage: make db-restore file=backups/backup_20240115_120000.sql$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Restoring database from $(file)...$(NC)"
	@docker-compose exec -T postgres psql -U postgres krafti_vibe < $(file)
	@echo "$(GREEN)Database restored$(NC)"

## swagger: Generate Swagger documentation
swagger:
	@echo "$(GREEN)Generating Swagger documentation...$(NC)"
	@swag init -g cmd/api/main.go -o docs/swagger

## proto: Generate protobuf code (if using gRPC)
proto:
	@echo "$(GREEN)Generating protobuf code...$(NC)"
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

## mock: Generate mocks for testing
mock:
	@echo "$(GREEN)Generating mocks...$(NC)"
	@go generate ./...

## security: Run security checks
security:
	@echo "$(GREEN)Running security checks...$(NC)"
	@which gosec > /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	@gosec -fmt=json -out=security-report.json ./...
	@echo "$(GREEN)Security report generated: security-report.json$(NC)"

## audit: Audit dependencies for vulnerabilities
audit:
	@echo "$(GREEN)Auditing dependencies...$(NC)"
	@go list -json -m all | nancy sleuth

## update-deps: Update all dependencies to latest
update-deps:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

## version: Show application version
version:
	@echo "$(GREEN)Krafti Vibe API$(NC)"
	@echo "Go version: $(GO_VERSION)"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Git branch: $$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build date: $$(date -u '+%Y-%m-%d %H:%M:%S UTC')"

## ci: Run CI pipeline locally
ci: deps check test-coverage security
	@echo "$(GREEN)CI pipeline completed successfully!$(NC)"

## setup: Initial setup for development environment
setup: deps install-tools
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@cp -n .env.example .env 2>/dev/null || true
	@echo "$(GREEN)Setup complete! Edit .env file with your configuration.$(NC)"
	@echo "$(YELLOW)Run 'make docker-compose-up' to start services$(NC)"

## logs: Tail application logs
logs:
	@docker-compose logs -f api

## ps: Show running containers
ps:
	@docker-compose ps

## shell: Open shell in API container
shell:
	@docker-compose exec api sh

## db-shell: Open PostgreSQL shell
db-shell:
	@docker-compose exec postgres psql -U postgres -d krafti_vibe

## redis-cli: Open Redis CLI
redis-cli:
	@docker-compose exec redis redis-cli -a redis_password

## watch-test: Watch and run tests on file changes
watch-test:
	@echo "$(GREEN)Watching for changes and running tests...$(NC)"
	@which reflex > /dev/null || go install github.com/cespare/reflex@latest
	@reflex -r '\.go$$' -s -- sh -c 'make test'

## perf: Run performance profiling
perf:
	@echo "$(GREEN)Starting performance profiling...$(NC)"
	@go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "$(GREEN)Profiles generated: cpu.prof, mem.prof$(NC)"
	@echo "View with: go tool pprof cpu.prof"

## size: Show binary size
size: build
	@ls -lh bin/$(BINARY_NAME)

## all: Build everything
all: clean deps check build docker-build
	@echo "$(GREEN)Build complete!$(NC)"

##
## Kubernetes Commands
##

## k8s-validate: Validate Kubernetes manifests
k8s-validate:
	@echo "$(GREEN)Validating Kubernetes manifests...$(NC)"
	@kubectl kustomize k8s/overlays/dev > /tmp/k8s-dev.yaml
	@kubectl kustomize k8s/overlays/prod > /tmp/k8s-prod.yaml
	@echo "$(GREEN)Validation successful!$(NC)"

## k8s-preview-dev: Preview development Kubernetes manifests
k8s-preview-dev:
	@echo "$(GREEN)Development manifests:$(NC)"
	@kubectl kustomize k8s/overlays/dev

## k8s-preview-prod: Preview production Kubernetes manifests
k8s-preview-prod:
	@echo "$(GREEN)Production manifests:$(NC)"
	@kubectl kustomize k8s/overlays/prod

## k8s-dev: Deploy to Kubernetes development environment
k8s-dev:
	@echo "$(GREEN)Deploying to Kubernetes (dev)...$(NC)"
	@kubectl apply -k k8s/overlays/dev
	@echo "$(GREEN)Deployment started!$(NC)"
	@echo "$(YELLOW)Waiting for pods to be ready...$(NC)"
	@kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=kraftivibe -n kraftivibe-dev --timeout=300s || true

## k8s-prod: Deploy to Kubernetes production environment
k8s-prod:
	@echo "$(GREEN)Deploying to Kubernetes (prod)...$(NC)"
	@kubectl apply -k k8s/overlays/prod
	@echo "$(GREEN)Deployment started!$(NC)"
	@echo "$(YELLOW)Waiting for pods to be ready...$(NC)"
	@kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=kraftivibe -n kraftivibe --timeout=300s || true

## k8s-status-dev: Check Kubernetes development status
k8s-status-dev:
	@echo "$(GREEN)Development environment status:$(NC)"
	@kubectl get all -n kraftivibe-dev

## k8s-status-prod: Check Kubernetes production status
k8s-status-prod:
	@echo "$(GREEN)Production environment status:$(NC)"
	@kubectl get all -n kraftivibe

## k8s-logs-dev: View Kubernetes development logs
k8s-logs-dev:
	@kubectl logs -n kraftivibe-dev -l app=kraftivibe-api -f

## k8s-logs-prod: View Kubernetes production logs
k8s-logs-prod:
	@kubectl logs -n kraftivibe -l app=kraftivibe-api -f

## k8s-delete-dev: Delete Kubernetes development environment
k8s-delete-dev:
	@echo "$(YELLOW)Deleting development environment...$(NC)"
	@kubectl delete -k k8s/overlays/dev

## k8s-delete-prod: Delete Kubernetes production environment
k8s-delete-prod:
	@echo "$(RED)WARNING: This will delete the production environment!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		kubectl delete -k k8s/overlays/prod; \
	fi

## k8s-shell-dev: Open shell in development API pod
k8s-shell-dev:
	@kubectl exec -it -n kraftivibe-dev $$(kubectl get pod -n kraftivibe-dev -l app=kraftivibe-api -o jsonpath='{.items[0].metadata.name}') -- /bin/sh

## k8s-shell-prod: Open shell in production API pod
k8s-shell-prod:
	@kubectl exec -it -n kraftivibe $$(kubectl get pod -n kraftivibe -l app=kraftivibe-api -o jsonpath='{.items[0].metadata.name}') -- /bin/sh

## k8s-port-forward-dev: Port forward development API to localhost:3000
k8s-port-forward-dev:
	@echo "$(GREEN)Port forwarding dev API to localhost:3000...$(NC)"
	@kubectl port-forward -n kraftivibe-dev svc/dev-api-service 3000:3000

## k8s-port-forward-prod: Port forward production API to localhost:3000
k8s-port-forward-prod:
	@echo "$(GREEN)Port forwarding prod API to localhost:3000...$(NC)"
	@kubectl port-forward -n kraftivibe svc/api-service 3000:3000

## deploy-dev: Build, push, and deploy to development
deploy-dev: docker-build docker-push k8s-dev
	@echo "$(GREEN)Deployment to development completed!$(NC)"

## deploy-prod: Build, push, and deploy to production
deploy-prod: docker-build docker-push k8s-prod
	@echo "$(GREEN)Deployment to production completed!$(NC)"
