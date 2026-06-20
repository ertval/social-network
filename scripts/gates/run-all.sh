#!/usr/bin/env bash
# Master gate runner — runs all check scripts, outputs JSON summary
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS=()
OVERALL="PASS"

run_gate() {
  local name="$1"
  local script="$2"
  local output
  local status

  if [ ! -f "$script" ]; then
    RESULTS+=("{\"gate\": \"$name\", \"status\": \"SKIP\", \"message\": \"script not found\"}")
    return
  fi

  output=$(bash "$script" 2>&1) && status="PASS" || status="FAIL"
  # Escape JSON special chars in output
  # Temporarily disable pipefail to avoid SIGPIPE (exit code 141) when truncating
  set +o pipefail
  output=$(echo "$output" | head -n 5 | tr '\n' ' ' | sed 's/"/\\"/g')
  set -o pipefail

  if [ "$status" = "FAIL" ]; then
    OVERALL="FAIL"
  fi

  RESULTS+=("{\"gate\": \"$name\", \"status\": \"$status\", \"message\": \"$output\"}")
}

run_gate "stack"           "$SCRIPT_DIR/check-stack.sh"
run_gate "d1-layout"       "$SCRIPT_DIR/check-d1-layout.sh"
run_gate "d5-boundaries"   "$SCRIPT_DIR/check-d5-boundaries.sh"
run_gate "d6-dag"          "$SCRIPT_DIR/check-d6-dag.sh"
run_gate "tdd"             "$SCRIPT_DIR/check-tdd.sh"
run_gate "migrations"      "$SCRIPT_DIR/check-migrations.sh"
run_gate "security"        "$SCRIPT_DIR/check-security.sh"
run_gate "branch"          "$SCRIPT_DIR/check-branch.sh"
run_gate "scope-drift"     "$SCRIPT_DIR/check-scope-drift.sh"
run_gate "coverage-delta"  "$SCRIPT_DIR/check-coverage-delta.sh"

# Output JSON
echo "{"
echo "  \"overall\": \"$OVERALL\","
echo "  \"gates\": ["
for i in "${!RESULTS[@]}"; do
  if [ "$i" -lt $(( ${#RESULTS[@]} - 1 )) ]; then
    echo "    ${RESULTS[$i]},"
  else
    echo "    ${RESULTS[$i]}"
  fi
done
echo "  ]"
echo "}"

[ "$OVERALL" = "PASS" ] && exit 0 || exit 1
