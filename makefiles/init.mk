##@ Initialization - Helpers to set up the application from the beginning

.PHONY: init
init: ## Initialize the project
	@echo "Starting project initialization..."
	@$(MAKE) init-env
	@$(MAKE) init-db
	@echo "Project initialization complete."

.PHONY: init-env
init-env: ## Initialize .env files
	@echo
	@echo "------------------------------"
	@echo "Initializing .env files "
	@echo
	@echo "--- [Server] .env.local ---"
	@if [ -f $(SERVER_DIR)/.env.local ]; then \
		echo "✓ $(SERVER_DIR)/.env.local already exists. Skipping."; \
	else \
		cp $(SERVER_DIR)/.env.local.example $(SERVER_DIR)/.env.local && \
		echo "Created $(SERVER_DIR)/.env.local from example."; \
	fi
	@echo
	@echo "--- [Server] .env.test ---"
	@if [ -f $(SERVER_DIR)/.env.test ]; then \
		echo "✓ $(SERVER_DIR)/.env.test already exists. Skipping."; \
	else \
		cp $(SERVER_DIR)/.env.test.example $(SERVER_DIR)/.env.test && \
		echo "Created $(SERVER_DIR)/.env.test from example."; \
	fi
	@echo
	@echo "--- Installing Dependencies ---"
	@echo "Installing backend dependencies..."
	@cd $(SERVER_DIR) && go mod download
	@echo "Backend dependencies installed."
	@echo
	@echo "Installing frontend dependencies..."
	@cd $(CLIENT_DIR) && bun install
	@echo "Frontend dependencies installed."
	@echo
	@echo ".env Initialization Complete "
	@echo

.PHONY: init-db
init-db: ## Initializes postgres in docker and runs existing migrations
	@echo
	@echo "------------------------------"
	@echo "Initializing Database"
	@echo
	@echo ">> [1/5] Starting Docker environment..."
	@(cd $(SERVER_DIR) && docker-compose up -d)
	@echo "✓ Docker environment started."
	@echo
	@echo ">> [2/5] Creating databases: local and test"
	@echo "--- Local Database ---"
	@$(call use_env,local) \
		&& (cd $(SERVER_DIR) && \
			docker-compose exec db psql -U postgres -c "DROP DATABASE IF EXISTS $$DATABASE_NAME;" && \
			docker-compose exec db psql -U postgres -c "CREATE DATABASE $$DATABASE_NAME;")
	@echo "✓ Local database ready."
	@echo
	@echo "--- Test Database ---"
	@$(call use_env,test) \
		&& (cd $(SERVER_DIR) && \
			docker-compose exec db psql -U postgres -c "DROP DATABASE IF EXISTS $$DATABASE_NAME;" && \
			docker-compose exec db psql -U postgres -c "CREATE DATABASE $$DATABASE_NAME;")
	@echo "✓ Test database ready."
	@echo
	@echo ">> [3/5] Running migrations for local DB..."
	@$(call use_env,local) \
		&& (cd $(SERVER_DIR) && \
			go tool goose -dir db/migrations postgres "$$DATABASE_URL" up)
	@echo "✓ Migrations applied to local DB."
	@echo
	@echo ">> [4/5] Running migrations for test DB..."
	@$(call use_env,test) \
		&& (cd $(SERVER_DIR) && \
			go tool goose -dir db/migrations postgres "$$DATABASE_URL" up)
	@echo "✓ Migrations applied to test DB."
	@echo
	@echo ">> [5/5] Formatting generated backend code..."
	@$(MAKE) format
	@echo
	@echo "Database initialization complete!"
	@echo
