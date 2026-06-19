#!/usr/bin/env bash
# Gate #9: Branch naming and conventional commits
set -euo pipefail

ERRORS=""

BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Skip check if on main
if [ "$BRANCH" = "main" ] || [ "$BRANCH" = "HEAD" ]; then
  echo "PASS: on main or detached HEAD"
  exit 0
fi

# Check branch naming: <username>/<ticket-ID>-<detail>
if ! echo "$BRANCH" | grep -qE '^[a-z]+/[A-Za-z0-9]+-[a-z0-9-]+$'; then
  ERRORS="$ERRORS\nBranch name '$BRANCH' does not match convention '<username>/<ticket-ID>-<detail>'"
fi

# Check conventional commits (last 10 commits not on main)
COMMITS=$(git log main..HEAD --format='%s' 2>/dev/null || true)
if [ -n "$COMMITS" ]; then
  ALLOWED_SCOPES="user|topic|follow|group|event|chat|notification|oauth|core|platform|comment|tracker"
  while IFS= read -r msg; do
    if ! echo "$msg" | grep -qE "^(feat|fix|test|refactor|chore|docs|style|perf|ci|build|revert)\(($ALLOWED_SCOPES)\):"; then
      ERRORS="$ERRORS\nNon-conventional commit: '$msg'"
    fi
  done <<< "$COMMITS"
fi

if [ -n "$ERRORS" ]; then
  echo -e "FAIL:$ERRORS"
  exit 1
fi
echo "PASS"
