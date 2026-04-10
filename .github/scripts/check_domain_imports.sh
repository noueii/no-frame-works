#!/usr/bin/env bash
# check_domain_imports.sh — No infrastructure imports in domain files
#
# Domain files must not import database, HTTP, or external SDK packages.
#
# Usage: ./check_domain_imports.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"
DOMAIN_FILES=$(echo "$MANIFEST" | jq -r '.changed_domain[]' 2>/dev/null)

if [ -z "$DOMAIN_FILES" ]; then
  echo '{"check":"domain_imports","pass":true,"score":0,"max_score":0,"details":"No domain files changed","failures":[],"skipped":true}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

BANNED_IMPORTS='database/sql\|net/http\|github\.com/go-jet/\|github\.com/ory/\|github\.com/aws/\|github\.com/redis/\|github\.com/lib/pq'

for f in $DOMAIN_FILES; do
  [ -f "$f" ] || continue
  case "$f" in *.go) ;; *) continue ;; esac

  MAX_SCORE=$((MAX_SCORE + 1))
  HITS=$(grep -n "$BANNED_IMPORTS" "$f" 2>/dev/null || true)
  if [ -n "$HITS" ]; then
    while IFS= read -r line; do
      LINENO=$(echo "$line" | cut -d: -f1)
      FAILURE_LIST+=("$f:$LINENO: infrastructure import in domain — domain must be pure business logic")
    done <<< "$HITS"
  else
    SCORE=$((SCORE + 1))
  fi
done

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"domain_imports\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE domain files free of infra imports\",\"failures\":$FAILURES,\"skipped\":false}"
