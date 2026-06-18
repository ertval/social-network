# Sync Audit: `conventions.md` vs Project Docs

**Date**: 2026-06-18
**Audited files**: `.agents/rules/conventions.md` vs architecture docs, sprint docs, requirements docs, AGENTS.md

---

## 1. Conventions in DOCS but MISSING from conventions.md

### 1.1 SQLite Max Open Connections (`SetMaxOpenConns(1)`)
- **Doc:** `docs/plan/sprints/sprint-1.md` (S1-BE-05 step 5)
- **Rule:** "Configure SQLite connection limit explicitly to prevent concurrency locking: `db.SetMaxOpenConns(1)` (since SQLite does not support concurrent write operations across multiple connections)."
- **conventions.md** mentions WAL and busy timeout but omits the single-connection pooling constraint.

### 1.2 Go Version Pin — Go 1.24
- **Doc:** `docs/plan/sprints/sprint-0.md` (S0-BE-01 step 3), `docs/plan/architecture/target-architecture-with-phases.md` (Tech table)
- **Rule:** "Ensure the Go version in `go.mod` is set to `1.24`."
- **conventions.md** doesn't specify Go version.

### 1.3 bcrypt Cost Factor ≥ 12
- **Doc:** `docs/plan/architecture/sds.md` §5.1
- **Rule:** "All user passwords must be hashed using `bcrypt` (with a cost factor of at least `12`)."
- **conventions.md** doesn't mention bcrypt cost factor.

### 1.4 Plaintext Password Memory Wiping
- **Doc:** `docs/plan/architecture/sds.md` §5.1
- **Rule:** "Plaintext password variables must be wiped immediately from memory when no longer needed."
- **conventions.md** doesn't mention this security rule.

### 1.5 Magic-Byte MIME Validation (`http.DetectContentType`)
- **Doc:** `docs/plan/architecture/sds.md` §5.2, `docs/plan/sprints/general-instructions.md` (Q1 bugs)
- **Rule:** Don't rely on `Content-Type` headers; validate using `http.DetectContentType` on first 512 bytes. Allowed: `image/jpeg`, `image/png`, `image/gif`.
- **conventions.md** doesn't mention upload validation / MIME rules.

### 1.6 SQL Parameterized Queries (No String Concatenation)
- **Doc:** `docs/plan/architecture/sds.md` §1.1
- **Rule:** "Standard parameterized queries using `?` placeholders. String concatenation or formatting for dynamic variables is strictly prohibited."
- **conventions.md** doesn't state this anti-SQL-injection rule.

### 1.7 ORDER BY Whitelist (No Raw Interpolation)
- **Doc:** `docs/plan/sprints/sprint-0.md` (S0-BE-02 B1.5), `docs/plan/architecture/target-architecture-with-phases.md` (Phase 1, bug 1.5)
- **Rule:** Whitelist `["ASC", "DESC"]` for dynamic ORDER BY; never interpolate raw input.
- **conventions.md** doesn't mention this specific SQL injection mitigation.

### 1.8 WebSocket CheckOrigin Must Validate Origins
- **Doc:** `docs/plan/architecture/sds.md` §4.1 (step 2), `docs/plan/architecture/target-architecture-with-phases.md` (Phase 1, bug 1.4)
- **Rule:** `CheckOrigin` must NOT return `true` unconditionally; must check request header against configured allowed CORS origins.
- **conventions.md** doesn't mention WebSocket origin validation.

### 1.9 Goroutine Panic Recovery (Read/Write Pumps)
- **Doc:** `docs/plan/architecture/sds.md` §4.1 (step 3), `docs/plan/architecture/target-architecture-with-phases.md` (Phase 1, bug 1.7)
- **Rule:** All client goroutine loops must `defer recover()` to prevent single-connection panics from crashing the server.
- **conventions.md** doesn't mention goroutine panic recovery.

### 1.10 RateLimiter Ticker Leak Prevention (`stop chan struct{}`)
- **Doc:** `docs/plan/architecture/sds.md` §5.3, `docs/plan/architecture/target-architecture-with-phases.md` (Phase 1, bug 1.8)
- **Rule:** Rate limiters using `time.Ticker` must include a `stop chan struct{}` closure channel to release system threads.
- **conventions.md** doesn't mention this resource-leak prevention rule.

### 1.11 WebSocket Deadlines & Rate Limits (Constants)
- **Doc:** `docs/plan/architecture/sds.md` §4.1 (step 4)
- **Rule:** `writeWait` (10s), `pongWait` (60s), `pingPeriod` (54s, < pongWait), max message size (512KB).
- **conventions.md** doesn't mention WS deadline constants or message size limits.

