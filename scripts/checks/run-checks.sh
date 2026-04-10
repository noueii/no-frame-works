#!/usr/bin/env bash
# run-checks.sh — Orchestrate all convention checks for a PR
#
# Usage: ./scripts/checks/run-checks.sh [base_ref]
# Default base_ref: origin/main
#
# Outputs JSON results to stdout, sets exit code 1 if any non-skipped check fails.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BASE_REF="${1:-origin/main}"

# Step 1: Analyze the diff
echo "::group::Analyzing diff against $BASE_REF" >&2
MANIFEST=$("$SCRIPT_DIR/analyze-diff.sh" "$BASE_REF")
echo "$MANIFEST" | jq . >&2
echo "::endgroup::" >&2

HAS_CHANGES=$(echo "$MANIFEST" | jq -r '.has_backend_changes')
if [ "$HAS_CHANGES" != "true" ]; then
  echo '{"checks":[],"total_score":0,"max_score":0,"percentage":100,"pass":true,"summary":"No backend changes detected"}' >&2
  echo '{"checks":[],"total_score":0,"max_score":0,"percentage":100,"pass":true,"summary":"No backend changes detected"}'
  exit 0
fi

# Step 2: Run all checks
RESULTS="[]"
TOTAL_SCORE=0
TOTAL_MAX=0
ANY_FAILED=false

for check_script in "$SCRIPT_DIR"/check_*.sh; do
  [ -x "$check_script" ] || chmod +x "$check_script"
  CHECK_NAME=$(basename "$check_script" .sh | sed 's/^check_//')

  echo "::group::Running $CHECK_NAME" >&2

  RESULT=$("$check_script" "$MANIFEST" 2>/dev/null || echo "{\"check\":\"$CHECK_NAME\",\"pass\":false,\"score\":0,\"max_score\":1,\"details\":\"Script error\",\"failures\":[\"Check script crashed\"],\"skipped\":false}")

  echo "$RESULT" | jq . >&2
  echo "::endgroup::" >&2

  # Only count non-skipped checks
  SKIPPED=$(echo "$RESULT" | jq '.skipped // false')
  if [ "$SKIPPED" != "true" ]; then
    SCORE=$(echo "$RESULT" | jq '.score // 0')
    MAX=$(echo "$RESULT" | jq '.max_score // 0')
    TOTAL_SCORE=$((TOTAL_SCORE + SCORE))
    TOTAL_MAX=$((TOTAL_MAX + MAX))

    PASS=$(echo "$RESULT" | jq '.pass')
    if [ "$PASS" != "true" ]; then
      ANY_FAILED=true
    fi
  fi

  RESULTS=$(echo "$RESULTS" | jq --argjson r "$RESULT" '. + [$r]')
done

# Step 3: Build summary
if [ "$TOTAL_MAX" -gt 0 ]; then
  PERCENTAGE=$((TOTAL_SCORE * 100 / TOTAL_MAX))
else
  PERCENTAGE=100
fi

OVERALL_PASS=true
if [ "$ANY_FAILED" = true ]; then
  OVERALL_PASS=false
fi

SUMMARY="$TOTAL_SCORE/$TOTAL_MAX checks passed (${PERCENTAGE}%)"

FINAL=$(jq -n \
  --argjson checks "$RESULTS" \
  --argjson total_score "$TOTAL_SCORE" \
  --argjson max_score "$TOTAL_MAX" \
  --argjson percentage "$PERCENTAGE" \
  --argjson pass "$OVERALL_PASS" \
  --arg summary "$SUMMARY" \
  '{checks: $checks, total_score: $total_score, max_score: $max_score, percentage: $percentage, pass: $pass, summary: $summary}')

echo "$FINAL"

# Exit with failure if any check failed
if [ "$ANY_FAILED" = true ]; then
  exit 1
fi
