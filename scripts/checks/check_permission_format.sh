#!/usr/bin/env bash
# check_permission_format.sh — Checks permission constants in touched modules
#
# Usage: ./check_permission_format.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"
TOUCHED_MODULES=$(echo "$MANIFEST" | jq -r '.touched_modules[]' 2>/dev/null)

if [ -z "$TOUCHED_MODULES" ]; then
  echo '{"check":"permission_format","pass":true,"score":0,"max_score":0,"details":"No modules touched","failures":[],"skipped":true}'
  exit 0
fi

VALID_PATTERN="^[a-z]+:[a-z_]+:[a-z_]+$"

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for mod in $TOUCHED_MODULES; do
  PERM_FILE="backend/internal/modules/$mod/permissions.go"
  [ -f "$PERM_FILE" ] || continue

  PERM_VALUES=$(grep -oE '"[a-zA-Z0-9:_]+"' "$PERM_FILE" 2>/dev/null | tr -d '"' || true)

  for perm in $PERM_VALUES; do
    MAX_SCORE=$((MAX_SCORE + 1))
    if echo "$perm" | grep -qE "$VALID_PATTERN"; then
      SCORE=$((SCORE + 1))
    else
      FAILURE_LIST+=("$mod: \"$perm\" doesn't match module:resource:action")
    fi
  done
done

if [ "$MAX_SCORE" -eq 0 ]; then
  echo '{"check":"permission_format","pass":true,"score":0,"max_score":0,"details":"No permission constants found","failures":[],"skipped":true}'
  exit 0
fi

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"permission_format\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE permissions correctly formatted\",\"failures\":$FAILURES,\"skipped\":false}"
