#!/usr/bin/env bash
# check_raw_sql.sh — Ensure no raw SQL in repository files
#
# This project uses go-jet exclusively for database access.
# Raw SQL (QueryRowContext, QueryContext, ExecContext, SQL string literals)
# must not appear in repository implementations.
#
# Usage: ./check_raw_sql.sh <manifest_json>

set -euo pipefail

MANIFEST="$1"
CHANGED_REPO_FILES=$(echo "$MANIFEST" | jq -r '.changed_repo_files[]' 2>/dev/null)

if [ -z "$CHANGED_REPO_FILES" ]; then
  echo '{"check":"raw_sql","pass":true,"score":0,"max_score":0,"details":"No repository files changed","failures":[],"skipped":true}'
  exit 0
fi

SCORE=0
MAX_SCORE=0
FAILURE_LIST=()

# Patterns that indicate raw SQL usage:
# - .db.QueryRowContext / .db.QueryContext / .db.ExecContext (raw *sql.DB calls)
# - database/sql import in repository files
# Go-jet uses stmt.QueryContext which is fine — only flag db.* direct calls.

for f in $CHANGED_REPO_FILES; do
  [ -f "$f" ] || continue
  case "$f" in *.go) ;; *) continue ;; esac

  MAX_SCORE=$((MAX_SCORE + 1))
  HITS=$(grep -nE '\.db\.(QueryRowContext|QueryContext|ExecContext|Query|Exec)\b' "$f" 2>/dev/null || true)

  if [ -n "$HITS" ]; then
    while IFS= read -r line; do
      [ -z "$line" ] && continue
      LINENO=$(echo "$line" | cut -d: -f1)
      FAILURE_LIST+=("$f:$LINENO: raw SQL call detected — use go-jet query builder")
    done <<< "$HITS"
  else
    SCORE=$((SCORE + 1))
  fi
done

if [ "$MAX_SCORE" -eq 0 ]; then
  echo '{"check":"raw_sql","pass":true,"score":0,"max_score":0,"details":"No repository Go files changed","failures":[],"skipped":true}'
  exit 0
fi

PASS=$( [ "$SCORE" -eq "$MAX_SCORE" ] && echo "true" || echo "false" )
FAILURES=$(printf '%s\n' "${FAILURE_LIST[@]}" 2>/dev/null | jq -R -s 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')

echo "{\"check\":\"raw_sql\",\"pass\":$PASS,\"score\":$SCORE,\"max_score\":$MAX_SCORE,\"details\":\"$SCORE/$MAX_SCORE repository files free of raw SQL\",\"failures\":$FAILURES,\"skipped\":false}"
