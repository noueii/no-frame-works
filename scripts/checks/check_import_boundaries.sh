#!/usr/bin/env bash
# check_import_boundaries.sh — Checks import rules on changed files only
#
# Rules:
#   1. Modules must not import other modules' internal packages (handler/service/domain/middleware)
#   2. Modules must not import concrete repository packages
#   3. core/ must not import from modules/
#
# Usage: ./check_import_boundaries.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"

HAS_CHANGES=$(echo "$MANIFEST" | jq -r '.has_backend_changes')
if [ "$HAS_CHANGES" != "true" ]; then
  echo '{"check":"import_boundaries","pass":true,"score":0,"max_score":0,"details":"No backend changes","failures":[],"skipped":true}'
  exit 0
fi

# Read go module path
GO_MOD_PATH=$(grep "^module " backend/go.mod 2>/dev/null | awk '{print $2}') || GO_MOD_PATH=""
if [ -z "$GO_MOD_PATH" ]; then
  echo '{"check":"import_boundaries","pass":false,"score":0,"max_score":1,"details":"Could not read go.mod","failures":["go.mod not found"],"skipped":false}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

# --- Check changed module files for cross-module internal imports ---
TOUCHED_MODULES=$(echo "$MANIFEST" | jq -r '.touched_modules[]' 2>/dev/null)

for mod in $TOUCHED_MODULES; do
  MODULE_DIR="backend/internal/modules/$mod"
  [ -d "$MODULE_DIR" ] || continue

  # Get only changed .go files in this module
  CHANGED_IN_MODULE=$(echo "$MANIFEST" | jq -r ".changed_files[] | select(startswith(\"$MODULE_DIR\"))" 2>/dev/null)

  for file in $CHANGED_IN_MODULE; do
    [ -f "$file" ] || continue
    echo "$file" | grep -qE "\.go$" || continue

    # Check 1: No imports of other modules' internal packages
    MAX_SCORE=$((MAX_SCORE + 1))
    BAD_IMPORTS=$(grep -E "\"$GO_MOD_PATH/internal/modules/" "$file" 2>/dev/null \
      | grep -v "\"$GO_MOD_PATH/internal/modules/$mod" \
      | grep -E "(handler|service|domain|middleware)" \
      || true)

    if [ -z "$BAD_IMPORTS" ]; then
      SCORE=$((SCORE + 1))
    else
      FAILURE_LIST+=("$file imports another module's internals")
    fi

    # Check 2: No imports of concrete repository
    MAX_SCORE=$((MAX_SCORE + 1))
    REPO_IMPORTS=$(grep "\"$GO_MOD_PATH/repository/" "$file" 2>/dev/null || true)

    if [ -z "$REPO_IMPORTS" ]; then
      SCORE=$((SCORE + 1))
    else
      FAILURE_LIST+=("$file imports concrete repository")
    fi
  done
done

# --- Check core/ does not import modules/ ---
CORE_CHANGED=$(echo "$MANIFEST" | jq -r '.changed_core')
if [ "$CORE_CHANGED" = "true" ]; then
  CORE_FILES=$(echo "$MANIFEST" | jq -r '.changed_files[] | select(contains("internal/core/"))' 2>/dev/null)

  for file in $CORE_FILES; do
    [ -f "$file" ] || continue
    echo "$file" | grep -qE "\.go$" || continue

    MAX_SCORE=$((MAX_SCORE + 1))
    MODULE_IMPORTS=$(grep "\"$GO_MOD_PATH/internal/modules/" "$file" 2>/dev/null || true)

    if [ -z "$MODULE_IMPORTS" ]; then
      SCORE=$((SCORE + 1))
    else
      FAILURE_LIST+=("$file (core/) imports from modules/")
    fi
  done
fi

if [ "$MAX_SCORE" -eq 0 ]; then
  echo '{"check":"import_boundaries","pass":true,"score":0,"max_score":0,"details":"No relevant files to check","failures":[],"skipped":true}'
  exit 0
fi

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"import_boundaries\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE import checks passed\",\"failures\":$FAILURES,\"skipped\":false}"
