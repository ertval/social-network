# Sprint 0: Foundation (Week 1)

**Outcome:** Setup project tooling, backend scaffold, frontend Next.js base, docker development environment, and fix critical bugs. All team members can run a clean build and development setup.

> **Note:** Several files referenced in Sprint 0 already exist in the repo (Makefile, `.golangci.yml`, `docker-compose.yml`, `docker-compose.dev.yml`, `frontend/`, etc.). Tickets describing creation from scratch **must audit and update** existing files rather than recreating them. Verify existing content first, then extend or fix as needed.

---

## BE-A (Backend A) Tickets

### S0-BE-01: Go Project Scaffold
* **Priority:** P0 (Blocks everything)
* **Type:** Scaffolding/Setup
* **Assignee:** BE-A
* **Story Points:** 3
* **Description:** Adjust the existing codebase layout and establish the directory structure according to the vertical-slice target architecture, reusing existing files.
* **Detailed Steps:**
  1. Initialize the module at the project root: `go mod init social-network` (or reuse the existing one, updating imports as needed).
  2. Adjust the existing structure to align with target structure:
     - `cmd/server/main.go` (main entry point)
     - `internal/core/` (sessions, middlewares, websocket base)
     - `internal/platform/` (database factory, cache, eventbus)
     - `internal/pkg/` (shared packages like hashing, uuid)
     - `db/migrations/` (SQL migration scripts)
  3. Ensure the Go version in `go.mod` is set to `1.24`.
* **Verification:** Run `go vet ./...` and `go build ./cmd/server/`. It should compile successfully (even with empty main function).

---

### S0-BE-02: Bug Fixes (B1.1, B1.2, B1.5)
* **Priority:** P0
* **Type:** Bug Fix (Existing Codebase)
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S0-BE-01
* **Description:** Fix migration path delimiter, SQLiteWAL/timeout settings, and SQL injections.
* **Detailed Steps:**
  1. **B1.1 (Migration Delimiter):** In `internal/infra/storage/sqlite/init.go`, change the SQL parsing delimiter from `":"` to `";"` so multi-statement SQL runs correctly.
  2. **B1.2 (SQLite DSN WAL/Timeout):** Modify SQLite connection setup in `internal/infra/storage/sqlite/init.go` and `.env` to pass query parameters: `?_journal_mode=WAL&_busy_timeout=5000` to prevent database locks.
  3. **B1.5 (SQL Injection):** In `internal/infra/storage/sqlite/topics/topicRepo.go` and `internal/infra/storage/sqlite/categories/categoryRepo.go`, sanitize inputs or use an allowed whitelist for dynamic `ORDER BY` directions (`ASC` or `DESC` only). Do NOT interpolate raw input directly into the query template.
* **Verification:** Write unit/integration tests reproducing the issues, ensure they fail before, and pass after implementing the fixes.

---

## BE-B (Backend B) Tickets

### S0-BE-03: Makefile + CI Pipeline
* **Priority:** P0 (Blocks everything)
* **Type:** Scaffolding/Setup
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S0-BE-01
* **Description:** Create a robust `Makefile` that handles project setup, linting, formatting, tests, and cleanups.
* **Detailed Steps:**
  1. Add standard commands:
     - `make setup` -> Install dependencies (`govulncheck`, `golangci-lint`, `goimports`).
     - `make format` -> Format backend using `gofumpt -s` and `goimports`.
     - `make check-format` -> Fail if code format changes.
     - `make lint` -> Run static checks and `golangci-lint`.
     - `make test` -> Run Go unit tests with race detector: `go test -race -cover ./...`.
     - `make be-ci` -> Backend: `go mod tidy`, format checking, linting (staticcheck + golangci-lint + govulncheck), running tests.
     - `make fe-ci` -> Frontend: Biome lint, Biome format check, `tsc --noEmit`, Vitest.
     - `make ci` -> Full gate (runs both `be-ci` and `fe-ci`).
     - `make db-reset` -> Helper to wipe local SQLite db files for fresh runs.
     - `make seed` -> seed database with test data
     - `make run-backend` -> Run backend application locally.
     - `make run-frontend` -> Run frontend application locally.
     - `make run-all` -> Run both backend and frontend concurrently locally.
     - `make dev` / `make docker-dev` -> Start development environment in Docker.
     - `make build-backend` -> Build backend application locally.
     - `make build-frontend` -> Build frontend application locally.
     - `make build` -> Build both backend and frontend locally.
     - `make docker-build` -> Build Docker images for both backend and frontend.
