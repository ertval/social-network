# Social Network Refactoring — General Instructions & Reference

> Derives from [target-architecture-with-phases.md](../architecture/target-architecture-with-phases.md).
> All design decisions (D1–D6) from that document apply.

---

## Meta

| Field | Value |
|-------|-------|
| Team | 4 devs (**BE-A**, **BE-B**, **FE-A**, **FE-B**) |
| Sprint length | 2 weeks |
| Total duration | ~14 weeks (7 sprints) |
| Methodology | TDD (Red → Green → Refactor), Strangler Fig, Trunk-Based Development |
| Branch naming | `username/type-detail` (e.g. `ekaramet/feat-user-slice`) |
| Ticket format | **ID** — component, priority, dependency, assignee, story points, acceptance criteria |

---

## Refactoring Strategy & TDD Methodology

### R1: Strangler Fig Pattern

Old code is NOT deleted until new code routes traffic.

```
Step 1: Write contract tests against OLD API  (verify current behavior)
Step 2: Build new slice alongside old code   (no routing changes)
Step 3: Verify contract tests pass on NEW    (identical behavior)
Step 4: Swap routing in bootstrap.go         (one-line change)
Step 5: Monitor via tests + manual smoke     (confidence window)
Step 6: Delete old directories               (domain/, app/, infra/)
```

**Rule:** Old code exists until ALL its features are migrated. No partial deletion.

### R2: TDD Workflow (Red → Green → Refactor)

Applied to every command, query, and store method.

```
For each use case (one command/query file):

1. RED: Write a failing test
   - Test file: commands/<use_case>_test.go, queries/<use_case>_test.go, store/sqlite_test.go
   - Test valid path, invalid input, edge cases, error states
   - Use table-driven tests (Go idiom)
   - Use subtests with t.Run()

2. GREEN: Minimum code to pass
   - Write the command/query handler
   - Write the store method (with real SQLite in-memory DB)
   - Run: go test -race ./...

3. REFACTOR: Clean up
   - Extract helpers if duplicated 3+ times
   - Ensure boundary rules (D5) are intact
   - Run full CI: make ci
```

**Test file convention:**
```go
// commands/register_test.go
func TestRegisterHandler_ValidInput(t *testing.T) { ... }
func TestRegisterHandler_InvalidEmail(t *testing.T) { ... }
func TestRegisterHandler_UnderAge(t *testing.T) { ... }
func TestRegisterHandler_DuplicateUser(t *testing.T) { ... }
func TestRegisterHandler_SQLiteError(t *testing.T) { ... }
```

**Contract test pattern (for migration verification):**
```go
// internal/user/store/sqlite_migration_test.go
// Tests that new store produces identical results to old sqlite/users/userRepo.go
// These are deleted AFTER the old repo is deleted.
func TestUserStore_Migrated_SameAsOld_RegisterUser(t *testing.T) { ... }
```

### R3: Database Migration Discipline

- **Every schema change**: pair of `.up.sql` / `.down.sql` files
- **Migration ID**: sequential `000001`, `000002`, ...
- **Apply**: startup via `platform/database/migrations.go`
- **Rollback**: `go run cmd/migrate/main.go down 1`
- **Test**: each migration has integration test: apply up → verify schema → apply down → verify clean
- **Data migration safety**: never drop column in same migration; first add new column, populate, then drop old in NEXT migration

### R4: Branch Strategy

```
main ← (protected, requires CI green + review)
  ↑
  feature branch (username/type-detail)
    ↑
    WIP commits (any message)
    ↓
  squash merge into main (single Conventional Commit)
```

- Branches live ≤ 3 days (trunk-based)
- Feature toggles for incomplete work: deploy dark, activate later
- No long-lived branches

### R5: Code Review Checklist

Every PR must pass:
- [ ] **Boundary rules** (D5): no cross-slice transport/store imports
- [ ] **Interface rules** (D2): within slice = full interface, across = narrow consumer-defined
- [ ] **Cross-slice communication** (D3): ID-only refs, consumer interfaces, event bus for mutations
- [ ] **Tests present**: unit tests for each command/query, store tests with real in-memory SQLite
- [ ] **Format + lint**: `make ci` green
- [ ] **No dead code**: removed imports/variables introduced by the change

