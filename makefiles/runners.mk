##@ Runners - Helpers to run server components

.PHONY: run-docker
run-docker: ## Runs the docker environment
	@cd $(SERVER_DIR) && docker-compose up -d

.PHONY: run-server
run-server: ## Start server and worker (Ctrl+C to stop)
	@echo "Starting server and worker, press Ctrl+C to stop..."
	@$(call use_env,local) && \
	cd $(SERVER_DIR) && \
	trap 'kill 0' EXIT; \
	go run ./cmd/webserver & go run ./cmd/worker & \
	wait

.PHONY: run-webserver
run-webserver: ## Start webserver (Ctrl+C to stop)
	@echo "Starting webserver, press Ctrl+C to stop..."
	@$(call use_env,local) \
		&& cd $(SERVER_DIR) && \
		go run ./cmd/webserver

.PHONY: run-worker
run-worker: ## Start worker (Ctrl+C to stop)
	@echo "Starting worker, press Ctrl+C to stop..."
	@$(call use_env,local) \
		&& cd $(SERVER_DIR) && \
		go run ./cmd/worker

.PHONY: run-client
run-client: ## Start frontend dev server (Ctrl+C to stop)
	@echo "Starting frontend dev server on http://localhost:3000..."
	@cd $(CLIENT_DIR) && bun run dev
