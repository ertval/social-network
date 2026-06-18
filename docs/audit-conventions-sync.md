# Sync Audit: `conventions.md` vs Project Docs

**Date**: 2026-06-18
**Audited files**: `.agents/rules/conventions.md` vs architecture docs, sprint docs, requirements docs, AGENTS.md

All mismatches, contradictions, and missing items between the core `.agents/rules/conventions.md` file and other documentation files have been fully synchronized and resolved.

---

## 1. Conventions in DOCS but MISSING from conventions.md

### [x] 1.1 SQLite Max Open Connections (`SetMaxOpenConns(1)`)
- **Status:** Resolved - Added to `conventions.md` §1 (Technology Stack).
- **Rule:** Configure SQLite explicitly with `db.SetMaxOpenConns(1)` because SQLite does not support concurrent write operations across multiple connections.

### [x] 1.2 Go Version Pin — Go 1.24
- **Status:** Resolved - Added to `conventions.md` §1 (Technology Stack).
- **Rule:** Ensure the Go version in `go.mod` is set to `1.24`.

### [x] 1.3 bcrypt Cost Factor ≥ 12
- **Status:** Resolved - Added to `conventions.md` §7 (Security Best Practices).
- **Rule:** All user passwords must be hashed using `bcrypt` (with a cost factor of at least `12`).

### [x] 1.4 Plaintext Password Memory Wiping
- **Status:** Resolved - Added to `conventions.md` §7 (Security Best Practices).
- **Rule:** Plaintext password variables must be wiped immediately from memory when no longer needed.

### [x] 1.5 Magic-Byte MIME Validation (`http.DetectContentType`)
- **Status:** Resolved - Added to `conventions.md` §7 (Security Best Practices).
- **Rule:** Validate uploaded files using `http.DetectContentType` on the first 512 bytes. Do not rely on request `Content-Type` headers. Allowed types: `image/jpeg`, `image/png`, `image/gif`.

### [x] 1.6 SQL Parameterized Queries (No String Concatenation)
- **Status:** Resolved - Added to `conventions.md` §7 (Security Best Practices).
- **Rule:** Always use standard parameterized queries with `?` placeholders. String concatenation/formatting for dynamic variables in SQL queries is strictly prohibited.

### [x] 1.7 ORDER BY Whitelist (No Raw Interpolation)
- **Status:** Resolved - Added to `conventions.md` §7 (Security Best Practices).
- **Rule:** Whitelist dynamic order by directions to `["ASC", "DESC"]`. Never interpolate raw user input.

### [x] 1.8 WebSocket CheckOrigin Must Validate Origins
- **Status:** Resolved - Added to `conventions.md` §7 (Security Best Practices).
- **Rule:** Validate WebSocket origins in `CheckOrigin`. It must not unconditionally return true; check the request header against allowed CORS origins.

### [x] 1.9 Goroutine Panic Recovery (Read/Write Pumps)
- **Status:** Resolved - Added to `conventions.md` §3 (TDD & Idiomatic Go).
- **Rule:** All client goroutine loops or WebSocket read/write loops must `defer recover()` to prevent single-connection panics from crashing the server.

### [x] 1.10 RateLimiter Ticker Leak Prevention (`stop chan struct{}`)
- **Status:** Resolved - Added to `conventions.md` §3 (TDD & Idiomatic Go).
- **Rule:** Rate limiters using `time.Ticker` must include a `stop chan struct{}` closure channel to release system threads and prevent leaks.

### [x] 1.11 WebSocket Deadlines & Rate Limits (Constants)
- **Status:** Resolved - Added to `conventions.md` §7 (Security Best Practices).
- **Rule:** Set WebSocket deadlines and rate limits: `writeWait` (10s), `pongWait` (60s), `pingPeriod` (54s, less than `pongWait`), and max message size (512KB).

### [x] 1.12 SSE for Live Notifications
- **Status:** Resolved - Added to `conventions.md` §8 (Frontend & UI Best Practices).
- **Rule:** Real-time notifications delivered via Server-Sent Events (`GET /api/notifications/stream`), with a 15-second polling fallback.

### [x] 1.13 Session Cookie Security (HttpOnly, Secure, SameSite=Lax)
- **Status:** Resolved - Added to `conventions.md` §7 (Security Best Practices).
- **Rule:** Manage session state via cookies with `HttpOnly`, `Secure`, and `SameSite=Lax` attributes.

