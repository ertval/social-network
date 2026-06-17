# Social Network Refactoring — Sprint Plan & Ticket Tracker

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

## Phase 0: Project Setup & Tooling Foundation

> **Goal:** Zero business logic. Every dev can `clone → make setup → make ci` and get a green pipeline.
> Runs in parallel with Phase 1 bug fixes — no dependency conflict.

### T0-01 — Backend Go Project Scaffold

| Field | Value |
|-------|-------|
| Component | BE (Shared) |
| Priority | P0 — Blocks everything |
| Assignee | BE-A |
| Story Points | 3 |

**Subtasks:**
- [ ] Create new Go module in project root (`go mod init`) or reuse current
- [ ] Create directory tree: `internal/{core,platform,pkg}`, `cmd/server/`, `db/migrations/`
- [ ] Install toolchain: `make tools` (goimports, staticcheck, golangci-lint v2.2.1)
- [ ] Verify go version ≥ 1.24
- [ ] Add `govulncheck` to CI

**Verify:** `go vet ./...` + `go build ./cmd/server/` compiles.

---

### T0-02 — Makefile + CI Pipeline

| Field | Value |
|-------|-------|
| Component | BE (Shared) |
| Priority | P0 — Blocks everything |
| Assignee | BE-B |
| Story Points | 2 |
| Dependencies | T0-01 |

**Subtasks:**
- [ ] Extend existing `Makefile` with new targets:
  - `make setup` — Install all tools (toolchain + bench-tools + govulncheck)
  - `make ci` — Full pipeline: `ci-mod → format → check-format → lint → test`
- [ ] Add `make test-watch` (running tests on file changes, optional)
- [ ] Add `make db-reset` — Wipe SQLite data for clean test state
- [ ] Add `make dev` — Run server with hot-reload (air or similar)

**Verify:** `make ci` passes on clean clone.

---

### T0-03 — golangci-lint Config

| Field | Value |
|-------|-------|
| Component | BE (Shared) |
| Priority | P0 |
| Assignee | BE-A |
| Story Points | 1 |
| Dependencies | T0-02 |

**Subtasks:**
- [ ] Review existing `.golangci.yml` — ensure it matches target-architecture doc (gofumpt, goimports, gci, staticcheck)
- [ ] Add linters for the new code conventions: forbid imports across slices
- [ ] Add `revive` rule to enforce boundary rules (D5)
- [ ] Set timeout ≥ 5m

**Verify:** `golangci-lint run --timeout=5m` passes.

---

### T0-04 — Next.js Frontend Scaffold

| Field | Value |
|-------|-------|
| Component | FE (Shared) |
| Priority | P0 — Blocks everything |
| Assignee | FE-A |
| Story Points | 5 |

**Subtasks:**
- [ ] Create `frontend/` directory at project root
- [ ] Init Next.js with App Router, TypeScript, Bun runtime
- [ ] Install: `shadcn/ui`, `tailwindcss`, `@biomejs/biome`
- [ ] Configure `biome.json` (linter + formatter + import sorting)
- [ ] Configure `tailwind.config.ts` with custom HSL theme (dark mode, glassmorphism)
- [ ] Configure `tsconfig.json` with strict mode, path aliases
- [ ] Configure Google Fonts (Inter or Outfit)
- [ ] Install Vitest + Playwright: `vitest.config.ts`, `playwright.config.ts`
- [ ] Add `package.json` scripts: `dev`, `build`, `lint`, `format`, `test`, `e2e`

**Verify:**
```bash
npx @biomejs/biome lint src/     # passes
npx @biomejs/biome format src/   # passes
tsc --noEmit                      # passes
npx vitest run                    # passes with 0 tests (placeholder)
```

---

### T0-05 — shadcn/ui Component Skeleton

| Field | Value |
|-------|-------|
| Component | FE (Shared) |
| Priority | P1 — Needed before any page work |
| Assignee | FE-B |
| Story Points | 3 |
| Dependencies | T0-04 |

**Subtasks:**
- [ ] Install shadcn primitives: Button, Input, Card, Dialog, DropdownMenu, Avatar, Badge, Separator, Tabs, Switch, Popover, ScrollArea, Toast, Tooltip
- [ ] Set up `src/components/ui/` with all primitives
- [ ] Create layout shell: `src/app/layout.tsx` with navigation sidebar, theme provider
- [ ] Create placeholder pages: `/login`, `/register`, `/feed`, `/profile/[id]`
- [ ] Add responsive breakpoints

**Verify:** `npm run dev` renders layout shell on port 3000. Biome + tsc pass.

---

### T0-06 — Docker Compose Development Environment

| Field | Value |
|-------|-------|
| Component | DevOps |
| Priority | P1 — Needed for integration testing |
| Assignee | BE-B + FE-A (pair) |
| Story Points | 3 |
| Dependencies | T0-01, T0-04 |

**Subtasks:**
- [ ] Create `docker-compose.dev.yml` with hot-reload for backend + frontend
- [ ] Create `docker-compose.yml` (production) per architecture doc Phase 7
- [ ] Add `scripts/docker-build.sh` for build + start
- [ ] Dev TLS: `scripts/makecerts.sh` using openssl for local HTTPS
- [ ] Document `make docker-dev` workflow

