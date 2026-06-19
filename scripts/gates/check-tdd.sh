#!/usr/bin/env bash
# Gate #6: Check that test files exist for feature slice code
set -euo pipefail

ERRORS=""
for feature_dir in internal/*/; do
  feature=$(basename "$feature_dir")
  # Skip non-feature dirs
  case "$feature" in core|platform|pkg|config|bootstrap|domain|infra|app) continue ;; esac

  # Check for test files in commands/ if commands/ exists with Go files
  if [ -d "${feature_dir}commands" ]; then
    GO_FILES=$(find "${feature_dir}commands" -name '*.go' ! -name '*_test.go' 2>/dev/null | head -1)
    if [ -n "$GO_FILES" ]; then
      TEST_FILES=$(find "${feature_dir}commands" -name '*_test.go' 2>/dev/null | head -1)
      [ -z "$TEST_FILES" ] && ERRORS="$ERRORS\n${feature_dir}commands/: has Go files but no test files"
    fi
  fi
done

if [ -n "$ERRORS" ]; then
  echo -e "FAIL: TDD violations:$ERRORS"
  exit 1
fi
echo "PASS"
