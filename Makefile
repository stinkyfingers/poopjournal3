# Variables
BINARY_NAME=poopjournal
LAMBDA_BINARY=bootstrap
GO_FILES=$(shell find . -name "*.go" | grep -v vendor/)

# Environment variables for local development
export AUTH0_DOMAIN ?= dev-domain.auth0.com
export AUTH0_CLIENT_ID ?= dev-client-id
export AUTH0_CLIENT_SECRET ?= dev-secret
export AUTH0_REDIRECT_URL ?= http://localhost:8080/callback
export STORAGE_TYPE ?= local
export DATA_DIR ?= ./data
export PORT ?= 8080

# Default target
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Download dependencies
	go mod download
	go mod tidy

.PHONY: build
build: ## Build the local binary
	@echo "Building local binary..."
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/$(BINARY_NAME) .

.PHONY: build-lambda
build-lambda: ## Build Lambda binary
	@echo "Building Lambda binary from project root for Go embed..."
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda -ldflags="-w -s" -o dist/$(LAMBDA_BINARY) .
	cd dist && zip lambda.zip $(LAMBDA_BINARY)

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/ dist/ data/

.PHONY: dev
dev: deps ## Run the application locally for development
	@echo "Starting local development server..."
	@echo "Visit http://localhost:$(PORT)"
	go run . 

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: lint
lint: ## Lint Go code (requires golangci-lint)
	golangci-lint run

.PHONY: vet
vet: ## Vet Go code
	go vet ./...

.PHONY: check
check: fmt vet lint test ## Run all checks (format, vet, lint, test)

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):latest .

.PHONY: docker-run
docker-run: docker-build ## Run Docker container
	docker run -p 8080:8080 \
		-e AUTH0_DOMAIN=$(AUTH0_DOMAIN) \
		-e AUTH0_CLIENT_ID=$(AUTH0_CLIENT_ID) \
		-e AUTH0_CLIENT_SECRET=$(AUTH0_CLIENT_SECRET) \
		-e AUTH0_REDIRECT_URL=http://localhost:8080/callback \
		-e STORAGE_TYPE=local \
		-e DATA_DIR=/app/data \
		-v $(PWD)/data:/app/data \
		$(BINARY_NAME):latest

# Terraform targets
.PHONY: tf-init
tf-init: ## Initialize Terraform
	cd terraform && terraform init

.PHONY: tf-plan
tf-plan: ## Plan Terraform deployment
	cd terraform && terraform plan

.PHONY: tf-apply
tf-apply: ## Apply Terraform deployment
	cd terraform && terraform apply

.PHONY: tf-destroy
tf-destroy: ## Destroy Terraform resources
	cd terraform && terraform destroy

.PHONY: tf-output
tf-output: ## Show Terraform outputs
	cd terraform && terraform output

# Deployment targets
.PHONY: deploy-lambda
deploy-lambda: build-lambda ## Build and deploy Lambda function
	cd terraform && terraform apply -auto-approve

.PHONY: deploy
deploy: check build-lambda deploy-lambda ## Full deployment (check, build, deploy)
	@echo "Deployment complete!"
	@$(MAKE) tf-output

# Development helpers
.PHONY: setup-dev
setup-dev: ## Set up development environment
	@echo "Setting up development environment..."
	@echo "1. Install dependencies..."
	$(MAKE) deps
	@echo "2. Create data directory..."
	mkdir -p data
	@echo "3. Copy example env file..."
	@if [ ! -f .env ]; then \
		echo "# Development environment variables" > .env; \
		echo "AUTH0_DOMAIN=dev-domain.auth0.com" >> .env; \
		echo "AUTH0_CLIENT_ID=dev-client-id" >> .env; \
		echo "AUTH0_CLIENT_SECRET=dev-secret" >> .env; \
		echo "AUTH0_REDIRECT_URL=http://localhost:8080/callback" >> .env; \
		echo "STORAGE_TYPE=local" >> .env; \
		echo "DATA_DIR=./data" >> .env; \
		echo "PORT=8080" >> .env; \
		echo "Created .env file with default values. Please update with your Auth0 credentials."; \
	fi
	@echo "Setup complete! Run 'make dev' to start the development server."

.PHONY: logs-lambda
logs-lambda: ## View Lambda function logs
	aws logs tail /aws/lambda/$(shell cd terraform && terraform output -raw lambda_function_name) --follow

.PHONY: invoke-lambda
invoke-lambda: ## Test invoke Lambda function
	aws lambda invoke \
		--function-name $(shell cd terraform && terraform output -raw lambda_function_name) \
		--payload '{"httpMethod":"GET","path":"/","headers":{},"body":""}' \
		response.json
	cat response.json | jq .

# Load testing
.PHONY: load-test
load-test: ## Run load tests (requires hey)
	@if ! command -v hey > /dev/null; then \
		echo "Installing hey..."; \
		go install github.com/rakyll/hey@latest; \
	fi
	hey -n 100 -c 10 http://localhost:$(PORT)

.PHONY: load-test-prod
load-test-prod: ## Run load tests against production
	@PROD_URL=$$(cd terraform && terraform output -raw api_gateway_url 2>/dev/null || terraform output -raw custom_domain_url 2>/dev/null); \
	if [ -n "$$PROD_URL" ]; then \
		hey -n 100 -c 10 "$$PROD_URL"; \
	else \
		echo "Production URL not found. Make sure Terraform is deployed."; \
	fi

# Database operations
.PHONY: backup-data
backup-data: ## Backup S3 data locally
	@BUCKET=$$(cd terraform && terraform output -raw s3_bucket_name 2>/dev/null); \
	if [ -n "$$BUCKET" ]; then \
		echo "Backing up data from S3 bucket: $$BUCKET"; \
		aws s3 sync s3://$$BUCKET ./backup/; \
		echo "Backup complete in ./backup/"; \
	else \
		echo "S3 bucket not found. Make sure Terraform is deployed."; \
	fi

.PHONY: restore-data
restore-data: ## Restore data to S3 from local backup
	@BUCKET=$$(cd terraform && terraform output -raw s3_bucket_name 2>/dev/null); \
	if [ -n "$$BUCKET" ]; then \
		echo "Restoring data to S3 bucket: $$BUCKET"; \
		aws s3 sync ./backup/ s3://$$BUCKET; \
		echo "Restore complete"; \
	else \
		echo "S3 bucket not found. Make sure Terraform is deployed."; \
	fi

# Maintenance
.PHONY: update-deps
update-deps: ## Update all dependencies
	go get -u ./...
	go mod tidy

.PHONY: security-check
security-check: ## Run security checks (requires gosec)
	@if ! command -v gosec > /dev/null; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...

# CI/CD helpers
.PHONY: ci-test
ci-test: deps check ## Run tests for CI
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: ci-build
ci-build: ## Build for CI
	$(MAKE) build
	$(MAKE) build-lambda

# All-in-one targets
.PHONY: all
all: clean deps check build build-lambda ## Build everything

.PHONY: quick
quick: build ## Quick build for development