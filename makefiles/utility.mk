##@ QoL - Helpers for QoL commands

.PHONY: format
format: ## Format all backend code
	@echo "Formatting all code"
	@cd $(SERVER_DIR) && go tool goimports -w .
	@cd $(SERVER_DIR) && go tool golines -w .
	@cd $(SERVER_DIR) && gofmt -w .
	@echo "Formatting completed!"

.PHONY: lint
lint: lint-format lint-arch lint-linter

.PHONY: lint-format
lint-format: ## Format code
	@echo "Formatting all code"
	@cd $(SERVER_DIR) && go tool goimports -w .
	@cd $(SERVER_DIR) && go tool golines -w .
	@cd $(SERVER_DIR) && gofmt -w .
	@echo "Formatting completed!"

.PHONY: lint-linter
lint-linter: ## Run golangci-lint
	@echo "Linting with golangci-lint..."
	@cd $(SERVER_DIR) && golangci-lint run --color always
	@echo "Linting completed!"

.PHONY: lint-arch
lint-arch: ## Run architecture tests
	@echo "Running go-arch-lint..."
	@bash $(SERVER_DIR)/scripts/arch-lint-wrapper.sh
	@echo "Architecture lint passed!"

.PHONY: lint-legacy
lint-legacy: ## Legacy lint command
	@echo "Linting all code"
	@cd $(SERVER_DIR) && go tool goimports -w .
	@cd $(SERVER_DIR) && go tool golines -w .
	@cd $(SERVER_DIR) && gofmt -w .
	@cd $(SERVER_DIR) && \
    mkdir -p .cache && \
    docker run --rm -t -v $$(pwd):/app -w /app \
        --user $$(id -u):$$(id -g) \
        -v $$(go env GOCACHE):/tmp/go-build -e GOCACHE=/tmp/go-build \
        -v $$(go env GOMODCACHE):/tmp/mod -e GOMODCACHE=/tmp/mod \
        -v $$(pwd)/.cache:/.cache \
        -e GOLANGCI_LINT_CACHE=/tmp/golangci-lint \
        -e HOME=/tmp \
        golangci/golangci-lint:latest golangci-lint run
	@echo "Running convention checks..."
	@.github/scripts/run-checks.sh HEAD~1 | .github/scripts/format-terminal.sh

.PHONY: test
test: ## Runs tests for the codebase
	@echo "Running tests"
	@$(call use_env,test) && (cd $(SERVER_DIR) && \
    go test --cover -covermode=count -coverpkg=./... ./test/... -coverprofile=cover.out && \
    go tool cover -html=cover.out)

