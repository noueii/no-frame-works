#!/usr/bin/env bash
# analyze-diff.sh — Determine what changed in the PR and what needs checking
#
# Outputs a JSON manifest describing what changed, so individual checks
# can decide whether they're relevant.
#
# Usage: ./scripts/checks/analyze-diff.sh <base_ref>
# Example: ./scripts/checks/analyze-diff.sh origin/main

set -euo pipefail

BASE_REF="${1:-origin/main}"

# Get all changed files relative to base
CHANGED_FILES=$(git diff --name-only "$BASE_REF" -- 'backend/' 2>/dev/null || echo "")

if [ -z "$CHANGED_FILES" ]; then
  echo '{"has_backend_changes":false,"changed_files":[],"touched_modules":[],"touched_repos":[],"new_modules":[],"changed_handlers":[],"changed_services":[],"changed_repos":[],"changed_providers":[],"changed_core":false}'
  exit 0
fi

# Detect touched modules (internal/modules/<name>/...)
TOUCHED_MODULES=$(echo "$CHANGED_FILES" \
  | grep -oE "backend/internal/modules/[^/]+" \
  | sed 's|backend/internal/modules/||' \
  | sort -u || true)

# Detect new modules (module dir exists now but not in base)
NEW_MODULES=""
for mod in $TOUCHED_MODULES; do
  if ! git show "$BASE_REF:backend/internal/modules/$mod/api.go" > /dev/null 2>&1; then
    NEW_MODULES="$NEW_MODULES $mod"
  fi
done

# Detect touched repository implementations
TOUCHED_REPOS=$(echo "$CHANGED_FILES" \
  | grep -oE "backend/repository/[^/]+" \
  | sed 's|backend/repository/||' \
  | sort -u || true)

# Categorize changed files
CHANGED_HANDLERS=$(echo "$CHANGED_FILES" | grep "handler/http/" || true)
CHANGED_SERVICES=$(echo "$CHANGED_FILES" | grep "/service/" || true)
CHANGED_REPO_FILES=$(echo "$CHANGED_FILES" | grep "^backend/repository/" || true)
CHANGED_PROVIDERS=$(echo "$CHANGED_FILES" | grep "internal/provider/" || true)
CHANGED_CORE=$(echo "$CHANGED_FILES" | grep -q "internal/core/" && echo "true" || echo "false")
CHANGED_DOMAIN=$(echo "$CHANGED_FILES" | grep "/domain/" || true)
CHANGED_MIDDLEWARE=$(echo "$CHANGED_FILES" | grep "/middleware/" || true)

# Build JSON
to_json_array() {
  if [ -z "$1" ]; then
    echo "[]"
  else
    echo "$1" | jq -R -s 'split("\n") | map(select(length > 0))'
  fi
}

jq -n \
  --argjson has_backend true \
  --argjson changed_files "$(to_json_array "$CHANGED_FILES")" \
  --argjson touched_modules "$(to_json_array "$TOUCHED_MODULES")" \
  --argjson touched_repos "$(to_json_array "$TOUCHED_REPOS")" \
  --argjson new_modules "$(to_json_array "$(echo "$NEW_MODULES" | xargs)")" \
  --argjson changed_handlers "$(to_json_array "$CHANGED_HANDLERS")" \
  --argjson changed_services "$(to_json_array "$CHANGED_SERVICES")" \
  --argjson changed_repo_files "$(to_json_array "$CHANGED_REPO_FILES")" \
  --argjson changed_providers "$(to_json_array "$CHANGED_PROVIDERS")" \
  --argjson changed_core "$CHANGED_CORE" \
  --argjson changed_domain "$(to_json_array "$CHANGED_DOMAIN")" \
  --argjson changed_middleware "$(to_json_array "$CHANGED_MIDDLEWARE")" \
  '{
    has_backend_changes: $has_backend,
    changed_files: $changed_files,
    touched_modules: $touched_modules,
    touched_repos: $touched_repos,
    new_modules: $new_modules,
    changed_handlers: $changed_handlers,
    changed_services: $changed_services,
    changed_repo_files: $changed_repo_files,
    changed_providers: $changed_providers,
    changed_core: $changed_core,
    changed_domain: $changed_domain,
    changed_middleware: $changed_middleware
  }'
