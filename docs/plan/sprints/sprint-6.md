# Sprint 6: Integration, Cleanup & Polish (Week 13–14)

**Outcome:** All legacy code structures are removed. Slices are fully integrated in `bootstrap.go`. Full automated test coverage verifies the codebase, including a specific automated test suite executing all requirements mapped from `audit.md`. Production Docker structures are deployed and verified.

> **Dependency chain warning:** Sprint 6 assumes Phase 5 migrations (Sprints 2, 3, 5) are fully complete. Cleanup tickets S6-BE-01..03 delete old layers (`domain/`, `app/`, `infra/`) — if migrations are incomplete, `bootstrap.go` still imports old packages and the build breaks. Do not start Sprint 6 until all vertical slices exist and compile independently.
>
> **S6-DEV-04 ordering:** S6-DEV-04 (12-factor env var config) changes env var names used by old code (`SERVER_HOST`, `CLIENT_HOST`, `SERVER_PORT` → `DATABASE_DRIVER`, `DATABASE_DSN`, etc.). Must run AFTER S6-BE-01..03 (old code deleted) or old code breaks. Re-order: S6-DEV-04 depends on S6-BE-03.
>
> **cmd/client/ cleanup:** Current Dockerfile builds both `cmd/server` and `cmd/client`. New 2-service design only needs `cmd/server`. S6-DEV-01 rewrites Dockerfile — include removal of `cmd/client/` binary build step. No separate cleanup ticket needed — fold into S6-DEV-01.
>
> **Sprint-level verification gate:** After every sprint, run: `go vet ./... && go build ./... && go test -race -coverprofile=coverage.out ./... && golangci-lint run`. This is required before marking the sprint complete (see general-instructions.md Q2).

---

## BE-A (Backend A) Tickets

### S6-BE-01: Clean Legacy Slices: Domain
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Safe removal of obsolete `internal/domain/` subfolders.
* **Detailed Steps:**
  1. Once all vertical slice migrations are completed and verified, completely remove the legacy `internal/domain/` directory from the repository.
* **Verification:** Run `go build ./...` and ensure there are no leftover imports referencing `internal/domain`.

---

### S6-BE-02: Clean Legacy Slices: App
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Safe removal of obsolete `internal/app/` subfolders.
* **Detailed Steps:**
  1. Delete legacy `internal/app/` containing CQRS handlers.
* **Verification:** Compilation check `go build ./...`.

---

## BE-B (Backend B) Tickets

### S6-BE-03: Clean Legacy Slices: Infra
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Safe removal of obsolete `internal/infra/` HTTP handlers and repositories.
* **Detailed Steps:**
  1. Delete legacy `internal/infra/` folder.
* **Verification:** Compilation check `go build ./...`.

---

## Joint BE-A & BE-B Tickets

### S6-BE-04: Bootstrap Wiring
* **Priority:** P0
* **Assignee:** BE-A + BE-B
* **Story Points:** 5
* **Dependencies:** S6-BE-01..03
* **Description:** Complete final wiring of all vertical slices inside the bootstrap module.
* **Detailed Steps:**
  1. Modify `internal/bootstrap/bootstrap.go`. Instantiate DB connections, EventBus implementations, session controllers, WS hubs.
  2. Wire repositories, instantiate HTTP handler routes for all slices, and start the HTTP server.
* **Verification:** Start server using `make dev` and assert that all routes are responsive.

---

## FE-A (Frontend A) Tickets

### S6-FE-02: Responsive Design Check
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Audit layouts across different viewports.
* **Detailed Steps:**
  1. Inspect responsiveness across mobile, tablet, and desktop views.
* **Verification:** Visually confirm correct scaling.

---

## FE-B (Frontend B) Tickets

### S6-FE-03: Components Error Boundaries & Loading States
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Build fallback components and loading skeletons for async cards.
* **Detailed Steps:**
  1. Integrate React Error Boundaries. Render card loading skeletons.
* **Verification:** Ensure page does not crash when backend APIs return server errors.

---

## Joint FE-A & FE-B Tickets

### S6-FE-06: Production Build Validation
* **Priority:** P1
* **Assignee:** FE-A + FE-B
* **Story Points:** 5
* **Description:** Build production bundle and execute full smoke tests.
* **Detailed Steps:**
  1. Run `bun run build`. Verify bundle compiles.
* **Verification:** Production bundle builds successfully.

---

## SD-QA (System Design/QA) Tickets

### S6-BE-05: Full Integration Test Suite
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 5
* **Dependencies:** S6-BE-04
* **Description:** Write global integration tests targeting workflows spanning across multiple slices.
* **Detailed Steps:**
  1. Implement tests where user instances are registered, follow relationships are created, chat messages are sent, and event notifications are received.
* **Verification:** Execute `make test` checking that all integration suites pass.

---

### S6-BE-06: Performance Benchmarks
* **Priority:** P2
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** Profile critical pathways (Home feed, logins, messaging).
* **Detailed Steps:**
  1. Execute performance benchmarks using `go test -bench`. Compare results against legacy benchmarks to flag regressions.
* **Verification:** Assert that latency does not increase.

---

### S6-BE-07: Vertical Slice Boundary Checks
* **Priority:** P2
* **Assignee:** SD-QA
* **Story Points:** 2
* **Description:** Automate analysis verifying slice isolation rules (e.g. no direct imports between feature stores/transports).
* **Detailed Steps:**
  1. Write shell scripts grep checks checking that imports do not violate boundary constraints. Add validation script check to the CI pipeline.
* **Verification:** Script detects and rejects circular or boundary-breaking imports.

---

