#!/usr/bin/env bash
# Gate #7: Validate database migration files
set -euo pipefail

MIGRATION_DIR="db/migrations"
ERRORS=""

# Skip if no migration directory
if [ ! -d "$MIGRATION_DIR" ]; then
  echo "PASS: No migration directory"
  exit 0
fi

# Check sequential naming and up/down pairs
for up_file in "$MIGRATION_DIR"/*.up.sql; do
  [ -f "$up_file" ] || continue
  base=$(basename "$up_file" .up.sql)
  down_file="$MIGRATION_DIR/${base}.down.sql"
  [ -f "$down_file" ] || ERRORS="$ERRORS\nMissing down migration for $up_file"
done

# Check delimiter (should be ";" not ":")
BAD_DELIMITERS=$(grep -rn '^\s*:' "$MIGRATION_DIR"/*.sql 2>/dev/null | head -5 || true)
if [ -n "$BAD_DELIMITERS" ]; then
  ERRORS="$ERRORS\nBad delimiter (use ';' not ':'):\n$BAD_DELIMITERS"
fi

if [ -n "$ERRORS" ]; then
  echo -e "FAIL:$ERRORS"
  exit 1
fi
echo "PASS"