### [x] 1.14 Frontend File Handling (10MB Limit, Client Validation)
- **Status:** Resolved - Added to `conventions.md` §8 (Frontend & UI Best Practices).
- **Rule:** Client validates file size (limit 10MB) and extension before upload.

### [x] 1.15 Frontend Project Structure
- **Status:** Resolved - Added to `conventions.md` §8 (Frontend & UI Best Practices).
- **Rule:** `src/app/`, `src/components/ui/`, `src/components/features/`, `src/lib/`, `src/styles/`

### [x] 1.16 Frontend Build Gates
- **Status:** Resolved - Added to `conventions.md` §8 (Frontend & UI Best Practices).
- **Rule:** Enforce frontend build gates: `bun run lint`, `bun run format:check`, `tsc --noEmit`, and `bun run test`.

### [x] 1.17 Trunk-Based Development
- **Status:** Resolved - Added to `conventions.md` §6 (Branching & Commits).
- **Rule:** Feature branches must live <= 3 days, squash merged into main, practicing Trunk-Based Development.

### [x] 1.18 Squash Merge Policy
- **Status:** Resolved - Added to `conventions.md` §6 (Branching & Commits).
- **Rule:** Squash merge into main (single conventional commit); PR description template required.

### [x] 1.19 Feature Toggle Pattern
- **Status:** Resolved - Added to `conventions.md` §2 (Refactoring & Slices).
- **Rule:** Use feature toggles for incomplete work (e.g. using `config.Features.Follow`) to deploy dark and activate later.

### [x] 1.20 Event Bus Panic/Error Isolation in Subscribers
- **Status:** Resolved - Added to `conventions.md` §2 (Refactoring & Slices).
- **Rule:** Subscriber callbacks must catch panic/errors using `defer recover()` so they do not crash the publisher context.

### [x] 1.21 Observability: `X-Request-ID`, Metrics
- **Status:** Resolved - Added to `conventions.md` §9 (Infrastructure, Scale & Observability Best Practices).
- **Rule:** Propagate `X-Request-ID` across services/requests. Log tracing context using structured `slog` fields, and expose Prometheus-compatible metrics where relevant.

### [x] 1.22 Endpoint Route Prefix Convention (`/api/`)
- **Status:** Resolved - Added to `conventions.md` §2 (Refactoring & Slices).
- **Rule:** New slices use `/api/` prefix. Old code uses `/api/v1/`. Both coexist during migration.

### [x] 1.23 Strangler Fig 6-Step Process (Detailed)
- **Status:** Resolved - Added to `conventions.md` §2 (Refactoring & Slices).
- **Rule:** Explicit 6-step Strangler Fig process documented in full.

### [x] 1.24 Testing Pyramid
- **Status:** Resolved - Added to `conventions.md` §3 (TDD & Idiomatic Go).
- **Rule:** Aim for ~20 E2E tests (Playwright), ~50 Integration tests, and ~300+ Unit tests.

### [x] 1.25 Definition of Done Checklist
- **Status:** Resolved - Added to `conventions.md` §5 (Code Review & Definition of Done Checklist).
- **Rule:** 12-item DoD checklist covers full BE/FE requirements including branch naming, linting, tests, PR templates, and manual smoke test runs.

### [x] 1.26 `// Used by:` Comments in Store
- **Status:** Resolved - Added to `conventions.md` §2 (Refactoring & Slices under D5 Boundary Rules).
- **Rule:** Use `// Used by: <Command/Query>` comments in store files to map methods to their commands or queries.

### [x] 1.27 New Route Prefix vs Old `/api/v1/` Coexistence
- **Status:** Resolved - Added to `conventions.md` §2 (Refactoring & Slices).
- **Rule:** During Strangler Fig migration, new slices `/api/` and legacy `/api/v1/` must coexist.

### [x] 1.28 OpenAPI 3.0 Contract Testing (BE↔FE)
- **Status:** Resolved - Added to `conventions.md` §3 (TDD & Idiomatic Go).
- **Rule:** Define OpenAPI spec in `docs/api/<feature>.yaml`. BE tests against spec (`kin-openapi`), FE mocks from spec (`msw`).

