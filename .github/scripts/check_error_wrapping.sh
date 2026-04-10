#!/usr/bin/env bash
# check_error_wrapping.sh — Errors must be wrapped with %w, never %v or %s
#
# All Errorf calls that wrap an error must use %w to preserve the error chain.
#
# Usage: ./check_error_wrapping.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"
CHANGED_FILES=$(echo "$MANIFEST" | jq -r '.changed_files[]' 2>/dev/null)

GO_FILES=""
for f in $CHANGED_FILES; do
  case "$f" in
    *.go)
      case "$f" in
        */generated/*) ;;
        *) GO_FILES="$GO_FILES $f" ;;
      esac
      ;;
  esac
done

if [ -z "$GO_FILES" ]; then
  echo '{"check":"error_wrapping","pass":true,"score":0,"max_score":0,"details":"No Go files changed","failures":[],"skipped":true}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for f in $GO_FILES; do
  [ -f "$f" ] || continue

  MAX_SCORE=$((MAX_SCORE + 1))
  # Match Errorf calls that wrap an error with %v or %s instead of %w.
  # Pattern: has multiple format verbs, and the last one is %v or %s (not %w).
  # Excludes single-verb calls like Errorf("%s", msg) which create, not wrap.
  HITS=$(grep -nE 'Errorf\(".*%.*%[vs]"' "$f" 2>/dev/null | grep -v '%w' || true)
  if [ -n "$HITS" ]; then
    while IFS= read -r line; do
      LINENO=$(echo "$line" | cut -d: -f1)
      FAILURE_LIST+=("$f:$LINENO: Errorf uses %v or %s — use %w to preserve error chain")
    done <<< "$HITS"
  else
    SCORE=$((SCORE + 1))
  fi
done

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"error_wrapping\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE files use correct error wrapping\",\"failures\":$FAILURES,\"skipped\":false}"
