default: help

include makefiles/common.mk
include makefiles/init.mk
include makefiles/openapi.mk
include makefiles/migrations.mk
include makefiles/runners.mk
include makefiles/utility.mk

.PHONY: help
help: ## Show a list of commands
	@clear
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[0;33m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""