### [x] 1.29 Performance Regression Check (`make ci-bench`, 10% threshold)
- **Status:** Resolved - Added to `conventions.md` §5 (Code Review & Definition of Done Checklist).
- **Rule:** Run `make ci-bench` on each sprint/PR, and fail the gate if a regression > 10% is detected.

### [x] 1.30 Pre-commit Hooks (Husky/lefthook)
- **Status:** Resolved - Added to `conventions.md` §8 (Frontend & UI Best Practices).
- **Rule:** Pre-commit runs format/lint checks, and pre-push runs type/validation checks.

### [x] 1.31 Scenario Naming for Smoke Tests (A1–D3)
- **Status:** Resolved - Added to `conventions.md` §5 (Code Review & Definition of Done Checklist).
- **Rule:** Verify manual smoke test scenarios (A1–D3) before merging.

### [x] 1.32 PR Description Template
- **Status:** Resolved - Added to `conventions.md` §6 (Branching & Commits).
- **Rule:** Copy `.github/PULL_REQUEST_TEMPLATE.md` into `.git/PR_DESCRIPTION.md` when preparing a pull request.

### [x] 1.33 Frontend CSS Variable / Glassmorphism Pattern
- **Status:** Resolved - Added to `conventions.md` §8 (Frontend & UI Best Practices).
- **Rule:** HSL color palette with glassmorphism effects (`--glass-effect`), Inter/Outfit fonts, and custom transitions.

### [x] 1.34 Confirmation Dialog for Destructive Operations
- **Status:** Resolved - Added to `conventions.md` §8 (Frontend & UI Best Practices).
- **Rule:** Destructive operations must use `shadcn/ui` Dialog overlays for user confirmation.

### [x] 1.35 Notification Panel Distinct from Chat
- **Status:** Resolved - Added to `conventions.md` §8 (Frontend & UI Best Practices).
- **Rule:** Notifications displayed in a dedicated panel, visually distinct from Chat.

### [x] 1.36 Frontend Testing: Vitest + Playwright
- **Status:** Resolved - Added to `conventions.md` §8 (Frontend & UI Best Practices).
- **Rule:** Unit/component tests: Vitest + React Testing Library. E2E: Playwright.

### [x] 1.37 Kubernetes Readiness (Health Probes, Graceful Shutdown, 12-Factor)
- **Status:** Resolved - Added to `conventions.md` §9 (Infrastructure, Scale & Observability Best Practices).
- **Rule:** Health check probes `/healthz` & `/readyz`, `SIGTERM`/`SIGINT` graceful shutdown, and 12-factor environment configuration.

### [x] 1.38 Microservice Promotion Readiness (No Shared Storage, No Cross-Slice Joins)
- **Status:** Resolved - Added to `conventions.md` §2 (Refactoring & Slices).
- **Rule:** Slices access only their own database tables. Cross-slice joins are forbidden.

### [x] 1.39 `cmd/server/main.go` as Entry Point
- **Status:** Resolved - Added to `conventions.md` §1 (Technology Stack).
- **Rule:** Application entry point is `cmd/server/main.go`.

### [x] 1.40 Module Path: `social-network`
- **Status:** Resolved - Added to `conventions.md` §1 (Technology Stack).
- **Rule:** Go module path is `social-network`.

---

## 2. Conventions in conventions.md that CONTRADICT or are OUTDATED vs docs

### [x] 2.1 CRITICAL: Known Devs List Conflict
- **Status:** Resolved - Updated `conventions.md` §6 to list: `epapamic`, `ekaramet`, `dkotsi`, `geoikonomou`, `smichail`, matching `AGENTS.md`.

### [x] 2.2 Missing `comment` as Commit Scope
- **Status:** Resolved - Updated `conventions.md` §6 (Branching & Commits) to include `comment` scope.

### [x] 2.3 Strangler Fig Process Oversimplified
- **Status:** Resolved - Documented the complete 6-step process in `conventions.md` §2.

### [x] 2.4 `govulncheck` Gate Not Documented
- **Status:** Resolved - Part of standard `make ci` gate in `conventions.md` §5.

---

## 3. Summary of Sync Gaps

All 40 missing security, infrastructure, frontend, process, communication, forward-compatibility, and scale rules have been synchronized into `.agents/rules/conventions.md`. Gaps, outdated dev list conflicts, and commit scopes are now fully aligned across the project repository guidelines.
