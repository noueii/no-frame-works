##@ OpenAPI - Helpers for openapi related commands

OPENAPI_SPEC := ./openapi.yaml

.PHONY: gen-openapi
gen-openapi: ## Generate OpenAPI frontend and backend code
	@echo "Generating backend code based on OpenAPI schema..."
	@(cd $(SERVER_DIR) && go tool oapi-codegen -config ./generated/oapi/codegen.yaml ../$(OPENAPI_SPEC))
	@echo "OpenAPI generation complete."