**Verify:** `docker-compose up` starts backend (8080) + frontend (3000). Both respond.

---

### T0-07 — Pre-commit Hooks + Quality Gates

| Field | Value |
|-------|-------|
| Component | Shared |
| Priority | P1 |
| Assignee | BE-A |
| Story Points | 2 |

**Subtasks:**
- [ ] Set up Husky (or lefthook) for pre-commit: format + lint staged files
- [ ] Backend: gofumpt + goimports on staged `.go` files
- [ ] Frontend: biome format + lint on staged `.ts/.tsx` files
- [ ] Pre-push: run `go vet ./...` + `tsc --noEmit` (fast gate)

**Verify:** Commit with unformatted file → hook rejects. Format → commit passes.

---

### T0-08 — Development Environment Documentation

| Field | Value |
|-------|-------|
| Component | Shared |
| Priority | P2 |
| Assignee | Rotating (first sprint retrospective) |
| Story Points | 2 |

**Subtasks:**
- [ ] Write `DEVELOPMENT.md`: one-page setup guide for new devs
- [ ] Document: required tools (go 1.24, bun, docker), `make setup`, `make dev`, `make ci`
- [ ] Document branch naming, commit conventions, PR template

---

## Refactoring Strategy & TDD Methodology

> **This section defines HOW we migrate, not just WHAT we migrate.**
> Every sprint follows these rules. No exceptions.

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
| B1.4 | WebSocket CheckOrigin returns true | `infra/ws/handler.go` | BE-B |
| B1.5 | SQL injection in ORDER BY | `sqlite/topics/topicRepo.go`, `sqlite/categories/categoryRepo.go` | BE-A |
| B1.6 | Prepared stmt uses `db.Exec` | `sqlite/users/userRepo.go` | BE-B |
| B1.7 | WS goroutine panic recovery | `infra/ws/client.go` | BE-B |
| B1.8 | RateLimiter ticker leak | `middleware/ratelimiter/rateLimiter.go` | BE-A |

**Process:**
1. Write reproducer test (failing) for each bug
2. Apply fix
3. Verify test passes
4. Run `make ci`

### Q2: Verification Gates (per sprint)

After every sprint, run the full verification from architecture doc:

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

## Sprint Plan — Parallel Tickets

> **Key:** BE-A + BE-B work on independent backend slices.
> FE-A + FE-B work on independent frontend pages.
> Frontend mocks BE APIs until they're ready.

---

## Sprint 0: Foundation (Week 1–2)

**Outcome:** All 4 devs have green pipeline. Can `make dev` and see running app.

### Backend Track

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S0-BE-01 | Go project scaffold (T0-01) | BE-A | 3 | — |
| S0-BE-02 | Makefile + CI pipeline (T0-02) | BE-B | 2 | S0-BE-01 |
| S0-BE-03 | golangci-lint config (T0-03) | BE-A | 1 | S0-BE-02 |
| S0-BE-04 | Bug fixes B1.1, B1.2, B1.5 | BE-A | 3 | S0-BE-01 |
| S0-BE-05 | Bug fixes B1.3, B1.4, B1.6, B1.7, B1.8 | BE-B | 3 | S0-BE-01 |

### Frontend Track

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S0-FE-01 | Next.js scaffold + tooling (T0-04) | FE-A | 5 | — |
| S0-FE-02 | shadcn/ui components + layout (T0-05) | FE-B | 3 | S0-FE-01 |

### DevOps Track

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S0-DEV-01 | Docker Compose dev env (T0-06) | BE-B + FE-A | 3 | S0-BE-01, S0-FE-01 |
| S0-DEV-02 | Pre-commit hooks (T0-07) | BE-A | 2 | S0-BE-01 |
| S0-DEV-03 | Dev environment docs (T0-08) | Rotating | 2 | All S0 |

**Verify Sprint 0:** `make ci` green. `make docker-dev` starts both services. Biome + tsc green on FE.

---

## Sprint 1: Platform + Core Infrastructure (Week 3–4)

**Outcome:** All platform abstractions (DB factory, eventbus, cache) + cross-cutting core (session, WS, middleware, server) are built. Features can now be built on top.

> **Critical path:** These are prerequisites for ALL feature work. Highest priority sprint.

### Backend Track

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S1-BE-01 | Platform: DB factory (`platform/database/`) | BE-A | 5 | S0 |
| S1-BE-02 | Platform: Event bus (`platform/eventbus/`) | BE-B | 3 | S0 |
| S1-BE-03 | Platform: Cache (`platform/cache/`) | BE-B | 2 | S0 |
| S1-BE-04 | Migration system (`platform/database/migrations.go` + `db/migrations/`) | BE-A | 5 | S1-BE-01 |
| S1-BE-05 | Core: Session (`core/session/`) | BE-A | 3 | S1-BE-01 |
| S1-BE-06 | Core: Realtime hub + client (`core/realtime/`) | BE-B | 5 | S0 |
| S1-BE-07 | Core: Middleware (`core/middleware/`) | BE-A | 3 | S1-BE-05 |
| S1-BE-08 | Core: Server + routes (`core/server/`) | BE-B | 3 | S1-BE-06, S1-BE-07 |
| S1-BE-09 | Shared: Rename `pkg/oAuth/` → `pkg/oauth/` + flatten | BE-B | 1 | S0 |
| S1-BE-10 | Shared: Create `pkg/imgutil/detect.go` | BE-A | 1 | S0 |

