#!/usr/bin/env bash
# Gate #3: D5 import boundary violations via grep
set -euo pipefail

ERRORS=""

# Check that transport/ doesn't import store/
TRANSPORT_IMPORTS_STORE=$(grep -rn '^\s*import' internal/*/transport/ --include='*.go' --exclude='*_test.go' 2>/dev/null | grep 'internal/' | grep '/store' || true)
if [ -n "$TRANSPORT_IMPORTS_STORE" ]; then
  ERRORS="$ERRORS\nD5: transport imports store:\n$TRANSPORT_IMPORTS_STORE"
fi

# Check that store/ doesn't import transport/, commands/, or queries/
STORE_IMPORTS_TRANSPORT=$(grep -rn '^\s*import' internal/*/store/ --include='*.go' --exclude='*_test.go' 2>/dev/null | grep 'internal/' | grep -E '/(transport|commands|queries)' || true)
if [ -n "$STORE_IMPORTS_TRANSPORT" ]; then
  ERRORS="$ERRORS\nD5: store imports transport/commands/queries:\n$STORE_IMPORTS_TRANSPORT"
fi

# Check that commands/ and queries/ don't import store/ or transport/
CMD_IMPORTS=$(grep -rn '^\s*import' internal/*/commands/ internal/*/queries/ --include='*.go' --exclude='*_test.go' 2>/dev/null | grep 'internal/' | grep -E '/(store|transport)' || true)
if [ -n "$CMD_IMPORTS" ]; then
  ERRORS="$ERRORS\nD5: commands/queries import store/transport:\n$CMD_IMPORTS"
fi

if [ -n "$ERRORS" ]; then
  echo -e "FAIL:$ERRORS"
  exit 1
fi
echo "PASS"