* **Verification:** Running `make ci` on the command line should execute both BE and FE checks and finish with exit code 0.

---

### S0-BE-04: Bug Fixes (B1.3, B1.4, B1.6, B1.7, B1.8)
* **Priority:** P0
* **Type:** Bug Fix (Existing Codebase)
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S0-BE-01
* **Description:** Fix OAuth Scanner, WebSocket origin policy, Prepared statements, WS panics, and RateLimiter leaks.
* **Detailed Steps:**
  1. **B1.3 (OAuth Scan):** In `internal/infra/storage/sqlite/oauth/oauthRepo.go`, adjust `Scan()` arguments to exclude `ctx` since the driver does not take context directly inside `Scan()`.
  2. **B1.4 (WS CheckOrigin):** In `internal/infra/http/ws/handler.go`, restrict origins. Do not return unconditional `true` for `CheckOrigin`. Read origin configuration environment variables.
  3. **B1.6 (Prepared Stmt db.Exec):** In `internal/infra/storage/sqlite/users/userRepo.go`, call `stmt.ExecContext(...)` instead of `db.Exec(...)` to ensure prepared statements are actually executed on the prepared query plan.
  4. **B1.7 (WS Panic Recovery):** In `internal/infra/ws/client.go` inside `ReadPump` and `WritePump` goroutines, add `defer func() { if r := recover(); r != nil { ... } }()` to prevent a single connection crash from bringing down the entire server.
  5. **B1.8 (RateLimiter Leak - core GCRA):** In `internal/infra/middleware/ratelimiter/rateLimiter.go` (not `internal/infra/middleware/rateLimiter.go` which is the HTTP wrapper), add a `stop` channel to close the cleanup ticker when shutting down rate limiting instances to prevent thread/memory leaks.
* **Verification:** Write specific unit tests and run `go test -race ./...` to check concurrent behaviors.

---

## FE-A (Frontend A) Tickets

### S0-FE-01: Next.js Scaffold + Tooling
* **Priority:** P0 (Blocks everything)
* **Type:** Scaffolding/Setup
* **Assignee:** FE-A
* **Story Points:** 5
* **Description:** Set up the frontend workspace using Next.js App Router, TypeScript, Tailwind CSS, Biome, and testing frameworks.
* **Detailed Steps:**
  1. Create a `frontend/` subdirectory in the project root.
  2. Bootstrap Next.js using `npx -y create-next-app@latest ./` (App router, TypeScript, Tailwind CSS, strict settings, Bun runtime compatibility).
  3. Configure Biome formatter and linter in `biome.json`. Set indentation, formatting, and strict rules. Add scripts in `package.json` to run `npx @biomejs/biome check src/`.
  4. Install `shadcn/ui` tooling and Tailwind custom presets.
  5. Install `Vitest` and `Playwright` testing setups (`vitest.config.ts`, `playwright.config.ts`).
* **Verification:** Run `bun run lint`, `bun run format`, `bun run build`, and `bun run test` to verify everything is green.

---

## FE-B (Frontend B) Tickets