**TDD subtasks per ticket:**
1. Write interface tests (verifying the interface shape compiles)
2. Write implementation tests (real SQLite in-memory, real channels for eventbus)
3. Write integration tests (session → DB, middleware → session, WS hub → clients)
4. Implement
5. Refactor

**Key verification for S1-BE-01:**
```go
// Test that DB factory returns interface, not *sql.DB
func TestNewDB_ReturnsInterface(t *testing.T) { ... }

// Test WAL mode is enabled
func TestSQLite_WALMode(t *testing.T) { ... }

// Test busy timeout is set
func TestSQLite_BusyTimeout(t *testing.T) { ... }
```

**Key verification for S1-BE-02:**
```go
// Test publish/subscribe round-trip
func TestEventBus_PubSub(t *testing.T) { ... }

// Test concurrent publish safety
func TestEventBus_Concurrent(t *testing.T) { ... }

// Test subscriber errors don't crash bus
func TestEventBus_SubscriberPanic(t *testing.T) { ... }
```

**Key verification for S1-BE-04 (migrations):**
```go
// Test all up migrations apply in order
func TestMigrations_Up_All(t *testing.T) { ... }

// Test each migration has matching down
func TestMigrations_Down_All(t *testing.T) { ... }

// Test idempotent: running up twice doesn't fail
func TestMigrations_Up_Idempotent(t *testing.T) { ... }
```

### Frontend Track

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S1-FE-01 | Auth pages: Login + Register UI | FE-A | 5 | S0 |
| S1-FE-02 | API client layer: fetch wrapper, error handling, session cookie | FE-A | 2 | S0 |
| S1-FE-03 | Layout shell: sidebar nav, theme toggle, responsive | FE-B | 3 | S0 |
| S1-FE-04 | API mock server (msw) for BE endpoints not built yet | FE-B | 3 | S1-FE-02 |

**TDD subtasks for FE:**
1. Write component tests (Vitest + Testing Library): render, interaction, accessibility
2. Write E2E test scaffold (Playwright): login flow
3. Build components
4. Verify Biome + tsc + Vitest + Playwright pass

**Verify Sprint 1:**
- `make ci` green (all Go tests pass)
- `npm run lint && npm run test` green on FE
- Can start server, apply migrations, create session, connect WebSocket
- Login page renders and posts to mock API

---

## Sprint 2: User + Topic Features (Week 5–6)

**Outcome:** User registration/login/profile + Topic posts with visibility work end-to-end.

> Backend slices are independent → BE-A (user) and BE-B (topic) work in parallel zero contention.
> Frontend pages are independent → FE-A (profile) and FE-B (feed) work in parallel.

### Backend Track — User (`internal/user/`)

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S2-BE-01 | `user/user.go` — Entity + Repository interface | BE-A | 2 | S1 |
| S2-BE-02 | `user/store/sqlite.go` — All user SQL | BE-A | 3 | S2-BE-01 |
| S2-BE-03 | `user/commands/register.go` + test | BE-A | 3 | S2-BE-01 |
| S2-BE-04 | `user/commands/login.go` + test | BE-A | 2 | S2-BE-01 |
| S2-BE-05 | `user/commands/logout.go` + test | BE-A | 1 | S2-BE-01 |
| S2-BE-06 | `user/commands/update_profile.go` + test | BE-A | 2 | S2-BE-01 |
| S2-BE-07 | `user/commands/toggle_privacy.go` + test | BE-A | 2 | S2-BE-01 |
| S2-BE-08 | `user/queries/get_profile.go` + test (privacy lock) | BE-A | 3 | S2-BE-01 |
| S2-BE-09 | `user/queries/get_activity.go` + test | BE-A | 2 | S2-BE-01 |
| S2-BE-10 | `user/queries/list_users.go` + test | BE-A | 2 | S2-BE-01 |
| S2-BE-11 | `user/transport/http.go` — Wire all handlers | BE-A | 3 | S2-BE-03..10 |
| S2-BE-12 | User contract tests (verify against old API behavior) | BE-A | 3 | S2-BE-11 |

**TDD example for S2-BE-03 (register):**
```go
func TestRegisterHandler_ValidInput(t *testing.T) {
    // Arrange: in-memory SQLite, real bcrypt
    db := sqlite.NewInMemory(t)
    repo := user.NewSQLiteRepo(db)
    handler := user.NewRegisterHandler(repo, bcrypt.Hasher)

    // Act
    u, err := handler.Handle(ctx, RegisterInput{
        Email: "test@example.com",
        Password: "validPass1",
        FirstName: "John",
        LastName: "Doe",
        DateOfBirth: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
    })

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, u.ID)
    assert.True(t, bcrypt.Verify("validPass1", u.PasswordHash))
}

func TestRegisterHandler_UnderAge(t *testing.T) { /* < 13 → error */ }
func TestRegisterHandler_DuplicateEmail(t *testing.T) { /* same email twice */ }
func TestRegisterHandler_WeakPassword(t *testing.T) { /* < 8 chars → error */ }
func TestRegisterHandler_InvalidEmail(t *testing.T) { /* no @ → error */ }
```

