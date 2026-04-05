SHELL := bash

# Paths
SERVER_DIR := ./backend
CLIENT_DIR := ./frontend
BIN_DIR := ./bin

# Environment loading function to source appropriate .env file
define use_env
	@echo "Using .env.$(1) as ENV"
	@set -o allexport; source $(SERVER_DIR)/.env.$(1); set +o allexport
endef
