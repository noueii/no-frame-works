#!/usr/bin/env bash
# format-pr-comment.sh — Convert check results JSON to a markdown PR comment
#
# Usage: echo "$RESULTS_JSON" | ./format-pr-comment.sh
# Or:    ./format-pr-comment.sh < results.json

set -euo pipefail

RESULTS=$(cat)

TOTAL=$(echo "$RESULTS" | jq '.total_score')
MAX=$(echo "$RESULTS" | jq '.max_score')
PCT=$(echo "$RESULTS" | jq '.percentage')
PASS=$(echo "$RESULTS" | jq '.pass')

if [ "$PASS" = "true" ]; then
  HEADER="## ✅ Convention checks passed ($TOTAL/$MAX — ${PCT}%)"
else
  HEADER="## ❌ Convention checks failed ($TOTAL/$MAX — ${PCT}%)"
fi

echo "$HEADER"
echo ""

# Summary table
echo "| Check | Result | Score | Details |"
echo "|-------|--------|-------|---------|"

echo "$RESULTS" | jq -r '.checks[] |
  if .skipped == true then
    "| \(.check) | ⏭️ skipped | — | \(.details) |"
  elif .pass == true then
    "| \(.check) | ✅ pass | \(.score)/\(.max_score) | \(.details) |"
  else
    "| \(.check) | ❌ fail | \(.score)/\(.max_score) | \(.details) |"
  end'

# Show failures if any
FAILURES=$(echo "$RESULTS" | jq -r '.checks[] | select(.pass == false and .skipped != true) | "### ❌ \(.check)\n\(.failures | map("- `\(.)`") | join("\n"))\n"')

if [ -n "$FAILURES" ]; then
  echo ""
  echo "<details>"
  echo "<summary>Failure details</summary>"
  echo ""
  echo "$FAILURES"
  echo "</details>"
fi