### Backend Track — Topic (`internal/topic/`)

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S2-BE-13 | `topic/topic.go` — Entity (Topic, Visibility, AllowedUser, Vote) + Repo iface | BE-B | 2 | S1 |
| S2-BE-14 | `topic/store/sqlite.go` — All topic + vote SQL | BE-B | 3 | S2-BE-13 |
| S2-BE-15 | `topic/commands/create_topic.go` + test (MIME validation) | BE-B | 3 | S2-BE-13 |
| S2-BE-16 | `topic/commands/cast_vote.go` + test | BE-B | 2 | S2-BE-13 |
| S2-BE-17 | `topic/queries/get_feed.go` + test (visibility filtering) | BE-B | 5 | S2-BE-13 |
| S2-BE-18 | `topic/queries/get_user_topics.go` + test | BE-B | 2 | S2-BE-13 |
| S2-BE-19 | `topic/queries/get_topic.go` + test (privacy check) | BE-B | 2 | S2-BE-13 |
| S2-BE-20 | `topic/queries/get_votes.go` + test | BE-B | 2 | S2-BE-13 |
| S2-BE-21 | `topic/transport/http.go` — Wire all handlers | BE-B | 3 | S2-BE-15..20 |
| S2-BE-22 | Topic contract tests (verify against old API behavior) | BE-B | 3 | S2-BE-21 |

**Key test for S2-BE-17 (feed visibility):**
```go
func TestGetFeed_PublicPost_VisibleToAll(t *testing.T) { ... }
func TestGetFeed_AlmostPrivate_VisibleToFollowers(t *testing.T) { ... }
func TestGetFeed_AlmostPrivate_HiddenFromNonFollowers(t *testing.T) { ... }
func TestGetFeed_Private_VisibleToAllowedUsersOnly(t *testing.T) { ... }
func TestGetFeed_Private_HiddenFromNonAllowed(t *testing.T) { ... }
```

### Frontend Track

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S2-FE-01 | Registration form (all 8 fields, validation, avatar upload) | FE-A | 5 | S1 |
| S2-FE-02 | Login page (email + password, OAuth buttons placeholder) | FE-A | 3 | S1 |
| S2-FE-03 | Profile page `/profile/[id]` (info, activity, privacy lock) | FE-A | 5 | S2-FE-01 |
| S2-FE-04 | Privacy toggle with confirmation popup | FE-A | 2 | S2-FE-03 |
| S2-FE-05 | Home feed page (post cards, visibility indicators) | FE-B | 5 | S1 |
| S2-FE-06 | Post creation form (text, image, visibility selector, user picker) | FE-B | 5 | S2-FE-05 |
| S2-FE-07 | Post card component (content, vote buttons, comment count) | FE-B | 3 | S2-FE-05 |
| S2-FE-08 | E2E: Registration → Login → Create Post → View Feed | FE-A + FE-B | 3 | S2-FE-01..07 |

**TDD:**
- Component tests: render registration form, fill fields, submit → mock API → verify redirect
- Component tests: render feed with mock posts → verify visibility filtering
- E2E: Playwright script — full registration + login + post flow

**Verify Sprint 2:**
- `make ci` green
- Register user → login → view profile → toggle privacy → works
- Create public / almost_private / private posts → feed shows correct visibility
- Contract tests pass (new user/topic store matches old API behavior)

---

## Sprint 3: Follow + Comment + Notification (Week 7–8)

**Outcome:** Social follow system, comment on posts, and event-driven notification pipeline.

> Follow is greenfield → BE-A builds from scratch.
> Comment is migration → BE-B moves existing code.
> Notification is migration + event consumer → BE-B after comment.
> Both BE devs work independently.

### Backend Track — Follow (`internal/follow/`, greenfield)

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S3-BE-01 | `follow/follow.go` — Entities (Follow, FollowRequest) + Repo iface | BE-A | 2 | S1 |
| S3-BE-02 | `follow/store/sqlite.go` — All follow SQL | BE-A | 3 | S3-BE-01 |
| S3-BE-03 | `follow/commands/follow_user.go` + test (public=auto, private=request) | BE-A | 3 | S3-BE-01 |
| S3-BE-04 | `follow/commands/unfollow_user.go` + test | BE-A | 2 | S3-BE-01 |
| S3-BE-05 | `follow/commands/accept_request.go` + test (publish event) | BE-A | 2 | S3-BE-03 |
| S3-BE-06 | `follow/commands/decline_request.go` + test | BE-A | 1 | S3-BE-03 |
| S3-BE-07 | `follow/queries/get_followers.go` + test | BE-A | 2 | S3-BE-01 |
| S3-BE-08 | `follow/queries/get_following.go` + test | BE-A | 2 | S3-BE-01 |
| S3-BE-09 | `follow/queries/get_pending_requests.go` + test | BE-A | 2 | S3-BE-01 |
| S3-BE-10 | `follow/queries/are_connected.go` + test (FollowChecker impl) | BE-A | 2 | S3-BE-01 |
| S3-BE-11 | `follow/transport/http.go` — Wire all handlers | BE-A | 3 | S3-BE-03..10 |
| S3-BE-12 | Event publishing tests (follow.requested, follow.accepted events) | BE-A | 2 | S3-BE-03..06 |