### 1.12 SSE for Live Notifications
- **Doc:** `docs/plan/sprints/sprint-3.md` (header note, S3-BE-57, S3-FE-18)
- **Rule:** Real-time notifications delivered via Server-Sent Events (`GET /api/notifications/stream`), with 15s polling fallback.
- **conventions.md** doesn't mention SSE pattern.

### 1.13 Session Cookie Security (HttpOnly, Secure, SameSite=Lax)
- **Doc:** `docs/plan/sprints/general-instructions.md` (F3)
- **Rule:** "Session state must be managed via cookies (`HttpOnly`, `Secure`, `SameSite=Lax`). Session must survive page refresh and prevent localStorage leakage."
- **conventions.md** doesn't mention cookie security attributes.

### 1.14 Frontend File Handling (10MB Limit, Client Validation)
- **Doc:** `docs/plan/sprints/general-instructions.md` (F4)
- **Rule:** Client validates file size (limit 10MB) and extension before upload.
- **conventions.md** doesn't mention frontend file validation rules.

### 1.15 Frontend Project Structure
- **Doc:** `docs/plan/sprints/general-instructions.md` (F5)
- **Rule:** `src/app/`, `src/components/ui/`, `src/components/features/`, `src/lib/`, `src/styles/`
- **conventions.md** mentions shadcn/Biome but not the frontend directory layout.

### 1.16 Frontend Build Gates
- **Doc:** `docs/plan/sprints/general-instructions.md` (F6)
- **Rule:** `bun run lint`, `bun run format:check`, `tsc --noEmit`, `bun run test`
- **conventions.md** mentions Biome/shadcn but not these explicit build commands.

### 1.17 Trunk-Based Development
- **Doc:** `docs/plan/sprints/general-instructions.md` (Meta, R4)
- **Rule:** "Trunk-Based Development" — branches live ≤ 3 days, squash merge, no long-lived branches.
- **conventions.md** mentions "Branches must live <= 3 days" but doesn't name the methodology "Trunk-Based Development" or the squash-merge requirement.

### 1.18 Squash Merge Policy
- **Doc:** `docs/plan/sprints/general-instructions.md` (R4, A6)
- **Rule:** Squash merge into main (single conventional commit); PR description template required.
- **conventions.md** doesn't mention squash merge or PR templates.

### 1.19 Feature Toggle Pattern
- **Doc:** `docs/plan/sprints/general-instructions.md` (R4, A3)
- **Rule:** "Feature toggles for incomplete work: deploy dark, activate later" — pattern shown with `config.Features.Follow`.
- **conventions.md** doesn't mention feature toggles.

### 1.20 Event Bus Panic/Error Isolation in Subscribers
- **Doc:** `docs/plan/sprints/sprint-1.md` (S1-BE-10 step 3)
- **Rule:** "Ensure that panic/errors in subscriber callbacks are caught using `recover()` so they do not crash the publisher context."
- **conventions.md** doesn't mention event bus error isolation.

### 1.21 Observability: `X-Request-ID`, Metrics
- **Doc:** `docs/plan/sprints/general-instructions.md` (A4)
- **Rule:** Request tracing with `X-Request-ID`, optional Prometheus metrics.
- **conventions.md** mentions `slog` for structured logging but omits `X-Request-ID` propagation and metrics.

### 1.22 Endpoint Route Prefix Convention (`/api/`)
- **Doc:** `docs/plan/sprints/sprint-3.md` (header note), various sprint route tables
- **Rule:** New slices use `/api/` prefix. Old code uses `/api/v1/`. Both coexist during migration.
- **conventions.md** doesn't document the route prefix convention.

### 1.23 Strangler Fig 6-Step Process (Detailed)
- **Doc:** `docs/plan/sprints/general-instructions.md` (R1)
- **Rule:** Explicit 6-step process: (1) contract tests against OLD, (2) build new alongside, (3) verify contract tests on NEW, (4) swap routing, (5) monitor, (6) delete old.
- **conventions.md** has a shorter version: "Build new vertical slices alongside old code; do not delete old code until routing is fully switched and verified." Missing the contract-testing-against-old step and the 6-step detail.

### 1.24 Testing Pyramid
- **Doc:** `docs/plan/sprints/general-instructions.md` (A1)
- **Rule:** ~20 E2E (Playwright), ~50 Integration, ~300+ Unit.
- **conventions.md** doesn't mention the testing pyramid or expected test counts.

