#!/usr/bin/env bash
# check_validate_first.sh — Service functions must call Validate() and CheckPermission()
#
# In service subfolder files (service/<name>/*.go), the exported function must:
# 1. NOT be named Execute (use the operation name instead)
# 2. Call Validate() in the first few lines
# 3. Call CheckPermission() somewhere in the function
#
# Usage: ./check_validate_first.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"

# Get service subfolder files (not service/service.go)
SVC_FILES=$(echo "$MANIFEST" | jq -r '.changed_services[]' 2>/dev/null \
  | grep -E 'service/[^/]+/[^/]+\.go$' \
  | grep -v 'service/service\.go' \
  || true)

if [ -z "$SVC_FILES" ]; then
  echo '{"check":"validate_first","pass":true,"score":0,"max_score":0,"details":"No service subfolder files changed","failures":[],"skipped":true}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for f in $SVC_FILES; do
  [ -f "$f" ] || continue

  # Check for Execute (should not exist — use named functions)
  EXEC_LINE=$(grep -n '^func Execute(' "$f" 2>/dev/null | head -1 | cut -d: -f1 || true)
  if [ -n "$EXEC_LINE" ]; then
    MAX_SCORE=$((MAX_SCORE + 1))
    FAILURE_LIST+=("$f:$EXEC_LINE: function named Execute — use the operation name (e.g. CreatePost, EditUsername)")
    continue
  fi

  # Find the first exported function (capitalized, not a method)
  FUNC_LINE=$(grep -n '^func [A-Z]' "$f" 2>/dev/null | head -1 | cut -d: -f1 || true)
  [ -z "$FUNC_LINE" ] && continue

  FUNC_NAME=$(grep -n '^func [A-Z]' "$f" 2>/dev/null | head -1 | sed 's/.*func \([A-Za-z]*\)(.*/\1/')

  # Check 1: Validate() in first 5 lines of function body
  MAX_SCORE=$((MAX_SCORE + 1))
  HAS_VALIDATE=$(sed -n "$((FUNC_LINE + 1)),$((FUNC_LINE + 5))p" "$f" | grep -c 'Validate()' || true)
  if [ "$HAS_VALIDATE" -eq 0 ]; then
    FAILURE_LIST+=("$f:$FUNC_LINE: $FUNC_NAME() does not call req.Validate() first")
  else
    SCORE=$((SCORE + 1))
  fi

  # Check 2: CheckPermission() somewhere in the function
  MAX_SCORE=$((MAX_SCORE + 1))
  HAS_PERM=$(grep -c 'CheckPermission(' "$f" || true)
  if [ "$HAS_PERM" -eq 0 ]; then
    FAILURE_LIST+=("$f:$FUNC_LINE: $FUNC_NAME() does not call req.CheckPermission()")
  else
    SCORE=$((SCORE + 1))
  fi
done

if [ "$MAX_SCORE" -eq 0 ]; then
  echo '{"check":"validate_first","pass":true,"score":0,"max_score":0,"details":"No service functions in changed files","failures":[],"skipped":true}'
  exit 0
fi

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"validate_first\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE service function checks passed\",\"failures\":$FAILURES,\"skipped\":false}"