### Backend Track — Comment (`internal/comment/`, migration) + Notification (`internal/notification/`, migration)

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S3-BE-13 | `comment/comment.go` — Entity + Repository iface | BE-B | 1 | S1 |
| S3-BE-14 | `comment/store/sqlite.go` — All comment SQL | BE-B | 2 | S3-BE-13 |
| S3-BE-15 | `comment/commands/create_comment.go` + test (MIME validation) | BE-B | 3 | S3-BE-13 |
| S3-BE-16 | `comment/queries/get_comments.go` + test | BE-B | 2 | S3-BE-13 |
| S3-BE-17 | `comment/transport/http.go` — Wire handlers | BE-B | 2 | S3-BE-15, S3-BE-16 |
| S3-BE-18 | Comment contract tests (verify against old API) | BE-B | 2 | S3-BE-17 |
| S3-BE-19 | `notification/notification.go` — Entity + Type enum + Repo iface | BE-B | 2 | S1 |
| S3-BE-20 | `notification/store/sqlite.go` — All notification SQL | BE-B | 2 | S3-BE-19 |
| S3-BE-21 | `notification/commands/consume_events.go` + test (subscribe to events) | BE-B | 5 | S3-BE-19, S3-BE-12 |
| S3-BE-22 | `notification/commands/mark_read.go` + test | BE-B | 1 | S3-BE-19 |
| S3-BE-23 | `notification/queries/list_notifications.go` + test | BE-B | 2 | S3-BE-19 |
| S3-BE-24 | `notification/transport/http.go` — Wire handlers | BE-B | 2 | S3-BE-21..23 |

**Key test for S3-BE-21 (event consumer):**
```go
func TestConsumeEvents_FollowRequested_CreatesNotification(t *testing.T) {
    bus := eventbus.NewMemoryBus()
    db := sqlite.NewInMemory(t)
    repo := notification.NewSQLiteRepo(db)
    consumer := notification.NewConsumer(repo)
    bus.Subscribe("follow.requested", consumer.OnFollowRequested)

    bus.Publish(ctx, "follow.requested", FollowRequestedPayload{FollowerID: "A", FolloweeID: "B"})
    time.Sleep(50 * time.Millisecond) // async delivery

    notifs := repo.List(ctx, "B")
    assert.Len(t, notifs, 1)
    assert.Equal(t, notification.TypeFollowRequested, notifs[0].Type)
}

func TestConsumeEvents_FollowAccepted_CreatesNotification(t *testing.T) { ... }
func TestConsumeEvents_GroupInvited_CreatesNotification(t *testing.T) { ... }
func TestConsumeEvents_GroupJoinRequested_CreatesNotification(t *testing.T) { ... }
func TestConsumeEvents_EventCreated_CreatesNotification(t *testing.T) { ... }
func TestConsumeEvents_UnsupportedType_NoError(t *testing.T) { ... }
```

### Frontend Track

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S3-FE-01 | Follow button component (toggle follow/unfollow, confirmation popup) | FE-A | 3 | S2 |
| S3-FE-02 | Followers/Following list pages | FE-A | 3 | S3-FE-01 |
| S3-FE-03 | Follow request notification card (accept/decline inline) | FE-A | 3 | S3-FE-01 |
| S3-FE-04 | Comment section (list + create comment with image/GIF) | FE-B | 5 | S2 |
| S3-FE-05 | Notification panel (sidebar dropdown, unread count badge) | FE-B | 3 | S2 |
| S3-FE-06 | SSE/WebSocket hook for live notifications | FE-B | 3 | S3-FE-05 |
| S3-FE-07 | E2E: Follow user → receive notification → accept → view profile | FE-A + FE-B | 3 | S3-FE-03 |
| S3-FE-08 | E2E: Create post → add comment → verify notification | FE-B | 2 | S3-FE-04 |

**Verify Sprint 3:**
- `make ci` green
- Follow public user → instant follow. Follow private → request created.
- Notification appears in panel when follow request/accept occurs.
- Comment created → appears in comment section.
- All contract tests pass.

---

## Sprint 4: Group + Event Features (Week 9–10)

**Outcome:** Groups with membership, chat, and event RSVP system.

> Group is greenfield → BE-A builds from scratch.
> Event is greenfield → BE-B builds from scratch (depends on group for membership check).
> FE-A builds group UI, FE-B builds event UI.

