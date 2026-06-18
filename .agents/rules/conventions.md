---
trigger: always_on
glob:
description: Coding conventions, tech stack (Go, SQLite, TypeScript, Next.js), TDD workflow, database migrations, vertical slicing boundary rules, and git/branch guidelines.
---

# Conventions and Technologies

Guidelines for project technologies, refactoring patterns, and Go development best practices. For detailed instructions, see [general-instructions.md](file://docs/sprints/general-instructions.md).

## 1. Technology Stack
- **Backend**: Go (standard library preferred, `slog` for structured logging, `kin-openapi`).
  - **Go Version Pin**: Ensure the Go version in `go.mod` is set to `1.24`.
  - **Go Module Path**: The Go module path must be set to `social-network`.
  - **Entry Point**: The application entry point is `cmd/server/main.go`.
- **Database**: SQLite.
  - **Test Isolation**: Store tests must run against real, isolated SQLite database instances (in-memory).
  - **Production/Dev Modes**: WAL mode and busy timeout enabled.
  - **Connection Pooling**: Configure SQLite explicitly with `db.SetMaxOpenConns(1)` because SQLite does not support concurrent write operations across multiple connections.
- **Frontend**: Next.js using TailwindCSS, `shadcn/ui` components, and Biome for linting/formatting.

## 2. Refactoring & Slices (KISS & Strangler Fig)
- **Strangler Fig Pattern**: Build new vertical slices alongside old code; do not delete old code until routing is fully switched and verified.
  - **Detailed 6-Step Strangler Fig Process**:
    1. Write contract tests against the **OLD** API first to verify current behavior.
    2. Build the **NEW** vertical slice alongside the old code.
    3. Verify that contract tests pass against the **NEW** slice, yielding identical results.
    4. Swap the routing (HTTP/WS) to point to the **NEW** slice.
    5. Monitor the production/dev environment for regressions.
    6. Delete the **OLD** code once routing is fully switched, verified, and stable.
- **Endpoint Route Prefix Convention**:
  - New slices use `/api/` prefix (e.g. `/api/users`).
  - Old legacy code uses `/api/v1/` prefix.
  - During the Strangler Fig migration, both routing prefixes must coexist and route to their respective implementations.
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
    - **Documentation**: Use `// Used by: <Command/Query>` comments in store files to map methods to their commands or queries.
  - `bootstrap/bootstrap.go`: composition root — imports everything, wires concrete implementations.
- **KISS Principle**: Minimum code to solve the problem. Do not write speculative code, single-use abstractions, or unused flexibility.
- **D6 Dependency Graph**: Import tree must remain strictly acyclic. `notification` is pure event subscriber — never imported by other features. Chain: `user → follow/topic → comment/vote → group → event → chat → notification`.
- **Microservice Promotion Readiness**: Slices must access only their own database tables. Cross-slice SQL joins are strictly forbidden.
- **Event Bus Error Isolation**: All subscriber callback routines must catch panic/errors using `defer recover()` so they do not crash the publisher context.
- **Feature Toggle Pattern**: Use feature toggles for incomplete work (e.g. using `config.Features.Follow`) to deploy dark and activate later.

## 3. TDD & Idiomatic Go
- **Red-Green-Refactor**: Always write a failing test before writing implementation code.
- **Contract Tests** (migration verification): Write tests against OLD API first, verify NEW slice produces identical results. Delete contract tests after old code is removed.
- **Go Test Style**: Use table-driven tests and subtests (`t.Run()`). Run `go test -race ./...`.
- **Test Naming**: `Test<Handler>_<Scenario>` (e.g. `TestRegisterHandler_ValidInput`). Contract tests: `Test<Feature>Store_Migrated_SameAsOld_<Method>`.
- **Surgical Changes**: Only modify lines required for the task. Remove any unused variables, functions, or imports created by your changes.
- **Testing Pyramid**: Aim for ~20 E2E tests (Playwright), ~50 Integration tests, and ~300+ Unit tests.
- **OpenAPI 3.0 Contract Testing**:
  - Define an OpenAPI 3.0 spec for each feature endpoint in `docs/api/<feature>.yaml`.
  - Backend integration/transport tests must validate request/response against this spec using `kin-openapi`.
  - Frontend must mock responses from this spec using Mock Service Worker (`msw`).
- **Goroutine Panic Recovery**: All WebSocket read/write loops or client-associated goroutines must use `defer recover()` to prevent single-connection panics from crashing the server.
- **RateLimiter Ticker Leak Prevention**: Any rate limiter using `time.Ticker` must include a `stop chan struct{}` closure channel to release system threads and prevent leaks.

## 4. Database Migrations
- **Schema Changes**: Sequential files (`000001_name.up.sql` / `000001_name.down.sql`).
- **Safety**: Never drop a column in the same migration where it is replaced. Add, populate, then drop in a separate subsequent migration.
- **Delimiter**: Split migrations by `";"` (never `":"`).
- **Rollback**: Run `go run cmd/migrate/main.go down N`.
- **Test**: For each migration: apply up → verify schema → apply down → verify clean.

## 5. Code Review & Definition of Done Checklist

Every PR must pass the following **Definition of Done (DoD)** checklist:
- [ ] Conforms to D5 boundary rules (no cross-slice transport/store imports).
- [ ] Interface strategy followed (within-slice full, across-slice narrow).
- [ ] Cross-slice communication boundaries followed (ID-only refs, Event Bus for side-effects).
- [ ] Concurrency and SQLite WAL, busy timeout, and pooling rules followed (`SetMaxOpenConns(1)`).
- [ ] Unit/integration tests written and verified passing (Vitest for FE, Go test for BE).
- [ ] Type checking passes cleanly (`tsc --noEmit` for FE, `go vet` for BE).
- [ ] Format & Lint gates pass cleanly (`make ci` for BE, Biome for FE).
- [ ] Branch named correctly and commits follow conventional naming.
- [ ] PR reviewed by a developer of the same discipline.
- [ ] PR description template used and filled in completely.
- [ ] Deployed to dev environment and manual smoke tests pass successfully.
- [ ] No dead code: remove unused imports, variables, or functions created by your changes.

### Smoke Test Scenarios
- Verify manual smoke test scenarios (A1–D3) with expected results documented in `docs/sprints/general-instructions.md` before merging.

### Performance Regression Gate
- Run `make ci-bench` on each sprint/PR. Compare against the performance baseline and fail the gate if a performance regression > 10% is detected.

## 6. Branching & Commits
- **Trunk-Based Development**: Feature branches must live <= 3 days. Squash merge into `main` with a single conventional commit.
- **Branch Naming**: `<username>/<type>-<detail>` (e.g. `epapamic/feat-user-slice`).
- **Username**: Your own Gitea username — known devs: `epapamic`, `ekaramet`, `dkotsi`, `geoikonomou`, `smichail`. Use your own (e.g. `ekaramet/...`), not the `origin` remote owner.
- **Type**: `feat`, `fix`, `chore`, `refactor`, `docs`, `arch`.
- **Detail**: kebab-case. Ticket ID (e.g. `S3-fix-`) may prefix type but is not required.
- **Commit Format**: Conventional Commits (e.g., `feat(user): add login handler`). Allowed scopes: `user`, `topic`, `follow`, `group`, `event`, `chat`, `notification`, `oauth`, `core`, `platform`, `comment`.
  - *Note*: `comment` is an allowed scope. `vote` is absorbed into `topic/` — not a separate scope.
- **PR Description Template**: Copy `.github/PULL_REQUEST_TEMPLATE.md` into `.git/PR_DESCRIPTION.md` when preparing a pull request and fill in the details.

## 7. Security Best Practices
- **bcrypt Cost Factor**: Hash all user passwords using `bcrypt` with a cost factor of at least `12`.
- **Plaintext Password Memory Wiping**: Wiping plaintext password variables from memory immediately when no longer needed.
- **SQL Parameterized Queries**: Always use standard parameterized queries with `?` placeholders. String concatenation/formatting for dynamic variables in SQL queries is strictly prohibited.
- **ORDER BY Whitelist**: Whitelist dynamic order by directions to `["ASC", "DESC"]`. Never interpolate raw user input.
- **MIME Type Validation**: Validate uploaded files using `http.DetectContentType` on the first 512 bytes. Do not rely on request `Content-Type` headers. Allowed types: `image/jpeg`, `image/png`, `image/gif`.
- **WebSocket Origin Validation**: Validate WebSocket origins in `CheckOrigin`. It must not unconditionally return true; check the request header against allowed CORS origins.
- **WebSocket Timeout Constants**: Set WebSocket deadlines and rate limits: `writeWait` (10s), `pongWait` (60s), `pingPeriod` (54s, less than `pongWait`), and max message size (512KB).
- **Session Cookie Security**: Manage session state via cookies with `HttpOnly`, `Secure`, and `SameSite=Lax` attributes.

## 8. Frontend & UI Best Practices
- **Project Structure**: Follow the frontend directory layout:
  - `src/app/` (routes & pages)
  - `src/components/ui/` (reusable shadcn primitives)
  - `src/components/features/` (feature-specific components)
  - `src/lib/` (utilities & clients)
  - `src/styles/` (global CSS & Tailwind configuration)
- **Build Gates**: Enforce frontend build gates: `bun run lint`, `bun run format:check`, `tsc --noEmit`, and `bun run test`.
- **Testing Tools**: Use Vitest + React Testing Library for unit/component tests, and Playwright for E2E tests.
- **File Upload Limits**: Enforce a 10MB limit on the client side, validating file size and extension before upload.
- **Design System / CSS**: Custom HSL color palette with glassmorphism effects (`--glass-effect: backdrop-blur-md bg-slate-900/70 border border-white/10`), modern typography (Inter/Outfit), and transitions (`transition-all duration-200 ease-in-out`).
- **Destructive Operation Confirmations**: Destructive operations (e.g. unfollow, visibility/privacy toggles, declining requests) must use `shadcn/ui` Dialog overlays for user confirmation.
- **Notification Panel vs Chat**: Notifications must be displayed in a dedicated notifications panel (bell icon, unread count) that is visually distinct from the Chat panel.
- **SSE for Notifications**: Deliver real-time notifications via Server-Sent Events (`GET /api/notifications/stream`) with a 15-second polling fallback.
- **Pre-commit / Pre-push Hooks**: Use Husky/lefthook. Pre-commit runs `gofumpt`/`goimports` (BE) and `biome format`/`biome lint` (FE). Pre-push runs `go vet` (BE) and `tsc --noEmit` (FE).

## 9. Infrastructure, Scale & Observability Best Practices
- **Kubernetes Probes**: Expose `/healthz` (always returns 200 OK) and `/readyz` (runs dynamic dependency checks).
- **Graceful Shutdown**: Intercept `SIGTERM` / `SIGINT` signals to allow active requests to complete gracefully.
- **12-Factor Configurations**: The application must be configured using environment variables only. No hardcoded or config-file secrets.
- **Observability & Tracing**: Propagate `X-Request-ID` across services and requests. Log tracing context using structured `slog` fields, and expose Prometheus-compatible metrics where relevant.

