#!/usr/bin/env bash
# Gate #2: Validate D1 vertical slice layout for feature directories
set -euo pipefail

ERRORS=""
for feature_dir in internal/*/; do
  feature=$(basename "$feature_dir")
  # Skip non-feature dirs
  case "$feature" in core|platform|pkg|config|bootstrap|domain|infra|app) continue ;; esac

  # Check required structure
  [ -f "${feature_dir}${feature}.go" ] || ERRORS="$ERRORS\n${feature_dir}: missing ${feature}.go (entity + repository interface)"
  [ -d "${feature_dir}commands" ] || ERRORS="$ERRORS\n${feature_dir}: missing commands/ directory"
  [ -d "${feature_dir}queries" ] || ERRORS="$ERRORS\n${feature_dir}: missing queries/ directory"
  [ -d "${feature_dir}transport" ] || ERRORS="$ERRORS\n${feature_dir}: missing transport/ directory"
  [ -d "${feature_dir}store" ] || ERRORS="$ERRORS\n${feature_dir}: missing store/ directory"
done

if [ -n "$ERRORS" ]; then
  echo -e "FAIL: D1 layout violations:$ERRORS"
  exit 1
fi
echo "PASS"
