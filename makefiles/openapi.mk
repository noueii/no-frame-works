##@ OpenAPI - Helpers for openapi related commands

OPENAPI_SRC := ./openapi/internal/openapi.yaml
OPENAPI_BUNDLED := ./openapi/internal/bundled.yaml

.PHONY: gen-openapi
gen-openapi: ## Generate OpenAPI frontend and backend code
	@echo "Bundling OpenAPI spec..."
	@npx --yes @redocly/cli bundle $(OPENAPI_SRC) -o $(OPENAPI_BUNDLED)
	@echo "Generating backend code based on OpenAPI schema..."
	@(cd $(SERVER_DIR) && go tool oapi-codegen -config ./generated/oapi/codegen.yaml ../$(OPENAPI_BUNDLED))
	@echo "Generating frontend code based on OpenAPI schema..."
	@(cd $(CLIENT_DIR) && bun run openapi)
	@echo "OpenAPI generation complete."
