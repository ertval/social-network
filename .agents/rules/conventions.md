---
trigger: always_on
glob:
description: Coding conventions, tech stack (Go, SQLite, TypeScript,Next.js), TDD workflow, database migrations, vertical slicing boundary rules, and git/branch guidelines.
---

# Conventions and Technologies

Guidelines for project technologies, refactoring patterns, and Go development best practices. For detailed instructions, see [general-instructions.md](file://docs/sprints/general-instructions.md).

## 1. Technology Stack
- **Backend**: Go (standard library preferred, `slog` for structured logging, `kin-openapi`).
- **Database**: SQLite (in-memory for unit/integration tests; WAL mode and busy timeout enabled for production/dev).
- **Frontend**: Next.js using tailwindcss, shadcn ui components, and Biome for linting/formatting.

## 2. Refactoring & Slices (KISS & Strangler Fig)
- **Strangler Fig Pattern**: Build new vertical slices alongside old code; do not delete old code until routing is fully switched and verified.
- **Vertical Slices**: Code resides in `internal/<slice>/`. D1 layout: `<feature>.go`, `commands/`, `queries/`, `transport/`, `store/`.
- **D2 Interface Strategy**:
  - Within-slice: commands/queries accept the full `Repository` interface from `<feature>.go`.
  - Across-slice: consumer defines narrow local interface. Producer satisfies via Go duck typing. Wired in `bootstrap/bootstrap.go`.
- **D3 Cross-Slice Communication**:
  - Data references: ID-only (e.g. `Comment.AuthorID string`, never `Author user.User`).
  - Sync behavior checks: consumer-defined narrow interfaces.
  - Mutation side-effects: Event Bus publish/subscribe.
- **D4 Database Factory**: Stores accept `platform/database.DB` interface, not `*sql.DB`. Factory returns `DB`; swap driver without touching feature code.
- **D5 Boundary Rules**:
  - `<feature>.go` + `commands/` + `queries/`: MUST NOT import own `transport/` or `store/`, MAY import `platform/eventbus` (interface only), MUST NOT import another feature's `transport/` or `store/`.
  - `commands/<use_case>.go`, `queries/<use_case>.go`: import own feature root, define cross-slice interfaces locally, MUST NOT import `store/` or `transport/`.
  - `transport/http.go`: imports own feature root + `commands/` + `queries/`, imports `core/session/` for auth context, MUST NOT import `store/`.
  - `store/sqlite.go`: imports own feature root + `platform/database` (DB interface), MUST NOT import `transport/` or `commands/` or `queries/`.
  - `bootstrap/bootstrap.go`: composition root — imports everything, wires concrete implementations.
- **KISS Principle**: Minimum code to solve the problem. Do not write speculative code, single-use abstractions, or unused flexibility.

## 3. TDD & Idiomatic Go
- **Red-Green-Refactor**: Always write a failing test before writing implementation code.
- **Contract Tests** (migration verification): Write tests against OLD API first, verify NEW slice produces identical results. Delete contract tests after old code is removed.
- **Test Isolation**: Store tests must run against real, isolated SQLite database instances (in-memory).
- **Go Test Style**: Use table-driven tests and subtests (`t.Run()`). Run `go test -race ./...`.
- **Surgical Changes**: Only modify lines required for the task. Remove any unused variables, functions, or imports created by your changes.

## 4. Database Migrations
- **Schema Changes**: Sequential files (`000001_name.up.sql` / `000001_name.down.sql`).
- **Safety**: Never drop a column in the same migration where it is replaced. Add, populate, then drop in a separate subsequent migration.

## 5. Code Review Checklist (R5)

Every PR must pass:
- [ ] D5 boundary rules: no cross-slice transport/store imports
- [ ] D2 interface rules: within-slice full interface, across narrow consumer-defined
- [ ] D3 cross-slice comm: ID-only refs, consumer interfaces, event bus for mutations
- [ ] Tests present: unit per command/query, store tests with real in-memory SQLite
- [ ] `make ci` green
- [ ] No dead code: remove unused imports/variables/functions from your changes

## 6. Branching & Commits
- **Branch Naming**: `<username>/<type>-<detail>` (e.g. `arnald/feat-user-slice`). Branches must live <= 3 days.
- **Commit Format**: Conventional Commits (e.g., `feat(user): add login handler`).
