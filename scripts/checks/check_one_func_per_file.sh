#!/usr/bin/env bash
# check_one_func_per_file.sh — Checks changed files in service/repo subfolders
#
# Only checks files that were added or modified in this PR.
#
# Usage: ./check_one_func_per_file.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"

CHANGED_SERVICES=$(echo "$MANIFEST" | jq -r '.changed_services[]' 2>/dev/null)
CHANGED_REPOS=$(echo "$MANIFEST" | jq -r '.changed_repo_files[]' 2>/dev/null)

ALL_CHANGED="$CHANGED_SERVICES
$CHANGED_REPOS"
ALL_CHANGED=$(echo "$ALL_CHANGED" | sed '/^$/d')

if [ -z "$ALL_CHANGED" ]; then
  echo '{"check":"one_func_per_file","pass":true,"score":0,"max_score":0,"details":"No service/repository files changed","failures":[],"skipped":true}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for file in $ALL_CHANGED; do
  [ -f "$file" ] || continue
  echo "$file" | grep -qE "\.go$" || continue
  echo "$file" | grep -qE "_test\.go$" && continue

  basename=$(basename "$file")

  # Skip root-level files (service.go, postgres.go) — the rule is about subfolders
  # Check if the file is in a subfolder (at least 2 levels deep under service/ or repository/<mod>/)
  # For services: backend/internal/modules/<mod>/service/<subfolder>/file.go
  # For repos: backend/repository/<mod>/<subfolder>/file.go

  if echo "$file" | grep -qE "/service/[^/]+/[^/]+\.go$"; then
    # It's in a service subfolder
    :
  elif echo "$file" | grep -qE "/repository/[^/]+/[^/]+/[^/]+\.go$"; then
    # It's in a repository subfolder
    :
  else
    # Root level file, skip
    continue
  fi

  MAX_SCORE=$((MAX_SCORE + 1))
  FUNC_COUNT=$(grep -cE "^func " "$file" 2>/dev/null || echo "0")

  if [ "$FUNC_COUNT" -le 1 ]; then
    SCORE=$((SCORE + 1))
  else
    FAILURE_LIST+=("$file has $FUNC_COUNT functions (expected 1)")
  fi
done

if [ "$MAX_SCORE" -eq 0 ]; then
  echo '{"check":"one_func_per_file","pass":true,"score":0,"max_score":0,"details":"No subfolder files changed","failures":[],"skipped":true}'
  exit 0
fi

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"one_func_per_file\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE changed subfolder files have single function\",\"failures\":$FAILURES,\"skipped\":false}"