### S6-BE-08: Audit.md Automation Test Suite (Gap Fix)
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 4
* **Dependencies:** S6-BE-05
* **Description:** Implement a specific automated verification script/runner that executes scenarios mapping directly to the questions checklist in `docs/requirements/audit.md`.
* **Detailed Steps:**
  1. Create a specialized integration suite under `internal/bootstrap/audit_compliance_test.go`.
  2. Implement tests mapping to each query in `audit.md`:
     - Test wrong passwords logging attempts (Q: "Did the application correctly detect and respond to the incorrect login details?")
     - Test duplicate registration blocks (Q: "Did the app detect if the email/user is already present in the database?")
     - Test profile visibility lock (Q: "Are you prevented from seeing a non-followed user private profile?")
     - Test chat relationship check blocks (Q: "Can you confirm that it was not possible to create a chat between these two users?")
     - Test WebSocket real-time delivery constraints.
* **Verification:** Running `go test -v ./internal/bootstrap -run=TestAuditCompliance` executes and passes all test scenarios.

---

### S6-FE-01: Full E2E Test Suite
* **Priority:** P0
* **Assignee:** SD-QA
* **Story Points:** 8
* **Description:** Implement Playwright automated test scripts checking standard user workflows.
* **Detailed Steps:**
  1. Write E2E browser flows: signup, signin, creating public/private posts, sending follower links, groups creations, messaging.
* **Verification:** Playwright runner executes successfully.

---

### S6-FE-04: Accessibility (a11y) Audit
* **Priority:** P2
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** Check keyboard navigation and screen-reader mappings.
* **Detailed Steps:**
  1. Ensure interactive buttons contain correct aria tags. Check color contrasts.
* **Verification:** Verify elements are fully keyboard navigable.

---

### S6-FE-05: Frontend Performance Audits
* **Priority:** P2
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** Audit bundles sizes and check asset loading.
* **Detailed Steps:**
  1. Perform Lighthouse audits. Target scoring >= 90.
* **Verification:** Lighthouse report is generated.

---

### S6-FE-07: E2E Audit.md Playwright Suite (Gap Fix)
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 4
* **Dependencies:** S6-FE-01
* **Description:** Implement Playwright browser E2E test scripts specifically mapping to the frontend verification steps listed in `docs/requirements/audit.md`.
* **Detailed Steps:**
  1. Create E2E test file `tests/audit-compliance.spec.ts`.
  2. Implement tests using Playwright viewport/device settings:
     - Open two browser contexts (e.g. User A on Chrome, User B on Firefox).
     - Test that User A logged in does not impact User B who remains guest (Q: "Can you confirm that both browsers continue with the right users?").
     - Test private profile follow acceptance flows (Q: "Is the user who received the request able to accept or decline the following request?").
     - Verify confirmation popups appear when unfollowing or toggling profile privacy (Q: "If you unfollow a user, do you get a confirmation pop-up?").
* **Verification:** Executing `npx playwright test tests/audit-compliance.spec.ts` passes.

---

### S6-DEV-01: Production Docker Setup
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 5
* **Description:** Rewrite production Docker setup from scratch per Phase 7 of the architecture plan. Old compose is single-service (forum on port 3001/8080) — replace with two-service design (backend:8080, frontend:3000).
* **Detailed Steps:**
   1. Rewrite `docker-compose.yml` with two services: backend (port 8080) and frontend (port 3000), with persistent volume for SQLite data.
   2. Create `frontend/Dockerfile` (multi-stage Node/Bun build).
   3. Update backend `Dockerfile` (multi-stage Go build -> minimal alpine image).
   4. Configure environment variables per arch spec: `DATABASE_DRIVER=sqlite`, `DATABASE_DSN=/app/data/social.db?_journal_mode=WAL&_busy_timeout=5000`, `NEXT_PUBLIC_API_URL=http://backend:8080`.
* **Verification:** Run `docker-compose up` and confirm both services connect. curl backend:8080/healthz returns 200. Frontend:3000 serves the app.

---

### S6-DEV-02: Health Check Endpoints
* **Priority:** P2
* **Assignee:** SD-QA
* **Story Points:** 1
* **Description:** Create probes reporting status.
* **Detailed Steps:**
  1. Map `GET /healthz` and `GET /readyz` routes checking database and system health.
* **Verification:** Hitting endpoints returns 200 OK statuses.

---

### S6-DEV-03: Graceful Server Shutdowns
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 2
* **Description:** Handle SIGTERM system signals.
* **Detailed Steps:**
  1. Terminate DB pools, drain WS hubs connections safely on exit.
* **Verification:** Test log indicates safe exits.

---

### S6-DEV-04: Twelve-Factor Configurations Mappings
* **Priority:** P2
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** S6-BE-03 (old code must be gone before env var rename)
* **Description:** Restructure config parameters loading strictly from environment variables, aligned with architecture spec env vars. **Must run after S6-BE-03** — old code references old env var names.
* **Detailed Steps:**
   1. Update `internal/config/config.go` to load from env vars: `DATABASE_DRIVER`, `DATABASE_DSN`, `SESSION_SECRET`, `PORT`, `CORS_ORIGIN`, `REDIS_URL` (optional), `RABBITMQ_URL` (optional).
   2. Remove legacy env var names (`SERVER_HOST`, `CLIENT_HOST`, `SERVER_PORT`).
   3. Update `docker-compose.yml` to pass new env var names (see S6-DEV-01).
* **Verification:** Configuration loads parameters correctly from env vars. Old env names produce errors.

---

### S6-DEV-05: Docker Smoke Verification Script
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 3
* **Dependencies:** S6-DEV-01
* **Description:** Automate startup checks verifying container statuses.
* **Detailed Steps:**
  1. Script that brings up docker containers, checks that `docker ps` returns active states, and runs curl queries.
* **Verification:** Automated verification script passes.