### Backend Track — Group (`internal/group/`, greenfield)

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S4-BE-01 | `group/group.go` — Entities (Group, Member, Invitation, JoinRequest) + Repo | BE-A | 2 | S1 |
| S4-BE-02 | `group/store/sqlite.go` — All group SQL | BE-A | 3 | S4-BE-01 |
| S4-BE-03 | `group/commands/create_group.go` + test (auto-add owner) | BE-A | 2 | S4-BE-01 |
| S4-BE-04 | `group/commands/invite_member.go` + test (publish event) | BE-A | 2 | S4-BE-01 |
| S4-BE-05 | `group/commands/respond_invite.go` + test | BE-A | 2 | S4-BE-04 |
| S4-BE-06 | `group/commands/request_join.go` + test (publish event) | BE-A | 2 | S4-BE-01 |
| S4-BE-07 | `group/commands/respond_join.go` + test | BE-A | 2 | S4-BE-06 |
| S4-BE-08 | `group/commands/create_group_post.go` + test (membership check) | BE-A | 3 | S4-BE-01 |
| S4-BE-09 | `group/commands/send_group_message.go` + test (membership check, WS) | BE-A | 3 | S4-BE-01 |
| S4-BE-10 | `group/queries/list_groups.go` + test | BE-A | 2 | S4-BE-01 |
| S4-BE-11 | `group/queries/get_group.go` + test | BE-A | 2 | S4-BE-01 |
| S4-BE-12 | `group/queries/get_group_feed.go` + test | BE-A | 2 | S4-BE-01 |
| S4-BE-13 | `group/queries/get_group_chat.go` + test | BE-A | 2 | S4-BE-01 |
| S4-BE-14 | `group/transport/http.go` — REST handlers | BE-A | 3 | S4-BE-03..13 |
| S4-BE-15 | `group/transport/ws.go` — Group chat WebSocket handlers | BE-A | 3 | S4-BE-14 |

### Backend Track — Event (`internal/event/`, greenfield)

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S4-BE-16 | `event/event.go` — Entities (Event, RSVP, Option) + Repo iface | BE-B | 2 | S1 |
| S4-BE-17 | `event/store/sqlite.go` — All event SQL | BE-B | 3 | S4-BE-16 |
| S4-BE-18 | `event/commands/create_event.go` + test (≥2 options, membership, publish) | BE-B | 3 | S4-BE-16, S4-BE-01 |
| S4-BE-19 | `event/commands/rsvp.go` + test | BE-B | 2 | S4-BE-16 |
| S4-BE-20 | `event/queries/list_group_events.go` + test (options + vote counts) | BE-B | 2 | S4-BE-16 |
| S4-BE-21 | `event/transport/http.go` — Wire all handlers | BE-B | 2 | S4-BE-18..20 |

**Cross-slice interface test for S4-BE-18 (event depends on group):**
```go
// event/commands/create_event.go defines:
type GroupMemberChecker interface {
    IsMember(ctx context.Context, groupID, userID string) (bool, error)
}

// event/commands/create_event_test.go
func TestCreateEvent_NonMember_Rejected(t *testing.T) {
    mockChecker := &GroupMemberCheckerMock{isMember: false}
    handler := event.NewCreateEventHandler(repo, mockChecker, bus)

    _, err := handler.Handle(ctx, input)
    assert.ErrorIs(t, err, event.ErrNotGroupMember)
}

func TestCreateEvent_LessThanTwoOptions_Rejected(t *testing.T) { ... }
func TestCreateEvent_ValidInput_PublishesEvent(t *testing.T) { ... }
```

### Frontend Track

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S4-FE-01 | Group directory page (list all groups, search) | FE-A | 3 | S2 |
| S4-FE-02 | Group page (detail, members list, join/invite buttons) | FE-A | 5 | S4-FE-01 |
| S4-FE-03 | Group post creation + feed within group | FE-A | 3 | S4-FE-02 |
| S4-FE-04 | Group chat interface (WebSocket, message list, input) | FE-A | 5 | S4-FE-02 |
| S4-FE-05 | Event creation form (title, description, datetime, options) | FE-B | 5 | S2 |
| S4-FE-06 | Event list + detail (option counts, RSVP button) | FE-B | 3 | S4-FE-05 |
| S4-FE-07 | RSVP interaction (toggle going/not going, real-time count update) | FE-B | 2 | S4-FE-06 |
| S4-FE-08 | E2E: Create group → invite user → accept → post in group → create event → RSVP | FE-A + FE-B | 3 | S4-FE-01..07 |

**Verify Sprint 4:**
- `make ci` green
- Create group → invite member → member accepts → posts visible in group feed
- Create event → group members notified → RSVP updates counts
- Group chat works via WebSocket (messages appear in real-time)
- Event notification appears in notification panel

---

## Sprint 5: Chat + OAuth (Week 11–12)

**Outcome:** 1-on-1 chat with follow-gating, and GitHub/Google OAuth login.

> Chat is migration (existing code → new slice). OAuth is migration.
> BE-A: Chat. BE-B: OAuth. Independent work.

