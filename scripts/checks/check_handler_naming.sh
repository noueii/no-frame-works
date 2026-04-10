#!/usr/bin/env bash
# check_handler_naming.sh — Only checks handler files that were changed/added
#
# Usage: ./check_handler_naming.sh <manifest_json>
# Reads the diff manifest and only checks changed handler files.

set -euo pipefail

MANIFEST="$1"
CHANGED_HANDLERS=$(echo "$MANIFEST" | jq -r '.changed_handlers[]' 2>/dev/null)

if [ -z "$CHANGED_HANDLERS" ]; then
  echo '{"check":"handler_naming","pass":true,"score":0,"max_score":0,"details":"No handler files changed","failures":[],"skipped":true}'
  exit 0
fi

VALID_PATTERN="^(get|post|put|patch|delete)_[a-z][a-z0-9_]*\.go$"
EXCEPTIONS="^(handler|routes)\.go$"

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for file in $CHANGED_HANDLERS; do
  basename=$(basename "$file")

  # Skip exceptions
  echo "$basename" | grep -qE "$EXCEPTIONS" && continue
  echo "$basename" | grep -qE "_test\.go$" && continue

  MAX_SCORE=$((MAX_SCORE + 1))

  if echo "$basename" | grep -qE "$VALID_PATTERN"; then
    SCORE=$((SCORE + 1))
  else
    FAILURE_LIST+=("$file → should match <verb>_<resource>.go")
  fi
done

if [ "$MAX_SCORE" -eq 0 ]; then
  echo '{"check":"handler_naming","pass":true,"score":0,"max_score":0,"details":"Only handler.go/routes.go changed","failures":[],"skipped":true}'
  exit 0
fi

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"handler_naming\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE changed handler files correctly named\",\"failures\":$FAILURES,\"skipped\":false}"
