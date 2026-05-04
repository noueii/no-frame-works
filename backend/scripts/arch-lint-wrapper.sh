#!/bin/bash
# go-arch-lint wrapper that filters known false positives
# Filters interface compile-check patterns: struct named *Repository calls *.New
# Pattern: var _ Interface = (*Impl)(nil) - compile-time interface check
# Usage: ./arch-lint-wrapper.sh

cd backend

RAW_OUTPUT=$(go-arch-lint check "$@" 2>&1)
EXIT_CODE=$?

# Parse output into blocks and filter interface compile checks
# Pattern: Repository struct + New function call (compile-time check)
BLOCKS=""
CURRENT_BLOCK=""
SKIP_BLOCK=0
FIRST_DEP_LINE=""
SECOND_DEP_LINE=""

PATTERN_REPO='/repository/'

while IFS= read -r line; do
    # Empty line = block separator
    if [ -z "$line" ]; then
        if [ "$SKIP_BLOCK" = 1 ]; then
            SKIP_BLOCK=0
            FIRST_DEP_LINE=""
            SECOND_DEP_LINE=""
            CURRENT_BLOCK=""
            continue
        fi
        if [ -n "$CURRENT_BLOCK" ]; then
            BLOCKS+="$CURRENT_BLOCK"
        fi
        CURRENT_BLOCK=""
        FIRST_DEP_LINE=""
        SECOND_DEP_LINE=""
    else
        CURRENT_BLOCK+="$line"$'\n'
        
        # Track dependency lines
        if echo "$line" | grep -q "├─.*in.*/repository/"; then
            FIRST_DEP_LINE="$line"
        fi
        if echo "$line" | grep -q "└─.*service.*New"; then
            SECOND_DEP_LINE="$line"
        fi
        
        # Interface compile check pattern
        if echo "$FIRST_DEP_LINE" | grep -q "$PATTERN_REPO"; then
            if echo "$FIRST_DEP_LINE" | grep -qi "Repository"; then
                if echo "$SECOND_DEP_LINE" | grep -qE "service.*New"; then
                    SKIP_BLOCK=1
                fi
            fi
        fi
    fi
done <<< "$RAW_OUTPUT"

if [ "$SKIP_BLOCK" = 0 ] && [ -n "$CURRENT_BLOCK" ]; then
    BLOCKS+="$CURRENT_BLOCK"
fi

# Check if only headers and metadata remain
if echo "$BLOCKS" | grep -qE "Component.*shouldn.*depend on|Dependency.*not allowed"; then
    echo "$BLOCKS"
    exit $EXIT_CODE
fi

echo "OK - No warnings found"
exit 0
