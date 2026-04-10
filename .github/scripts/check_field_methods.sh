#!/usr/bin/env bash
# check_field_methods.sh — No field-specific repository methods
#
# Repository interfaces must use full domain models (model-in, model-out).
# Methods like UpdateUsername(id, username) are violations — should be Update(ctx, domain.User).
#
# Usage: ./check_field_methods.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"

# Find repository interface files (internal/modules/*/repository.go)
REPO_IFACE_FILES=$(echo "$MANIFEST" | jq -r '.changed_files[]' 2>/dev/null \
  | grep -E 'internal/modules/[^/]+/repository\.go$' \
  || true)

if [ -z "$REPO_IFACE_FILES" ]; then
  echo '{"check":"field_methods","pass":true,"score":0,"max_score":0,"details":"No repository interface files changed","failures":[],"skipped":true}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for f in $REPO_IFACE_FILES; do
  [ -f "$f" ] || continue

  MAX_SCORE=$((MAX_SCORE + 1))
  # Match field-specific method patterns: Update<Field>, Set<Field>, Change<Field>
  # But allow plain Update, Create, Delete, Find
  HITS=$(grep -nE '(Update[A-Z][a-zA-Z]+|Set[A-Z][a-zA-Z]+|Change[A-Z][a-zA-Z]+)\(' "$f" 2>/dev/null || true)
  if [ -n "$HITS" ]; then
    while IFS= read -r line; do
      LINENO=$(echo "$line" | cut -d: -f1)
      FAILURE_LIST+=("$f:$LINENO: field-specific method — use Update(ctx, domain.Model) instead")
    done <<< "$HITS"
  else
    SCORE=$((SCORE + 1))
  fi
done

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"field_methods\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE repository interfaces use model-in/model-out\",\"failures\":$FAILURES,\"skipped\":false}"
