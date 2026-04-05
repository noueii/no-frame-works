##@ Migrations - Helpers to run database migration commands

.PHONY: migration-up
migration-up: ## Run database migrations
	@echo "Running migrations..."
	@$(MAKE) migration-up-local
	@$(MAKE) migration-up-test
	@echo "Finished running migrations."

.PHONY: migration-up-local
migration-up-local: ## Migrate local database
	@echo "Migrating local DB..."
	@$(call use_env,local) \
		&& (cd $(SERVER_DIR) && \
		 go tool goose -dir db/migrations postgres "$$DATABASE_URL" up && \
		echo "Generating DB models with go-jet..." && \
		 go tool jet -dsn="$$DATABASE_URL" -schema=public -path=./db)
	@$(MAKE) format
	@echo "Local DB migration complete."

.PHONY: migration-up-test
migration-up-test: ## Migrate test database
	@echo "Migrating test DB..."
	@$(call use_env,test) \
		&& (cd $(SERVER_DIR) && \
		 go tool goose -dir db/migrations postgres "$$DATABASE_URL" up)
	@echo "Test DB migration complete."

.PHONY: migration-down
migration-down: ## Roll back database migrations
	@echo "Rolling back database migrations..."
	@$(MAKE) migration-down-local
	@$(MAKE) migration-down-test
	@echo "Finished rolling back migrations."

.PHONY: migration-down-local
migration-down-local: ## Rollback local database migrations
	@echo "Rolling back local DB..."
	@$(call use_env,local) \
		&& (cd $(SERVER_DIR) && \
		 go tool goose -dir db/migrations postgres "$$DATABASE_URL" down) || true
	@echo "Local DB rollback complete."

.PHONY: migration-down-test
migration-down-test: ## Rollback test database migrations
	@echo "Rolling back test DB..."
	@$(call use_env,test) \
		&& (cd $(SERVER_DIR) && \
		 go tool goose -dir db/migrations postgres "$$DATABASE_URL" down) || true
	@echo "Test DB rollback complete."

.PHONY: migration-reset
migration-reset: ## Reset database migrations
	@echo "Resetting local DB..."
	@$(call use_env,local) \
		&& (cd $(SERVER_DIR) && \
		 go tool goose -dir db/migrations postgres "$$DATABASE_URL" reset) || true
	@echo "Resetting test DB..."
	@$(call use_env,test) \
		&& (cd $(SERVER_DIR) && \
		 go tool goose -dir db/migrations postgres "$$DATABASE_URL" reset) || true
	@echo "DB reset complete."

.PHONY: migration-status
migration-status: ## Show migration status
	@echo "Checking migration status..."
	@$(call use_env,local) \
		&& (cd $(SERVER_DIR) && go tool goose -dir db/migrations postgres "$$DATABASE_URL" status)
	@echo "Migration status check complete."