### Backend Track — Chat (`internal/chat/`, migration)

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S5-BE-01 | `chat/chat.go` — Entity (PrivateMessage) + Repository + FollowChecker iface | BE-A | 2 | S1, S3 |
| S5-BE-02 | `chat/store/sqlite.go` — All chat SQL | BE-A | 2 | S5-BE-01 |
| S5-BE-03 | `chat/commands/send_private_msg.go` + test (follow check + WS dispatch) | BE-A | 3 | S5-BE-01 |
| S5-BE-04 | `chat/queries/get_chat_history.go` + test | BE-A | 2 | S5-BE-01 |
| S5-BE-05 | `chat/queries/list_conversations.go` + test | BE-A | 2 | S5-BE-01 |
| S5-BE-06 | `chat/transport/http.go` — REST handlers | BE-A | 2 | S5-BE-03..05 |
| S5-BE-07 | `chat/transport/ws.go` — WS handlers (migrate from infra/ws/handlers/) | BE-A | 5 | S5-BE-06 |
| S5-BE-08 | Chat contract tests (verify against old API) | BE-A | 2 | S5-BE-07 |

**Key test for S5-BE-03 (follow-gated messaging):**
```go
func TestSendPrivateMsg_NotFollowing_Rejected(t *testing.T) {
    mockChecker := &FollowCheckerMock{connected: false}
    handler := chat.NewSendPrivateMsgHandler(repo, mockChecker, hub)

    _, err := handler.Handle(ctx, input)
    assert.ErrorIs(t, err, chat.ErrNotConnected)
}

func TestSendPrivateMsg_Following_Success(t *testing.T) {
    mockChecker := &FollowCheckerMock{connected: true}
    handler := chat.NewSendPrivateMsgHandler(repo, mockChecker, hub)

    msg, err := handler.Handle(ctx, input)
    assert.NoError(t, err)
    // verify hub received dispatch call
    assert.Len(t, hub.Messages, 1)
}
```

### Backend Track — OAuth (`internal/oauth/`, migration)

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S5-BE-09 | `oauth/oauth.go` — Entity (OAuthState) + Provider enum + Repo iface | BE-B | 1 | S1 |
| S5-BE-10 | `oauth/store/sqlite.go` — OAuth SQL | BE-B | 1 | S5-BE-09 |
| S5-BE-11 | `oauth/commands/initiate.go` + test (generate state, redirect URL) | BE-B | 2 | S5-BE-09 |
| S5-BE-12 | `oauth/commands/callback.go` + test (exchange code, upsert user, session) | BE-B | 3 | S5-BE-09 |
| S5-BE-13 | `oauth/transport/http.go` — Wire handlers | BE-B | 2 | S5-BE-11, S5-BE-12 |
| S5-BE-14 | `pkg/oauth/github/client.go` — GitHub OAuth client (migrate from old) | BE-B | 2 | S1 |
| S5-BE-15 | `pkg/oauth/google/client.go` — Google OAuth client (migrate from old) | BE-B | 2 | S1 |
| S5-BE-16 | OAuth contract tests (verify against old API) | BE-B | 2 | S5-BE-13 |

### Frontend Track

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S5-FE-01 | Chat page (conversation list + message pane) | FE-A | 5 | S2 |
| S5-FE-02 | WebSocket hook (connect, send, receive, typing indicator, online presence) | FE-A | 5 | S5-FE-01 |
| S5-FE-03 | Chat message component (text, timestamp, emoji support) | FE-A | 2 | S5-FE-01 |
| S5-FE-04 | GitHub OAuth button + callback flow | FE-B | 3 | S2 |
| S5-FE-05 | Google OAuth button + callback flow | FE-B | 3 | S5-FE-04 |
| S5-FE-06 | E2E: Follow user → start chat → send message → receive reply | FE-A | 3 | S5-FE-02 |
| S5-FE-07 | E2E: Login with GitHub OAuth → complete registration | FE-B | 3 | S5-FE-04 |

**Verify Sprint 5:**
- `make ci` green
- Users who follow each other can chat via WebSocket
- Users who don't follow → "Follow to start chatting" message
- GitHub OAuth: click button → authorize → redirect → logged in
- Google OAuth: same flow
- All contract tests pass

---

## Sprint 6: Integration, Cleanup & Polish (Week 13–14)

**Outcome:** Old code deleted. Full E2E suite passes. Docker Compose production-ready. Performance verified.

### Backend Track — Cleanup

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S6-BE-01 | Delete `internal/domain/` (all features migrated) | BE-A | 2 | S5 |
| S6-BE-02 | Delete `internal/app/` (all commands/queries migrated) | BE-A | 2 | S5 |
| S6-BE-03 | Delete `internal/infra/` (all transport + storage migrated) | BE-B | 2 | S5 |
| S6-BE-04 | Bootstrap wiring (`bootstrap/bootstrap.go`) — wire ALL slices | BE-A + BE-B | 5 | S6-BE-01..03 |
| S6-BE-05 | Full integration test suite (all slices wired together) | BE-A + BE-B | 5 | S6-BE-04 |
| S6-BE-06 | Performance benchmarks (feed query, login, WS delivery) | BE-A | 3 | S6-BE-04 |
| S6-BE-07 | Run boundary verification script, fix violations | BE-B | 2 | S6-BE-04 |

