.PHONY: help build test lint clean docker-build docker-up docker-down healthcheck enroll logs

# Variables
SERVICES = idp-mock ca-service enrollment-agent protected-service diagnostics-api
GO := $(shell which go)
DOCKER := $(shell which docker)

help:
	@echo "StepDeploy Lab - Build Targets"
	@echo ""
	@echo "Local Development:"
	@echo "  make build          - Build all Go services"
	@echo "  make test           - Run all tests"
	@echo "  make lint           - Run linters (golangci-lint)"
	@echo "  make run-local      - Run all services locally"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   - Build all Docker images"
	@echo "  make docker-up      - Start all services (docker-compose up -d)"
	@echo "  make docker-down    - Stop all services (docker-compose down)"
	@echo "  make docker-logs    - View docker-compose logs"
	@echo ""
	@echo "Testing & Diagnostics:"
	@echo "  make healthcheck    - Run health checks on all services"
	@echo "  make enroll         - Test enrollment flow"
	@echo "  make inject-failure - Inject test failure (set FAILURE=<type>)"
	@echo "  make reset          - Reset environment to clean state"
	@echo "  make logs           - Collect logs for debugging"
	@echo ""
	@echo "Infrastructure:"
	@echo "  make terraform-plan  - Plan terraform deployment"
	@echo "  make terraform-apply - Apply terraform deployment"
	@echo "  make terraform-destroy - Destroy terraform deployment"
	@echo ""

# Build
build: $(SERVICES)

$(SERVICES):
	@echo "Building $@..."
	@mkdir -p services/$@/bin
	@cd services/$@ && go build -o bin/$(basename $@) ./cmd/main.go

test:
	@echo "Running tests..."
	@for dir in services/*/; do \
		if [ -d "$$dir" ]; then \
			echo "Testing $$(basename $$dir)..."; \
			cd "$$dir" && go test ./... -v && cd - > /dev/null; \
		fi \
	done
	@echo "✓ All tests passed"

lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run ./...
	@echo "✓ Lint passed"

clean:
	@echo "Cleaning build artifacts..."
	@for dir in services/*/; do \
		if [ -d "$$dir" ]; then \
			rm -rf "$$dir/bin"; \
		fi \
	done
	@echo "✓ Clean complete"

# Local development
run-local: build
	@echo "Starting services locally..."
	@echo "Note: Run each in separate terminal:"
	@echo "  cd services/idp-mock && ./bin/idp-mock"
	@echo "  cd services/ca-service && ./bin/ca-service"
	@echo "  cd services/enrollment-agent && ./bin/enrollment-agent"
	@echo "  cd services/protected-service && ./bin/protected-service"
	@echo "  cd services/diagnostics-api && ./bin/diagnostics-api"
	@echo "  cd frontend && npm run dev"

# Docker
docker-build:
	@echo "Building Docker images..."
	@docker compose build --no-cache
	@echo "✓ Docker images built"

docker-up:
	@echo "Starting services with Docker Compose..."
	@docker compose up -d
	@echo "Waiting for services to start..."
	@sleep 5
	@$(MAKE) healthcheck

docker-down:
	@echo "Stopping services..."
	@docker compose down

docker-logs:
	@docker compose logs -f

# Testing & Diagnostics
healthcheck:
	@./scripts/run-healthcheck.sh

enroll:
	@echo "Testing enrollment flow..."
	@curl -s http://localhost:8003/enroll | jq '.'

inject-failure:
	@if [ -z "$(FAILURE)" ]; then \
		echo "Usage: make inject-failure FAILURE=<type>"; \
		echo "Available types: expired-oauth-token, wrong-oauth-audience, bad-client-cert, service-unavailable"; \
	else \
		./scripts/inject-failure.sh "$(FAILURE)"; \
	fi

reset:
	@./scripts/reset-environment.sh

logs:
	@./scripts/collect-logs.sh

# Infrastructure
terraform-plan:
	@echo "Planning Terraform deployment..."
	@cd infra/terraform && terraform init && terraform plan

terraform-apply:
	@echo "Applying Terraform deployment..."
	@cd infra/terraform && terraform init && terraform apply

terraform-destroy:
	@echo "Destroying Terraform deployment..."
	@cd infra/terraform && terraform destroy

# CI/CD
github-actions-local:
	@echo "Running GitHub Actions locally (requires act)"
	@which act > /dev/null || (echo "Install act: https://github.com/nektos/act" && exit 1)
	@act

fmt:
	@echo "Formatting Go code..."
	@gofmt -s -w .
	@echo "✓ Formatted"

vendor:
	@echo "Downloading dependencies..."
	@for dir in services/*/; do \
		if [ -d "$$dir" ]; then \
			cd "$$dir" && go mod download && go mod tidy && cd - > /dev/null; \
		fi \
	done
	@echo "✓ Dependencies downloaded"

dev: docker-up
	@echo "✓ Development environment ready"
	@echo "  Frontend: http://localhost:5173"
	@echo "  Diagnostics: http://localhost:8080"
	@echo "  Services: 8001-8004"

status:
	@docker compose ps

shell-%:
	@docker compose exec $* /bin/sh

# Convenience targets
all: lint test docker-build docker-up healthcheck
	@echo "✓ All targets completed successfully"

.PHONY: $(SERVICES)
