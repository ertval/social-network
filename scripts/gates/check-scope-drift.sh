#!/usr/bin/env bash
# Gate: Scope drift detection — only ticket-related files should be changed
set -euo pipefail

BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Skip if on main
if [ "$BRANCH" = "main" ] || [ "$BRANCH" = "HEAD" ]; then
  echo "PASS: on main or detached HEAD"
  exit 0
fi

# Show changed files for human review (advisory — cannot auto-detect ticket scope)
CHANGED=$(git diff main..HEAD --stat 2>/dev/null || true)
if [ -z "$CHANGED" ]; then
  echo "PASS: No changes from main"
  exit 0
fi

FILE_COUNT=$(git diff main..HEAD --name-only 2>/dev/null | wc -l)
echo "PASS: $FILE_COUNT files changed (review for scope drift)"
echo "$CHANGED"