**Key integration test for S6-BE-05:**
```go
func TestFullIntegration_RegisterToChat(t *testing.T) {
    // Wire everything through bootstrap
    app := bootstrap.NewTestApp(t)

    // Register 2 users
    user1 := app.Register(t, "alice@example.com", "password1", ...)
    user2 := app.Register(t, "bob@example.com", "password2", ...)

    // Follow each other
    app.Follow(t, user1.ID, user2.ID)
    app.Follow(t, user2.ID, user1.ID)

    // Send chat message
    msg := app.SendMessage(t, user1.ID, user2.ID, "Hello Bob!")
    assert.Equal(t, "Hello Bob!", msg.Content)

    // Verify notification not created (chat messages ≠ notifications)
    notifs := app.GetNotifications(t, user2.ID)
    assert.Empty(t, notifs)
}

func TestFullIntegration_PrivateProfileFollowFlow(t *testing.T) { ... }
func TestFullIntegration_GroupEventFlow(t *testing.T) { ... }
```

### Frontend Track — Polish

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S6-FE-01 | Full E2E test suite (all manual smoke scenarios automated) | FE-A + FE-B | 8 | S5 |
| S6-FE-02 | Responsive design audit (mobile + tablet + desktop) | FE-A | 3 | S5 |
| S6-FE-03 | Loading states, error boundaries, empty states for all components | FE-B | 3 | S5 |
| S6-FE-04 | Accessibility audit (keyboard nav, screen reader, contrast) | FE-A | 3 | S6-FE-02 |
| S6-FE-05 | Performance: Lighthouse audit, bundle analysis, image optimization | FE-B | 3 | S5 |
| S6-FE-06 | E2E: Smoke test all features (full user journey) | FE-A + FE-B | 5 | S6-FE-01 |

### DevOps Track — Productionization

| Ticket | Description | Assignee | SP | Depends |
|--------|-------------|----------|-----|---------|
| S6-DEV-01 | Production Docker Compose (2 services) per Phase 7 | BE-B | 2 | S6-BE-04 |
| S6-DEV-02 | Add `/healthz` + `/readyz` endpoints (K8s probes) | BE-A | 1 | S6-BE-04 |
| S6-DEV-03 | Graceful shutdown (`SIGTERM` handling, drain conns) | BE-A | 2 | S6-BE-04 |
| S6-DEV-04 | Environment-based config (12-factor, ConfigMaps ready) | BE-B | 2 | S6-BE-04 |
| S6-DEV-05 | Full `docker-compose up` → automated smoke test script | BE-B + FE-A | 3 | S6-DEV-01 |

**Verify Sprint 6:**
- `make ci` green. Full test suite passes.
- `docker-compose up` → both services running. Smoke script passes.
- Old code directories deleted. No import violations.
- E2E Playwright suite: all scenarios pass.
- Lighthouse score ≥ 90 on all pages.

---

## Post-Sprint: Optional Extensions (Phase 8–10)

These are optional learning phases from the architecture doc. File tickets when the team decides to pursue them.

### Phase 8: PostgreSQL Support

| ID | Ticket | SP |
|----|--------|-----|
| P8-01 | `platform/database/postgres.go` — connection pool | 2 |
| P8-02 | `case "postgres"` in factory | 1 |
| P8-03 | Per-feature `store/postgres.go` (one per feature, 10 total) | 5 each |
| P8-04 | Migration compatibility (SQLite SQL → PostgreSQL SQL) | 3 |
| P8-05 | Docker Compose: add postgres service | 1 |

### Phase 9: Redis Support

| ID | Ticket | SP |
|----|--------|-----|
| P9-01 | `platform/cache/redis.go` | 3 |
| P9-02 | `core/session/store/redis.go` (cache layer) | 2 |
| P9-03 | Rate limiter → Redis backend | 2 |
| P9-04 | Realtime pub/sub → Redis channels | 3 |
| P9-05 | Docker Compose: add redis service | 1 |

### Phase 10: RabbitMQ

| ID | Ticket | SP |
|----|--------|-----|
| P10-01 | `platform/rabbitmq/` client (connection, channel, reconnect) | 5 |
| P10-02 | `publisher.go` — implements `eventbus.EventBus` | 3 |
| P10-03 | `consumer.go` — dispatches to service methods | 3 |
| P10-04 | Exchange/queue/binding topology | 2 |
| P10-05 | Dead-letter handling + retry config | 3 |
| P10-06 | Swap one line in `bootstrap.go` | 1 |
| P10-07 | Docker Compose: add rabbitmq service | 1 |

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
| Sprint 1 | 10 | 4 | 0 | 14 |
| Sprint 2 | 22 | 8 | 0 | 30 |
| Sprint 3 | 24 | 8 | 0 | 32 |
| Sprint 4 | 21 | 8 | 0 | 29 |
| Sprint 5 | 16 | 7 | 0 | 23 |
| Sprint 6 | 7 | 6 | 5 | 18 |
| **Total** | **105** | **43** | **8** | **156** |

> Tickets marked as unit (1-2 SP) can be merged for efficiency. Actual count in project tracker may differ.