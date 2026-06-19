#!/usr/bin/env bash
# Gate #8: Security pattern checks
set -euo pipefail

ERRORS=""

# Check for SQL string concatenation (potential injection)
SQL_CONCAT=$(grep -rn 'fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT\|fmt.Sprintf.*UPDATE\|fmt.Sprintf.*DELETE' internal/ --include="*.go" 2>/dev/null | \
  grep -v '_test.go' | head -10 || true)
if [ -n "$SQL_CONCAT" ]; then
  ERRORS="$ERRORS\nPotential SQL injection (string concatenation in queries):\n$SQL_CONCAT"
fi

# Check for unconditional WebSocket CheckOrigin
WS_ORIGIN=$(grep -rn 'CheckOrigin.*return true' internal/ --include="*.go" 2>/dev/null | head -5 || true)
if [ -n "$WS_ORIGIN" ]; then
  ERRORS="$ERRORS\nInsecure WebSocket CheckOrigin (unconditional true):\n$WS_ORIGIN"
fi

# Check bcrypt cost (should be >= 12)
LOW_BCRYPT=$(grep -rn 'bcrypt\.\(GenerateFromPassword\|Cost\)' internal/ --include="*.go" 2>/dev/null | \
  grep -v '_test.go' | grep -E 'cost\s*[:=]\s*[0-9]' | grep -v -E '(1[2-9]|[2-9][0-9])' | head -5 || true)
if [ -n "$LOW_BCRYPT" ]; then
  ERRORS="$ERRORS\nbcrypt cost may be < 12:\n$LOW_BCRYPT"
fi

if [ -n "$ERRORS" ]; then
  echo -e "FAIL:$ERRORS"
  exit 1
fi
echo "PASS"