---

## Debugging & Quality Assurance Plan

### Q1: Bug Fix First (Phase 1)

| ID | Bug | Current Location | BE Assignee |
|----|-----|------------------|-------------|
| B1.1 | Migration delimiter `":"` → `";"` | `infra/storage/sqlite/init.go` | BE-A |
| B1.2 | SQLite DSN missing WAL/busy timeout | `init.go`, `.env` | BE-A |
| B1.3 | OAuth `Scan()` with `ctx` arg | `infra/storage/sqlite/oauth/oauthRepo.go` | BE-B |
| B1.4 | WebSocket CheckOrigin returns true | `infra/http/ws/handler.go` | BE-B |
| B1.5 | SQL injection in ORDER BY | `sqlite/topics/topicRepo.go`, `sqlite/categories/categoryRepo.go` | BE-A |
| B1.6 | Prepared stmt uses `db.Exec` | `sqlite/users/userRepo.go` | BE-B |
| B1.7 | WS goroutine panic recovery | `infra/ws/client.go` | BE-B |
| B1.8 | RateLimiter ticker leak (core GCRA, not HTTP wrapper) | `infra/middleware/ratelimiter/rateLimiter.go` | BE-B |

**Process:**
1. Write reproducer test (failing) for each bug
2. Apply fix
3. Verify test passes
4. Run `make ci`

### Q2: Verification Gates (per sprint)

**Mandatory:** After every sprint, before marking complete, run:

```bash
# Backend
go vet ./...
go build ./...
go test -race -coverprofile=coverage.out ./...
golangci-lint run
govulncheck ./...

# Frontend
npx @biomejs/biome lint src/
npx @biomejs/biome format src/
tsc --noEmit
npx vitest run

# Boundary check
grep -rn 'import' internal/*/transport/ internal/*/store/ | grep 'internal/' | grep -v 'platform/' | grep -v 'pkg/'
```

# Boundary check
grep -rn 'import' internal/*/transport/ internal/*/store/ | grep 'internal/' | grep -v 'platform/' | grep -v 'pkg/'
```

### Q3: Manual Smoke Test Scenarios

Run these after each feature migration to catch regression:

| Test | Steps | Expected |
|------|-------|----------|
| A1 | Register under-13 user | Rejected (age validation) |
| A2 | Register without nickname/about | Succeeds |
| A3 | Upload non-image as avatar | Rejected (magic bytes) |
| A4 | Upload valid image as avatar | Accepted |
| B1 | Set user B private → A follows | Follow request + notification |
| B2 | A views B's profile (not accepted) | "Private" lock screen |
| B3 | B accepts → A views profile | Full profile visible |
| B4 | A unfollows | Confirmation popup, relationship severed |
| C1 | Create "almost_private" post | Visible to followers, hidden from non-followers |
| C2 | Create "private" post for specific user | Visible to selected user only |
| D1 | Create group → invite member | Member gets notification, joins |
| D2 | Create event in group | All members notified |
| D3 | RSVP going/not going | Count updates in real-time |

### Q4: Contract Testing (BE ↔ FE)

- Define OpenAPI 3.0 spec for each feature endpoint in `docs/api/<feature>.yaml`
- BE tests against spec (use `kin-openapi` or manual validation)
- FE mocks from spec (use `msw` or manual mock handlers)
- CI gate: spec must match implementation (drift detection)

### Q5: Performance Regression Check

- Each sprint: run `make ci-bench` — compare against baseline from previous sprint
- Flag any regression > 10%
- Critical paths: feed query, login, WebSocket message delivery

---

## Appendix A: Best Practices Summary

### A1: Testing Pyramid

```
    ╱ E2E ╲          ~20 tests (Playwright)
   ╱─────────╲
  ╱ Integration ╲    ~50 tests (Go: wired app, FE: component composition)
 ╱───────────────╲
