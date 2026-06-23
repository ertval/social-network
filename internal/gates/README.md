# Verification Gates (`internal/gates/`)

Deterministic Go-based validation gates enforcing architectural rules, security posture, and code conventions. Each gate implements the `Gate` interface (`Name() string`, `Run() Result`) and registers in `cmd/gates/main.go`.

## Goal

Catch violations early — at commit time, pre-push, and CI — before code reaches review. Gates replace subjective review with objective checks: boundary rules, dependency acyclicity, security patterns, branch naming, coverage thresholds, and scope drift.

## Gate Catalog

| # | Gate | File | Rule | What It Checks |
|---|------|------|------|----------------|
| 1 | `stack` | `gate_stack.go` | Go version | `go.mod`: Go ≥ 1.24, module `social-network` |
| 2 | `d1-layout` | `gate_layout.go` | D1 — Vertical Slice Layout | Each feature has `<feature>.go`, `commands/`, `queries/`, `transport/`, `store/` |
| 3 | `d5-boundaries` | `gate_boundaries.go` | D5 — Import Boundaries | No feature root/commands/queries imports own store/transport; no store imports commands/queries/transport |
| 4 | `d6-dag` | `gate_dag.go` | D6 — Dependency DAG | Feature import graph acyclic (DFS cycle detection); no feature imports `notification` |
| 5 | `tdd` | `gate_tdd.go` | Test Coverage per Slice | Each `commands/` dir with Go files must have matching `*_test.go` |
| 6 | `migrations` | `gate_migrations.go` | Migration Integrity | Every `.up.sql` has matching `.down.sql`; no colon-terminated statements |
| 7 | `security` | `gate_security.go` | Security Patterns | `gosec` scan (if available) + AST checks: bcrypt cost ≥ 12, no SQL concat, no unconditional `CheckOrigin` |
| 8 | `branch` | `gate_branch.go` | Branch & Commits | Branch matches `<user>/<ticket-ID>-<detail>`; commits follow Conventional Commits |
| 9 | `coverage-delta` | `gate_coverage.go` | Coverage Threshold | Test coverage drop ≤ 5% vs base branch (git worktree) |
| 10 | `scope-drift` | `gate_scopedrift.go` | Scope Drift | Advisory: files changed vs base branch count |

## File Map

```
internal/gates/
├── README.md                   # This file
├── runner.go                   # Gate interface, Runner, Result, Report, JSON output
├── runner_test.go              # Runner unit tests
├── git.go                      # Git helpers: branch, log, diff, merge-base
├── git_test.go                 # Git helper tests
├── helper_test.go              # Mock exec.Command for all gate tests
├── gate_stack.go               # Gate #1: Go version + module path
├── gate_layout.go              # Gate #2: D1 — vertical slice directory structure
├── gate_boundaries.go          # Gate #3: D5 — import boundary rules (golangci-lint depguard / AST fallback)
├── gate_dag.go                 # Gate #4: D6 — dependency graph acyclicity (go-arch-lint / DFS)
├── gate_tdd.go                 # Gate #5: TDD — test file presence per commands/ dir
├── gate_migrations.go          # Gate #6: Migration naming + delimiter checks
├── gate_security.go            # Gate #7: Security — gosec + bcrypt cost + SQL concat + CheckOrigin
├── gate_branch.go              # Gate #8: Branch naming + conventional commits
├── gate_branch_test.go         # Branch gate tests
├── gate_coverage.go            # Gate #9: Coverage delta vs base branch
├── gate_coverage_test.go       # Coverage gate tests
├── gate_scopedrift.go          # Gate #10: Scope drift advisory
├── gate_scopedrift_test.go     # Scope drift gate tests
├── gate_boundaries_test.go     # Boundaries gate tests
├── gate_dag_test.go            # DAG gate tests
├── gate_layout_test.go         # Layout gate tests
├── gate_migrations_test.go     # Migrations gate tests
├── gate_security_test.go       # Security gate tests
├── gate_stack_test.go          # Stack gate tests
└── gate_tdd_test.go            # TDD gate tests
```

## How to Run

```bash
# All gates
make review-gates
# or
go run cmd/gates/main.go --all

# Single gate
go run cmd/gates/main.go --gate=boundaries
```

Output: JSON with `overall: PASS|FAIL` and per-gate results.

## Architecture

Each gate has a primary tool (e.g. `golangci-lint`, `go-arch-lint`, `gosec`) and a pure Go fallback (AST scan, DFS, `go list`). Tests mock `ExecCommand` via `helper_test.go` to verify both the primary and fallback paths without requiring external tools.

Gates are integration-tested as a Go package: `go test ./internal/gates/...` with mock CLI binaries injected through the `ExecCommand` variable.
