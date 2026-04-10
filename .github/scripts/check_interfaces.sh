#!/usr/bin/env bash
# check_interfaces.sh — Checks that request structs in changed files implement Validate() and CheckPermission()
#
# Only inspects modules that were touched in this PR.
#
# Usage: ./check_interfaces.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"
TOUCHED_MODULES=$(echo "$MANIFEST" | jq -r '.touched_modules[]' 2>/dev/null)

if [ -z "$TOUCHED_MODULES" ]; then
  echo '{"check":"interfaces","pass":true,"score":0,"max_score":0,"details":"No modules touched","failures":[],"skipped":true}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for mod in $TOUCHED_MODULES; do
  MODULE_DIR="backend/internal/modules/$mod"
  [ -d "$MODULE_DIR" ] || continue

  # Find request structs defined anywhere in the module
  REQUEST_STRUCTS=$(grep -rh "type [A-Z][A-Za-z]*Request struct" "$MODULE_DIR" 2>/dev/null \
    | sed -E 's/type ([A-Z][A-Za-z]*Request) struct.*/\1/' \
    || true)

  for struct in $REQUEST_STRUCTS; do
    [ -z "$struct" ] && continue

    # Check Validate
    MAX_SCORE=$((MAX_SCORE + 1))
    if grep -rl "func (.*$struct) Validate()" "$MODULE_DIR" > /dev/null 2>&1; then
      SCORE=$((SCORE + 1))
    else
      FAILURE_LIST+=("$mod: $struct missing Validate() error")
    fi

    # Check CheckPermission
    MAX_SCORE=$((MAX_SCORE + 1))
    if grep -rl "func (.*$struct) CheckPermission(" "$MODULE_DIR" > /dev/null 2>&1; then
      SCORE=$((SCORE + 1))
    else
      FAILURE_LIST+=("$mod: $struct missing CheckPermission()")
    fi
  done
done

if [ "$MAX_SCORE" -eq 0 ]; then
  echo '{"check":"interfaces","pass":true,"score":0,"max_score":0,"details":"No request structs found in touched modules","failures":[],"skipped":true}'
  exit 0
fi

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"interfaces\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE interface methods found\",\"failures\":$FAILURES,\"skipped\":false}"
