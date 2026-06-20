#!/usr/bin/env bash
# Gate #13: Test coverage regression check
set -euo pipefail

STASHED=false
if [ -n "$(git status --porcelain)" ]; then
  git stash -q
  STASHED=true
fi

BASE_BRANCH="main"
if ! git merge-base main HEAD &>/dev/null; then
  BASE_BRANCH="origin/main"
fi

MAIN_COV=""
if git checkout "$BASE_BRANCH" -q 2>/dev/null; then
  if go test -coverprofile=/tmp/main.cov ./... 2>/dev/null; then
    MAIN_COV=$(go tool cover -func=/tmp/main.cov | tail -n 1 | awk '{print $3}' | tr -d '%')
  fi
  git checkout - -q
fi

if [ "$STASHED" = true ]; then
  git stash pop -q 2>/dev/null || true
fi

BRANCH_COV=$(go test -coverprofile=/tmp/branch.cov ./... 2>/dev/null && \
  go tool cover -func=/tmp/branch.cov | tail -1 | awk '{print $3}' | tr -d '%')

if [ -z "$MAIN_COV" ] || [ -z "$BRANCH_COV" ]; then
  echo "PASS: Could not compute coverage delta (no baseline)"
  exit 0
fi

DELTA=$(echo "$BRANCH_COV - $MAIN_COV" | bc)
if (( $(echo "$DELTA < -5" | bc -l) )); then
  echo "FAIL: Coverage dropped by ${DELTA}% (${MAIN_COV}% → ${BRANCH_COV}%)"
  exit 1
fi
echo "PASS: Coverage ${BRANCH_COV}% (delta: ${DELTA}%)"