### S0-FE-02: shadcn/ui Components + Layout
* **Priority:** P1
* **Type:** Scaffolding/Setup
* **Assignee:** FE-B
* **Story Points:** 3
* **Dependencies:** S0-FE-01
* **Description:** Set up core UI shell structure and shadcn primitives.
* **Detailed Steps:**
  1. Add shadcn components: Button, Input, Card, Dialog, DropdownMenu, Avatar, Badge, Switch, Toast, Tooltip.
  2. Create a global layout file `src/app/layout.tsx` containing:
     - Navigation sidebar (responsive)
     - Header area (user profile dropdown, theme switcher, notification icon placeholder)
     - Main content container (glassmorphic styling, responsive layout)
  3. Create routing page files for `/login`, `/register`, `/feed`, and `/profile/[id]`.
* **Verification:** Start the dev server (`bun run dev`) and visually check that the layout and components render correctly on port 3000.

---

## SD-QA (System Design/QA) Tickets

### S0-SD-01: golangci-lint Config
* **Priority:** P0
* **Type:** Scaffolding/Setup
* **Assignee:** SD-QA
* **Story Points:** 1
* **Dependencies:** S0-BE-03
* **Description:** Configure `.golangci.yml` at the project root to enforce code cleanliness and strict slice boundaries.
* **Detailed Steps:**
  1. Enable linters: `gofumpt`, `goimports`, `gci`, `staticcheck`, `errcheck`, `govet`, `revive`.
  2. Configure `revive` rules to enforce boundary checks (e.g. preventing direct imports between transport/store directories of different slices).
  3. Set execution timeout to 5 minutes.
* **Verification:** Execute `golangci-lint run --timeout=5m` and confirm it succeeds.

---

### S0-SD-02: Docker Compose Development Environment
* **Priority:** P1
* **Type:** Scaffolding/Setup
* **Assignee:** SD-QA
* **Story Points:** 3
* **Dependencies:** S0-BE-01, S0-FE-01
* **Description:** Set up local docker compose environments for both development (hot reload) and production.
* **Detailed Steps:**
  1. Create a `Dockerfile` for the Go backend (multi-stage build: builder -> final minimal alpine image).
  2. Create a `frontend/Dockerfile` for the Next.js frontend.
  3. Create `docker-compose.dev.yml` to mount volumes (e.g. backend code, SQLite local file, node_modules) for hot-reloading development.
  4. Create a production `docker-compose.yml` that configures environment variables, ports (8080 for backend, 3000 for frontend), and volumes for persistence.
  5. Create `scripts/docker-build.sh` to automate building and launching.
  6. Create `scripts/makecerts.sh` using openssl to issue local SSL certificates for localhost HTTPS.
* **Verification:** Run `sh scripts/docker-build.sh` and make sure both services start up and communicate properly.

---

### S0-SD-03: Pre-commit Hooks
* **Priority:** P1
* **Type:** Scaffolding/Setup
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** S0-BE-01
* **Description:** Establish quality git hooks using Husky/lefthook to prevent bad code from committing or pushing.
* **Detailed Steps:**
  1. Setup pre-commit hooks that run checks only on staged files:
     - Backend: runs `gofumpt` and `goimports` formatting.
     - Frontend: runs `biome format` and `biome lint`.
  2. Setup a pre-push hook:
     - Runs `go vet ./...` on backend.
     - Runs `tsc --noEmit` on frontend.
* **Verification:** Try to commit an unformatted file and verify the commit gets blocked. Format it, verify it succeeds.

---

### S0-SD-04: Dev Environment Docs
* **Priority:** P2
* **Type:** Scaffolding/Setup
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** All Sprint 0 tickets
* **Description:** Create onboarding documentation for the development environment.
* **Detailed Steps:**
  1. Create `DEVELOPMENT.md` in the project root.
  2. Document:
     - Prerequisite tools (Go 1.24, Bun, Docker, openssl)
     - Step-by-step setup (`make setup`, `make dev`, `make ci` (full gate))
     - Docker compose command sequences
     - Branch naming conventions and pull request rules
* **Verification:** A fresh developer should be able to set up their machine by following `DEVELOPMENT.md` alone.