### 1.25 Definition of Done Checklist
- **Doc:** `docs/plan/sprints/general-instructions.md` (A6)
- **Rule:** 7-item DoD: TDD written, all tests pass, boundary rules verified, PR reviewed by same-discipline dev, squash merged, deployed to dev, manual smoke test passes.
- **conventions.md**'s code review checklist (§5) is a subset. Missing: PR review by same discipline, deploy to dev, manual smoke test.

### 1.26 `// Used by:` Comments in Store
- **Doc:** `docs/plan/architecture/sds.md` §2.1.5
- **Rule:** "Comments like `// Used by:` make the mapping from store method → command/query slice explicit without splitting files."
- **conventions.md** doesn't mention this documentation pattern.

### 1.27 New Route Prefix vs Old `/api/v1/` Coexistence
- **Doc:** `docs/plan/sprints/sprint-3.md` (note)
- **Rule:** "New slices use `/api/` prefix. Old code uses `/api/v1/`. During Strangler Fig migration, both must coexist."
- **conventions.md** doesn't mention this coexistence pattern.

### 1.28 OpenAPI 3.0 Contract Testing (BE↔FE)
- **Doc:** `docs/plan/sprints/general-instructions.md` (Q4)
- **Rule:** "Define OpenAPI 3.0 spec for each feature endpoint in `docs/api/<feature>.yaml`. BE tests against spec (use `kin-openapi`). FE mocks from spec (use `msw`). CI gate: spec must match implementation."
- **conventions.md** mentions `kin-openapi` in tech stack but not the OpenAPI spec-first contract flow.

### 1.29 Performance Regression Check (`make ci-bench`, 10% threshold)
- **Doc:** `docs/plan/sprints/general-instructions.md` (Q5)
- **Rule:** Run `make ci-bench` each sprint, compare against baseline, flag >10% regression.
- **conventions.md** doesn't mention benchmark regression gates.

### 1.30 Pre-commit Hooks (Husky/lefthook)
- **Doc:** `docs/plan/sprints/sprint-0.md` (S0-SD-03), `docs/plan/sprints/general-instructions.md` (Q2)
- **Rule:** Pre-commit: `gofumpt`/`goimports` for BE, `biome format`/`biome lint` for FE. Pre-push: `go vet` for BE, `tsc --noEmit` for FE.
- **conventions.md** doesn't mention hook setup or pre-push checks.

### 1.31 Scenario Naming for Smoke Tests (A1–D3)
- **Doc:** `docs/plan/sprints/general-instructions.md` (Q3)
- **Rule:** Labelled scenarios A1–D3 with expected results.
- **conventions.md** doesn't mention manual smoke test scenarios.

### 1.32 PR Description Template
- **Doc:** `docs/plan/sprints/general-instructions.md` (PR Template section)
- **Rule:** Structured PR template with ticket metadata, overview, changes, audit checklist, verification results, DoD.
- **conventions.md** doesn't mention PR template.

### 1.33 Frontend CSS Variable / Glassmorphism Pattern
- **Doc:** `docs/plan/architecture/sds.md` §6.2
- **Rule:** Custom HSL color palette with `--glass-effect: backdrop-blur-md bg-slate-900/70 border border-white/10`. Modern fonts (Inter/Outfit). `transition-all duration-200 ease-in-out`.
- **conventions.md** mentions tailwindcss/shadcn but not the design system specifics.

### 1.34 Confirmation Dialog for Destructive Operations
- **Doc:** `docs/plan/architecture/sds.md` §6.3, `docs/plan/architecture/architecture.md` §6
- **Rule:** Destructive operations (unfollow, privacy toggle, decline requests) must use `shadcn/ui` Dialog overlays for confirmation.
- **conventions.md** doesn't mention this UI convention.

### 1.35 Notification Panel Distinct from Chat
- **Doc:** `docs/plan/architecture/sds.md` §6.3, `docs/plan/architecture/architecture.md` §6
- **Rule:** Notifications displayed in a dedicated panel (bell icon, unread count), visually distinct from the Chat panel.
- **conventions.md** doesn't mention this UI separation rule.

### 1.36 Frontend Testing: Vitest + Playwright
- **Doc:** `docs/plan/architecture/sds.md` §7.2, `docs/plan/sprints/general-instructions.md` (F6)
- **Rule:** Unit/component tests: Vitest + React Testing Library. E2E: Playwright.
- **conventions.md** mentions Biome for linting but omits Vitest/Playwright for testing.

