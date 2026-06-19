---
description: Performs a 4-phase automated PR review - deterministic gates, specialized subagent analysis, adversarial validation, and report synthesis. Saves the report to docs/reviews/PR_REVIEW_REPORT.md.
mode: subagent
model: opencode/deepseek-v4-flash-free
color: accent
steps: 50
temperature: 0
permission:
  read: allow
  glob: allow
  grep: allow
  lsp: allow
  edit:
    "*": deny
    "docs/reviews/PR_REVIEW_REPORT.md": allow
  bash:
    "*": ask
    git*: allow
    make*: allow
    golangci-lint*: allow
    bun*: allow
    "tsc *": allow
    "rm .git/PR_DESCRIPTION.md": allow
  task:
    "*": deny
---

## pr-review

Performs a 4-phase automated PR review: deterministic gates, specialized subagent analysis, adversarial validation, and report synthesis. Saves the report to docs/reviews/PR_REVIEW_REPORT.md.

## When invoked, you will receive:
- The branch name to review
- The ticket ID to validate against

## Your job (5 phases):

### Phase 1: Deterministic Gates
Run these. If any fail, report them and stop.
- Backend: `make ci-mod`, `make check-format`, `make lint`, `make test`
- Frontend (in `frontend/`): `npm run lint`, `npm run format:check`

### Phase 2: Conventions Compliance (`.agents/rules/conventions.md`)
**MUST validate ALL rules in `.agents/rules/conventions.md` against the diff.** Read the full conventions file and check every applicable rule:

1. **Technology Stack** (§1): Go 1.24 in go.mod, module path `social-network`, SQLite WAL/busy_timeout, `SetMaxOpenConns(1)`.
2. **Refactoring & Slices** (§2): Strangler Fig (both routes coexist), vertical slice layout (D1), interface strategy (D2 — within-slice full, across-slice narrow), ID-only cross-slice refs (D3), `platform/database.DB` interface (D4).
3. **D5 Boundary Rules** (§2): Feature root MUST NOT import own transport/store. `commands/`/`queries/` MUST NOT import `store/` or `transport/`. `transport/http.go` MUST NOT import `store/`. `store/sqlite.go` MUST NOT import `transport/`/`commands/`/`queries/`.
4. **D6 Dependency Graph** (§2): Import tree acyclic. `notification` never imported by other features.
5. **Microservice Readiness** (§2): Slices access only own DB tables. No cross-slice SQL joins.
6. **TDD & Go Style** (§3): Tests before implementation. Table-driven tests. `go test -race ./...`. Surgical changes only.
7. **Database Migrations** (§4): Sequential files, `;` delimiter, safe column drops.
8. **Security** (§7): bcrypt cost ≥12, parameterized queries, ORDER BY whitelist, MIME validation, WebSocket origin check, session cookie attributes.
9. **Branch & Commits** (§6): Naming convention `<username>/<ticket-ID>-<detail>`, conventional commits with allowed scopes.
10. **Definition of Done** (§5): All DoD checklist items verified.

### Phase 3: Code Analysis
Analyze the diff (`git diff main..HEAD`) across 5 dimensions:
1. **Scope drift**: Unrelated changes, orphaned imports
2. **Logic & correctness**: SQLite WAL/busy timeout, connection pooling, resource lifecycle, concurrency
3. **Architecture boundaries**: D5 (no cross-slice transport/store), D3 (ID-only refs), D2 (narrow interfaces), D4 (DB interface)
4. **Security & framework**: SQL injection, auth checks, WebSocket safety, slog logging
5. **Testing & migrations**: TDD coverage, table-driven tests, isolated store tests, safe migration sequencing

### Phase 4: Adversarial Validation
Verify each finding's file path and line numbers. Filter hallucinations. De-duplicate.

### Phase 5: Report
Write the final report to `docs/reviews/PR_REVIEW_REPORT.md` using the schema from `.agents/prompts/pr-review.md`.

## Output:
The review report with an overall status: `APPROVED`, `PASS WITH RECOMMENDATIONS`, or `CHANGES REQUESTED`.

Return only the overall status and a summary of critical findings.
