#!/usr/bin/env bash
# Gate #4: D6 dependency DAG acyclicity check via go list
set -euo pipefail

FEATURES=$(ls -d internal/*/ 2>/dev/null | xargs -I{} basename {} | grep -v -E '^(core|platform|pkg|config|bootstrap|domain|infra|app)$')
ERRORS=""

for feature in $FEATURES; do
  DEPS=$(go list -f '{{join .Imports "\n"}}' "social-network/internal/$feature/..." 2>/dev/null | \
    grep "^social-network/internal/" | \
    sed 's|social-network/internal/||' | \
    cut -d/ -f1 | \
    sort -u | \
    grep -v "^$feature$" | \
    grep -v -E '^(core|platform|pkg|config|bootstrap|domain|infra|app)$')

  for dep in $DEPS; do
    # Check if dep also imports feature (cycle)
    REVERSE=$(go list -f '{{join .Imports "\n"}}' "social-network/internal/$dep/..." 2>/dev/null | \
      grep "social-network/internal/$feature" || true)
    [ -n "$REVERSE" ] && ERRORS="$ERRORS\nCIRCULAR: $feature ↔ $dep"
  done
done

# Check notification is never imported
NOTIF_IMPORTERS=$(grep -rn "social-network/internal/notification" internal/ --include="*.go" | \
  grep -v "internal/notification/" | grep -v "internal/bootstrap/" || true)
[ -n "$NOTIF_IMPORTERS" ] && ERRORS="$ERRORS\nD6: notification imported by:\n$NOTIF_IMPORTERS"

if [ -n "$ERRORS" ]; then
  echo -e "FAIL:$ERRORS"
  exit 1
fi
echo "PASS"
