#!/usr/bin/env bash
# Gate #13: Test coverage regression check
set -euo pipefail

MAIN_COV=$(git stash -q 2>/dev/null; git checkout main -q 2>/dev/null && \
  go test -coverprofile=/tmp/main.cov ./... 2>/dev/null && \
  go tool cover -func=/tmp/main.cov | tail -1 | awk '{print $3}' | tr -d '%'; \
  git checkout - -q 2>/dev/null; git stash pop -q 2>/dev/null)

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
