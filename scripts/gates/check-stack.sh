#!/usr/bin/env bash
# Gate #1: Verify Go version and module path
set -euo pipefail

ERRORS=""

# Check Go version in go.mod (expect 1.24)
GOMOD_VERSION=$(grep '^go ' go.mod | awk '{print $2}')
if [[ "$GOMOD_VERSION" != "1.24" ]]; then
  ERRORS="$ERRORS\nExpected go.mod to declare Go 1.24, got $GOMOD_VERSION"
fi

# Check module path
MODULE=$(go list -m 2>/dev/null || echo "UNKNOWN")
if [[ "$MODULE" != "social-network" ]]; then
  ERRORS="$ERRORS\nExpected module path 'social-network', got '$MODULE'"
fi

if [ -n "$ERRORS" ]; then
  echo -e "FAIL:$ERRORS"
  exit 1
fi
echo "PASS"