### 1.37 Kubernetes Readiness (Health Probes, Graceful Shutdown, 12-Factor)
- **Doc:** `docs/plan/architecture/target-architecture-with-phases.md` (Future Scale §2)
- **Rule:** `/healthz` (200) and `/readyz` (dynamic checks). `SIGTERM`/`SIGINT` graceful shutdown. Config via env vars only.
- **conventions.md** doesn't mention health probes or 12-factor config.

### 1.38 Microservice Promotion Readiness (No Shared Storage, No Cross-Slice Joins)
- **Doc:** `docs/plan/architecture/target-architecture-with-phases.md` (Future Scale §1)
- **Rule:** Slices access only their own database tables. Cross-slice joins forbidden.
- **conventions.md** doesn't mention this forward-compatibility constraint.

### 1.39 `cmd/server/main.go` as Entry Point
- **Doc:** `docs/plan/sprints/sprint-0.md` (S0-BE-01), `docs/plan/architecture/architecture.md` §2
- **Rule:** `cmd/server/main.go` as application entry point.
- **conventions.md** doesn't specify the entry point path.

### 1.40 Module Path: `social-network`
- **Doc:** `docs/plan/sprints/sprint-0.md` (S0-BE-01 step 1)
- **Rule:** `go mod init social-network` — module path is `social-network`.
- **conventions.md** doesn't specify the Go module path.

---

## 2. Conventions in conventions.md that CONTRADICT or are OUTDATED vs docs

### 2.1 CRITICAL: Known Devs List Conflict
- **conventions.md**: "known devs: `ekaramet`, `arnald`, `dkotsi`, `ertval`"
- **AGENTS.md §7**: "known devs: `epapamic`, `ekaramet`, `dkotsi`, `geoikonomou`, `smichail`"
- **Status**: Completely different name sets. Must reconcile against `git remote -v` for authoritative list.

### 2.2 Missing `comment` and `vote` as Commit Scopes
- **conventions.md** scopes: `user`, `topic`, `follow`, `group`, `event`, `chat`, `notification`, `oauth`, `core`, `platform`
- **Sprint docs**: `comment` is a distinct vertical slice (sprint-3). `internal/comment/` exists in codebase.
- **Status**: Gap — `comment` should be an allowed scope.

### 2.3 Strangler Fig Process Oversimplified
- **conventions.md**: "Build new vertical slices alongside old code; do not delete old code until routing is fully switched and verified."
- **general-instructions.md R1**: 6-step process: (1) contract tests against OLD, (2) build new alongside, (3) verify contract tests on NEW, (4) swap routing, (5) monitor, (6) delete old.
- **Status**: conventions.md missing contract-test-against-OLD step and full 6-step detail.

### 2.4 `govulncheck` Gate Not Documented
- **conventions.md**: `make ci` green
- **general-instructions.md Q2**: `govulncheck ./...` as separate verification gate
- **Status**: conventions.md doesn't clarify that `govulncheck` is an extra gate outside `make ci`.

---

## 3. Summary of Sync Gaps

- **Critical conflict**: Known dev usernames in conventions.md vs AGENTS.md — completely different sets. Must check `git remote -v` for truth.
- **Missing security rules**: bcrypt cost ≥12, password memory wiping, MIME magic-byte validation, SQL parameterized queries, ORDER BY whitelist, WS CheckOrigin, goroutine panic recovery, rateLimiter stop channel, WS deadline constants.
- **Missing infrastructure rules**: SQLite `SetMaxOpenConns(1)`, Go 1.24 version pin, module path `social-network`, `cmd/server/main.go` entry point.
- **Missing frontend rules**: Directory layout (F5), cookie security attrs (F3), file validation 10MB limit (F4), build gate commands (F6), CSS/glassmorphism convention, confirmation dialog convention, notification/chat panel separation, Vitest/Playwright testing.
- **Missing process rules**: Squash merge, feature toggles, PR template, trunk-based development naming, testing pyramid, DoD full checklist, benchmark regression gate, pre-commit/pre-push hooks, OpenAPI contract testing flow.
- **Missing communication rules**: SSE for notifications with polling fallback, `/api/` vs `/api/v1/` route prefix coexistence, event bus subscriber panic isolation.
- **Missing forward-compat rules**: Microservice promotion (no cross-slice joins), K8s readiness (health probes, 12-factor config).
- **Missing documentation conventions**: `// Used by:` comments in store files.
- **Commit scope gap**: `comment` missing from allowed scopes despite being a distinct vertical slice.
- **Strangler Fig detail gap**: conventions.md summarizes 6-step process too briefly; missing contract-tests-against-OLD-then-verify-NEW step.
- **EventBus safety gap**: subscriber `recover()` not documented in conventions.
