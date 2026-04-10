#!/usr/bin/env bash
# check_error_imports.sh — Ensure no fmt.Errorf or stdlib "errors" in changed Go files
#
# This project uses github.com/go-errors/errors exclusively.
# Generated files are excluded.
#
# Usage: ./check_error_imports.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"
CHANGED_FILES=$(echo "$MANIFEST" | jq -r '.changed_files[]' 2>/dev/null)

# Filter to non-generated .go files
GO_FILES=""
for f in $CHANGED_FILES; do
  case "$f" in
    *.go)
      case "$f" in
        */generated/*) ;; # skip generated
        *) GO_FILES="$GO_FILES $f" ;;
      esac
      ;;
  esac
done

if [ -z "$GO_FILES" ]; then
  echo '{"check":"error_imports","pass":true,"score":0,"max_score":0,"details":"No Go files changed","failures":[],"skipped":true}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

for f in $GO_FILES; do
  [ -f "$f" ] || continue

  # Check for fmt.Errorf
  MAX_SCORE=$((MAX_SCORE + 1))
  FMT_HITS=$(grep -n 'fmt\.Errorf' "$f" 2>/dev/null || true)
  if [ -n "$FMT_HITS" ]; then
    while IFS= read -r line; do
      LINENO=$(echo "$line" | cut -d: -f1)
      FAILURE_LIST+=("$f:$LINENO: uses fmt.Errorf — use errors.Errorf from github.com/go-errors/errors")
    done <<< "$FMT_HITS"
  else
    SCORE=$((SCORE + 1))
  fi

  # Check for stdlib "errors" import
  MAX_SCORE=$((MAX_SCORE + 1))
  # Match bare "errors" import but not "github.com/go-errors/errors"
  STDLIB_HITS=$(grep -n '"errors"' "$f" 2>/dev/null | grep -v 'go-errors' || true)
  if [ -n "$STDLIB_HITS" ]; then
    while IFS= read -r line; do
      LINENO=$(echo "$line" | cut -d: -f1)
      FAILURE_LIST+=("$f:$LINENO: imports stdlib \"errors\" — use \"github.com/go-errors/errors\"")
    done <<< "$STDLIB_HITS"
  else
    SCORE=$((SCORE + 1))
  fi
done

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"error_imports\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE error import checks passed\",\"failures\":$FAILURES,\"skipped\":false}"