╱   Unit Tests    ╲  ~300+ tests (Go: per command/query/store, FE: per component)
╰─────────────────╯
```

### A2: Commit Convention

```
type(scope): description

type: feat, fix, refactor, test, chore, docs
scope: feature name (user, topic, follow, group, event, chat, notification, oauth, core, platform)
```

Examples:
- `feat(user): add register command with age validation`
- `fix(core): recover from WebSocket goroutine panic`
- `refactor(topic): migrate topic store to vertical slice`
- `test(event): add rsvp command table-driven tests`

### A3: Feature Toggle Pattern

For greenfield features that can ship dark:

```go
// bootstrap.go
if config.Features.Follow {
    follow.RegisterRoutes(router, followSvc)
}
```

### A4: Observability (Add in Sprint 1 if time)

- Structured logging: `slog` (standard library)
- Request tracing: `X-Request-ID` header, propagated through context
- Metrics: request duration, error rate, DB query time (optional: Prometheus endpoint)

### A5: Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Old code breaks during migration | Contract tests verify old API → new API behavior match |
| Migration takes too long per feature | Strangler Fig — ship one feature at a time, old code still runs |
| Breaking API change for FE | OpenAPI spec defined before BE implementation; FE mocks from spec |
| Database migration fails in production | Every migration has `.down.sql`; tested in CI |
| Performance regression | `make ci-bench` every sprint; flag > 10% degradation |
| Dev blocked waiting for other dev | Independent slices per BE dev; FE mocks BE APIs |

### A6: Definition of Done

A ticket is DONE when:
- [ ] Code written (TDD: tests first, then implementation)
- [ ] All tests pass: `make ci` for BE, `npm run lint && npm run test` for FE
- [ ] Boundary rules verified (no cross-slice transport/store imports)
- [ ] PR reviewed by other dev in same discipline (BE reviews BE, FE reviews FE)
- [ ] Merged to main via squash merge
- [ ] Deployed to dev environment (Docker Compose)
- [ ] Manual smoke test passes (relevant scenario from Q3)

---

## Appendix B: Dependency Map (Visual)

```
Sprint 0:  Setup ─────────────────────────────────────────────────────
           │ │ │ │
Sprint 1:  Platform ──────────────────────────────────────────────────
           │ │ │ │
           ├─ DB factory ──┬── Core session ──┬── All features
           ├─ EventBus     │                  │
           ├─ Cache        │                  │
           ├─ Migrations   │                  │
           └───────────────┘                  │
           │                                  │
Sprint 2:  ├── User (migration) ──────────────┤
           │  └── absorb activity             │
           ├── Topic (migration) ─────────────┤
           │  └── absorb category + vote      │
           │                                  │
Sprint 3:  ├── Follow (greenfield) ───────────┤
           ├── Comment (migration) ───────────┤
           └── Notification (migration) ──────┤
           │  └── becomes event consumer      │
           │                                  │
Sprint 4:  ├── Group (greenfield) ────────────┤
           └── Event (greenfield) ────────────┤
           │  └── depends on Group            │
           │                                  │
Sprint 5:  ├── Chat (migration) ──────────────┤
           │  └── depends on Follow           │
           └── OAuth (migration) ─────────────┤
           │                                  │
Sprint 6:  Cleanup + Integration + Docker ────┘
```

---

## Appendix C: Ticket Count Summary

| Sprint | BE Tickets | FE Tickets | DevOps | Total |
|--------|-----------|-----------|--------|-------|
| Sprint 0 | 5 | 2 | 3 | 10 |
| Sprint 1 | 11 | 4 | 0 | 15 |
| Sprint 2 | 23 | 8 | 0 | 31 |
| Sprint 3 | 26 | 8 | 0 | 34 |
| Sprint 4 | 22 | 8 | 0 | 30 |
| Sprint 5 | 17 | 7 | 0 | 24 |
| Sprint 6 | 8 | 7 | 5 | 20 |
| **Total** | **112** | **44** | **8** | **164** |
