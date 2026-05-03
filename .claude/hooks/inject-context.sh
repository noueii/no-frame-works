#!/usr/bin/env bash
set -euo pipefail

input=$(cat)
file_path=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')

[ -z "$file_path" ] && exit 0

rules_dir="${CLAUDE_PROJECT_DIR:-$(pwd)}/.agents/rules"
rules=()

case "$file_path" in
  */services/*)   rules+=("service.md") ;;
  */repository/*) rules+=("repository.md") ;;
  */handlers/*)   rules+=("handler.md") ;;
  */domain/*)     rules+=("domain.md") ;;
esac

case "$file_path" in
  *.go) rules+=("patterns.md" "structure.md" "tx.md") ;;
esac

[ ${#rules[@]} -eq 0 ] && exit 0

context=""
for rule in "${rules[@]}"; do
  f="$rules_dir/$rule"
  [ -f "$f" ] || continue
  context+="=== .agents/rules/$rule ==="$'\n'
  context+="$(cat "$f")"$'\n\n'
done

[ -z "$context" ] && exit 0

jq -n --arg ctx "$context" '{
  hookSpecificOutput: {
    hookEventName: "PreToolUse",
    additionalContext: $ctx
  }
}'
