#!/usr/bin/env bash
# format-terminal.sh — Pretty-print convention check JSON results for terminal
#
# Reads JSON from stdin (output of run-checks.sh), prints a table with colors.
# Exits with code 1 if any check failed.

set -euo pipefail

INPUT=$(cat)

PASS=$(echo "$INPUT" | jq -r '.pass')
SUMMARY=$(echo "$INPUT" | jq -r '.summary')
CHECKS=$(echo "$INPUT" | jq -c '.checks[]')

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
DIM='\033[2m'
RESET='\033[0m'

echo ""
if [ "$PASS" = "true" ]; then
  echo -e "${GREEN}✅ Convention checks passed${RESET} ($SUMMARY)"
else
  echo -e "${RED}❌ Convention checks failed${RESET} ($SUMMARY)"
fi
echo ""

printf "  %-25s %-10s %s\n" "CHECK" "RESULT" "SCORE"
printf "  %-25s %-10s %s\n" "─────" "──────" "─────"

echo "$CHECKS" | while read -r check; do
  NAME=$(echo "$check" | jq -r '.check')
  CHECK_PASS=$(echo "$check" | jq -r '.pass')
  SKIPPED=$(echo "$check" | jq -r '.skipped // false')
  SCORE=$(echo "$check" | jq -r '.score')
  MAX=$(echo "$check" | jq -r '.max_score')

  if [ "$SKIPPED" = "true" ]; then
    printf "  ${DIM}%-25s ⏭  skipped  —${RESET}\n" "$NAME"
  elif [ "$CHECK_PASS" = "true" ]; then
    printf "  %-25s ${GREEN}✅ pass${RESET}     %s/%s\n" "$NAME" "$SCORE" "$MAX"
  else
    printf "  %-25s ${RED}❌ fail${RESET}     %s/%s\n" "$NAME" "$SCORE" "$MAX"
  fi
done

# Print failures
FAILURES=$(echo "$INPUT" | jq -c '.checks[] | select(.pass == false and (.skipped // false) == false)')
if [ -n "$FAILURES" ]; then
  echo ""
  echo "$FAILURES" | while read -r check; do
    NAME=$(echo "$check" | jq -r '.check')
    echo -e "  ${RED}$NAME:${RESET}"
    echo "$check" | jq -r '.failures[]' | while read -r msg; do
      echo "    - $msg"
    done
  done
fi

echo ""

if [ "$PASS" != "true" ]; then
  exit 1
fi
