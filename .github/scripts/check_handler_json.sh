#!/usr/bin/env bash
# check_handler_json.sh — No manual JSON decode/encode in handler files
#
# Handlers must use oapi-codegen typed request/response objects.
# Manual json.NewDecoder, json.NewEncoder, json.Unmarshal, http.Error are violations.
#
# Usage: ./check_handler_json.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"

HANDLER_FILES=$(echo "$MANIFEST" | jq -r '.changed_files[]' 2>/dev/null \
  | grep -E 'handler/' | grep '\.go$' \
  | grep -v generated \
  | grep -v 'dto_' \
  || true)

if [ -z "$HANDLER_FILES" ]; then
  echo '{"check":"handler_json","pass":true,"score":0,"max_score":0,"details":"No handler files changed","failures":[],"skipped":true}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for f in $HANDLER_FILES; do
  [ -f "$f" ] || continue

  MAX_SCORE=$((MAX_SCORE + 1))
  HITS=$(grep -nE 'json\.NewDecoder|json\.NewEncoder|json\.Unmarshal|json\.Marshal|http\.Error\(' "$f" 2>/dev/null || true)
  if [ -n "$HITS" ]; then
    while IFS= read -r line; do
      LINENO=$(echo "$line" | cut -d: -f1)
      FAILURE_LIST+=("$f:$LINENO: manual JSON or http.Error — use oapi-codegen typed request/response objects")
    done <<< "$HITS"
  else
    SCORE=$((SCORE + 1))
  fi
done

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"handler_json\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE handler files free of manual JSON\",\"failures\":$FAILURES,\"skipped\":false}"
