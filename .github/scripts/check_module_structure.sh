#!/usr/bin/env bash
# check_module_structure.sh — Checks structure only for NEW modules in the PR
#
# If the PR adds a brand new module, it must have the full folder structure.
# Existing modules are not checked (they were validated when first created).
#
# Usage: ./check_module_structure.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"
NEW_MODULES=$(echo "$MANIFEST" | jq -r '.new_modules[]' 2>/dev/null)

if [ -z "$NEW_MODULES" ]; then
  echo '{"check":"module_structure","pass":true,"score":0,"max_score":0,"details":"No new modules in this PR","failures":[],"skipped":true}'
  exit 0
fi

REQUIRED_DIRS=("service" "domain")
REQUIRED_FILES=("api.go" "repository.go" "errors.go" "service/service.go")

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for mod in $NEW_MODULES; do
  MODULE_DIR="backend/internal/modules/$mod"

  for dir in "${REQUIRED_DIRS[@]}"; do
    MAX_SCORE=$((MAX_SCORE + 1))
    if [ -d "$MODULE_DIR/$dir" ]; then
      SCORE=$((SCORE + 1))
    else
      FAILURE_LIST+=("$mod: missing directory $dir")
    fi
  done

  for file in "${REQUIRED_FILES[@]}"; do
    MAX_SCORE=$((MAX_SCORE + 1))
    if [ -f "$MODULE_DIR/$file" ]; then
      SCORE=$((SCORE + 1))
    else
      FAILURE_LIST+=("$mod: missing file $file")
    fi
  done

  # Check repository dir exists
  MAX_SCORE=$((MAX_SCORE + 1))
  if [ -d "backend/repository/$mod" ]; then
    SCORE=$((SCORE + 1))
  else
    FAILURE_LIST+=("$mod: missing repository/$mod directory")
  fi
done

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"module_structure\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE structural checks passed for new modules\",\"failures\":$FAILURES,\"skipped\":false}"
