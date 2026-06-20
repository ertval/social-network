# Sprint 6: Integration, Cleanup & Polish (Week 13–14)

**Outcome:** All legacy code structures are removed. Slices are fully integrated in `bootstrap.go`. Full automated test coverage verifies the codebase, including a specific automated test suite executing all requirements mapped from `audit.md`. Production Docker structures are deployed and verified.

> **Dependency chain warning:** Sprint 6 assumes Phase 5 migrations (Sprints 2, 3, 5) are fully complete. Cleanup tickets S6-BE-100..03 delete old layers (`domain/`, `app/`, `infra/`) — if migrations are incomplete, `bootstrap.go` still imports old packages and the build breaks. Do not start Sprint 6 until all vertical slices exist and compile independently.
>
> **S6-SD-29 ordering:** S6-SD-29 (12-factor env var config) changes env var names used by old code (`SERVER_HOST`, `CLIENT_HOST`, `SERVER_PORT` → `DATABASE_DRIVER`, `DATABASE_DSN`, etc.). Must run AFTER S6-BE-100..03 (old code deleted) or old code breaks. Re-order: S6-SD-29 depends on S6-BE-102.
>
> **cmd/client/ cleanup:** Current Dockerfile builds both `cmd/server` and `cmd/client`. New 2-service design only needs `cmd/server`. S6-SD-26 rewrites Dockerfile — include removal of `cmd/client/` binary build step. No separate cleanup ticket needed — fold into S6-SD-26.
>
> **Sprint-level verification gate:** After every sprint, run: `go vet ./... && go build ./... && go test -race -coverprofile=coverage.out ./... && golangci-lint run`. This is required before marking the sprint complete (see general-instructions.md Q2).

---

## BE-A (Backend A) Tickets

### S6-BE-100: Clean Legacy Slices: Domain
* **Priority:** P1
* **Type:** Cleanup/Integration
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Safe removal of obsolete `internal/domain/` subfolders. This is REMOVING old layered code that has been fully replaced by vertical slices — all domain logic has already been migrated to the new slice structure under `internal/user/`, `internal/topic/`, etc. The `internal/domain/` directory is dead code and can be deleted without any loss of functionality.
* **Detailed Steps:**
  1. Once all vertical slice migrations are completed and verified, completely remove the legacy `internal/domain/` directory from the repository. All functionality previously in `internal/domain/` now lives in the vertical slices — this step only deletes dead code.
* **Verification:** Run `go build ./...` and ensure there are no leftover imports referencing `internal/domain`.

---

### S6-BE-101: Clean Legacy Slices: App
* **Priority:** P1
* **Type:** Cleanup/Integration
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Safe removal of obsolete `internal/app/` subfolders. This is REMOVING old layered code that has been fully replaced by vertical slices — all CQRS handlers and application logic have already been migrated to the new slice structure. The `internal/app/` directory is dead code and can be deleted without any loss of functionality.
* **Detailed Steps:**
  1. Delete legacy `internal/app/` containing CQRS handlers. All functionality previously in `internal/app/` now lives in the vertical slices — this step only deletes dead code.
* **Verification:** Compilation check `go build ./...`.

---

## BE-B (Backend B) Tickets

### S6-BE-102: Clean Legacy Slices: Infra
* **Priority:** P1
* **Type:** Cleanup/Integration
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Safe removal of obsolete `internal/infra/` HTTP handlers and repositories. This is REMOVING old layered code that has been fully replaced by vertical slices — all HTTP handlers and repository implementations have already been migrated to the new slice structure. The `internal/infra/` directory is dead code and can be deleted without any loss of functionality.
* **Detailed Steps:**
  1. Delete legacy `internal/infra/` folder. All functionality previously in `internal/infra/` now lives in the vertical slices — this step only deletes dead code.
* **Verification:** Compilation check `go build ./...`.

---

## Joint BE-A & BE-B Tickets

### S6-BE-103: Bootstrap Wiring
* **Priority:** P0
* **Type:** Cleanup/Integration
* **Assignee:** BE-A + BE-B
* **Story Points:** 5
* **Dependencies:** S6-BE-100..03
* **Description:** Complete final wiring of all 10 vertical slices inside the bootstrap module. This is the final integration step after all old layered code has been deleted — wiring was impossible while `bootstrap.go` still imported old packages. With the old `internal/domain/`, `internal/app/`, and `internal/infra/` directories gone (S6-BE-100..03), bootstrap can now fully reference only vertical slice packages. Due to bootstrap complexity, this is done in incremental wiring steps.
* **Detailed Steps:**
  1. **Phase 1: Platform & Core:** Instantiate DB connection factory, EventBus, Cache, session manager, and the WS hub in `internal/bootstrap/bootstrap.go`. All old package imports are now removed — only vertical slice packages remain.
  2. **Phase 2: User & Topic Slices:** Wire User store/transport and Topic/Vote store/transport (with stubs replaced by real services).
  3. **Phase 3: Follow & Notification Slices:** Wire Follow store/transport, register the Notification subscriber on the EventBus, and wire Notification store/transport.
  4. **Phase 4: Group & Event Slices:** Wire Group store/transport (with WS routing and Group Commenting), and Event store/transport (injecting the GroupMemberChecker dependency).
  5. **Phase 5: Chat & OAuth Slices:** Wire Chat store/transport/WS, and OAuth state store/providers.
  6. Register all vertical slice HTTP handlers and WS routes on the central HTTP server mux and WS router.
* **Verification:** Start server using `make dev` and assert that all 10 vertical slice routes are responsive and operational.

---

## FE-A (Frontend A) Tickets

### S6-FE-33: Responsive Design Check
* **Priority:** P1
* **Type:** Testing/Verification
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Audit layouts across different viewports. This is a verification/testing ticket ensuring responsive correctness across the full application.
* **Detailed Steps:**
  1. Inspect responsiveness across mobile, tablet, and desktop views. Verify no layout breakage exists — this is a quality assurance pass, not a feature build.
* **Verification:** Visually confirm correct scaling.

---

## FE-B (Frontend B) Tickets

### S6-FE-34: Components Error Boundaries & Loading States
* **Priority:** P1
* **Type:** Cleanup/Integration (Frontend error boundaries & skeleton states)
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Build fallback components and loading skeletons for async cards. This is a frontend cleanup task adding missing production-quality error boundary and skeleton state components that were deferred during feature development.
* **Detailed Steps:**
  1. Integrate React Error Boundaries. Render card loading skeletons. These are production hardening components, not new features — they fill gaps left during vertical slice frontend implementation.
* **Verification:** Ensure page does not crash when backend APIs return server errors.

---

## Joint FE-A & FE-B Tickets

### S6-FE-35: Production Build Validation
* **Priority:** P1
* **Type:** Testing/Verification (Production build validation)
* **Assignee:** FE-A + FE-B
* **Story Points:** 5
* **Description:** Build production bundle and execute smoke tests. This is a build validation test verifying that the production frontend artifact compiles, bundles correctly, and passes smoke checks.
* **Detailed Steps:**
  1. Run `bun run build`. Verify bundle compiles. This is a verification-only pass — catch build failures, missing imports, and bundling regressions before deployment.
* **Verification:** Production bundle builds successfully.

---

## SD-QA (System Design/QA) Tickets

### S6-SD-22: Full Integration Test Suite
* **Priority:** P1
* **Type:** Testing/Verification
* **Assignee:** SD-QA
* **Story Points:** 5
* **Dependencies:** S6-BE-103
* **Description:** Write global integration tests targeting workflows spanning across multiple slices. This is a verification/testing ticket ensuring full cross-slice system correctness after all vertical slice migrations and old code removal are complete.
* **Detailed Steps:**
  1. Implement tests where user instances are registered, follow relationships are created, chat messages are sent, and event notifications are received. These integration tests validate that slices work together correctly end-to-end.
* **Verification:** Execute `make test` checking that all integration suites pass.

---

### S6-SD-23: Performance Benchmarks
* **Priority:** P2
* **Type:** Testing/Verification
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** Profile critical pathways (Home feed, logins, messaging). This is a verification/testing ticket ensuring no performance regressions were introduced during the vertical slice migration.
* **Detailed Steps:**
  1. Execute performance benchmarks using `go test -bench`. Compare results against legacy benchmarks to flag regressions. This is a performance verification pass — not an optimization task.
* **Verification:** Assert that latency does not increase.

---

### S6-SD-24: Vertical Slice Boundary Checks
* **Priority:** P2
* **Type:** Testing/Verification
* **Assignee:** SD-QA
* **Story Points:** 2
* **Description:** Automate analysis verifying slice isolation rules (e.g. no direct imports between feature stores/transports). This is a verification/testing ticket ensuring architectural boundary constraints are enforced across all vertical slices.
* **Detailed Steps:**
  1. Write shell scripts grep checks checking that imports do not violate boundary constraints. Add validation script check to the CI pipeline. This enforces slice isolation as a continuous verification gate.
* **Verification:** Script detects and rejects circular or boundary-breaking imports.

---

### S6-SD-25: Audit.md Automation Test Suite (Gap Fix)
* **Priority:** P1
* **Type:** Testing/Verification
* **Assignee:** SD-QA
* **Story Points:** 4
* **Dependencies:** S6-SD-22
* **Description:** Implement a specific automated verification script/runner that executes scenarios mapping directly to the questions checklist in `docs/requirements/audit.md`. This is a verification/testing ticket ensuring audit compliance — every question in audit.md must have a passing automated test.
* **Detailed Steps:**
  1. Create a specialized integration suite under `internal/bootstrap/audit_compliance_test.go`. This is a compliance verification suite, not feature development.
  2. Implement tests mapping to each query in `audit.md`:
     - Test wrong passwords logging attempts (Q: "Did the application correctly detect and respond to the incorrect login details?")
     - Test duplicate registration blocks (Q: "Did the app detect if the email/user is already present in the database?")
     - Test profile visibility lock (Q: "Are you prevented from seeing a non-followed user private profile?")
     - Test chat relationship check blocks (Q: "Can you confirm that it was not possible to create a chat between these two users?")
     - Test WebSocket real-time delivery constraints.
* **Verification:** Running `go test -v ./internal/bootstrap -run=TestAuditCompliance` executes and passes all test scenarios.

---

### S6-SD-31: Full E2E Test Suite
* **Priority:** P0
* **Type:** Testing/Verification
* **Assignee:** SD-QA
* **Story Points:** 8
* **Description:** Implement Playwright automated test scripts checking standard user workflows. This is a verification/testing ticket ensuring full E2E browser-level coverage of all core user flows.
* **Detailed Steps:**
  1. Write E2E browser flows: signup, signin, creating public/private posts, sending follower links, groups creations, messaging. These are comprehensive E2E verification tests, not feature implementation.
* **Verification:** Playwright runner executes successfully.

---

### S6-SD-32: Accessibility (a11y) Audit
* **Priority:** P2
* **Type:** Testing/Verification
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** Check keyboard navigation and screen-reader mappings. This is a verification/testing ticket ensuring accessibility compliance across the frontend.
* **Detailed Steps:**
  1. Ensure interactive buttons contain correct aria tags. Check color contrasts. This is an accessibility audit pass, not a feature build.
* **Verification:** Verify elements are fully keyboard navigable.

---

### S6-SD-33: Frontend Performance Audits
* **Priority:** P2
* **Type:** Testing/Verification
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** Audit bundles sizes and check asset loading. This is a verification/testing ticket ensuring frontend performance meets production standards.
* **Detailed Steps:**
  1. Perform Lighthouse audits. Target scoring >= 90. This is a performance verification pass, not an optimization task.
* **Verification:** Lighthouse report is generated.

---

### S6-SD-34: E2E Audit.md Playwright Suite (Gap Fix)
* **Priority:** P1
* **Type:** Testing/Verification
* **Assignee:** SD-QA
* **Story Points:** 4
* **Dependencies:** S6-SD-31
* **Description:** Implement Playwright browser E2E test scripts specifically mapping to the frontend verification steps listed in `docs/requirements/audit.md`. This is a verification/testing ticket ensuring audit compliance at the E2E browser level — every frontend-relevant question in audit.md must have a passing Playwright test.
* **Detailed Steps:**
  1. Create E2E test file `tests/audit-compliance.spec.ts`. This is a compliance verification suite, not feature development.
  2. Implement tests using Playwright viewport/device settings:
     - Open two browser contexts (e.g. User A on Chrome, User B on Firefox).
     - Test that User A logged in does not impact User B who remains guest (Q: "Can you confirm that both browsers continue with the right users?").
     - Test private profile follow acceptance flows (Q: "Is the user who received the request able to accept or decline the following request?").
     - Verify confirmation popups appear when unfollowing or toggling profile privacy (Q: "If you unfollow a user, do you get a confirmation pop-up?").
* **Verification:** Executing `npx playwright test tests/audit-compliance.spec.ts` passes.

---

## DevOps Tickets

### S6-SD-26: Production Docker Setup
* **Priority:** P1
* **Type:** Scaffolding/Setup
* **Assignee:** SD-QA
* **Story Points:** 5
* **Description:** Audit and rewrite the existing Docker setup (`Dockerfile`, `docker-compose.yml`, etc.) from single-service to two-service architecture per Phase 7 of the architecture plan. This is updating existing infrastructure tooling for the two-service architecture — not creating from scratch, but restructuring what already exists.
* **Detailed Steps:**
   1. Rewrite the existing `docker-compose.yml` to define two separate services: `backend` (port 8080) and `frontend` (port 3000), with persistent volume for SQLite database storage. This updates existing Docker config to reflect the new architecture.
   2. Audit and rewrite/extend `frontend/Dockerfile` for the Next.js frontend (using Node/Bun multi-stage build).
   3. Rewrite the root `Dockerfile` for the Go backend (multi-stage Go build -> minimal alpine image, removing the binary build step for the obsolete `cmd/client/`).
   4. Configure environment variables per arch spec: `DATABASE_DRIVER=sqlite`, `DATABASE_DSN=/app/data/social.db?_journal_mode=WAL&_busy_timeout=5000`, `NEXT_PUBLIC_API_URL=http://backend:8080`.
* **Verification:** Run `docker-compose up` and confirm both services start and connect. Hitting `backend:8080/healthz` returns 200. Hitting `frontend:3000` serves the frontend application.

---

### S6-SD-27: Health Check Endpoints
* **Priority:** P2
* **Type:** Greenfield (New Module/Feature - Health/readiness probes)
* **Assignee:** SD-QA
* **Story Points:** 1
* **Description:** Create probes reporting status. This is a NEW feature with NO existing implementation — health/readiness probe endpoints do not exist yet and must be built from scratch.
* **Detailed Steps:**
  1. Map `GET /healthz` and `GET /readyz` routes checking database and system health. These endpoints are greenfield — no prior health check logic exists in the codebase.
* **Verification:** Hitting endpoints returns 200 OK statuses.

---

### S6-SD-28: Graceful Server Shutdowns
* **Priority:** P1
* **Type:** Greenfield (New Module/Feature - Graceful server shutdown handlers)
* **Assignee:** SD-QA
* **Story Points:** 2
* **Description:** Handle SIGTERM system signals. This is a NEW feature with NO existing implementation — graceful shutdown handlers do not exist yet and must be built from scratch.
* **Detailed Steps:**
  1. Terminate DB pools, drain WS hubs connections safely on exit. This is greenfield shutdown logic — no prior graceful shutdown code exists in the codebase.
* **Verification:** Test log indicates safe exits.

---

### S6-SD-29: Twelve-Factor Configurations Mappings
* **Priority:** P2
* **Type:** Scaffolding/Setup
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** S6-BE-102 (old code must be gone before env var rename)
* **Description:** Restructure config parameters loading strictly from environment variables, aligned with architecture spec env vars. This is updating existing infrastructure configuration for the two-service architecture — restructuring the config module that already exists. **Must run after S6-BE-102** — old code references old env var names.
* **Detailed Steps:**
   1. Update `internal/config/config.go` to load from env vars: `DATABASE_DRIVER`, `DATABASE_DSN`, `SESSION_SECRET`, `PORT`, `CORS_ORIGIN`, `REDIS_URL` (optional), `RABBITMQ_URL` (optional). This updates existing config loading, not a new module.
   2. Remove legacy env var names (`SERVER_HOST`, `CLIENT_HOST`, `SERVER_PORT`).
   3. Update `docker-compose.yml` to pass new env var names (see S6-SD-26).
* **Verification:** Configuration loads parameters correctly from env vars. Old env names produce errors.

---

### S6-SD-30: Docker Smoke Verification Script
* **Priority:** P1
* **Type:** Testing/Verification
* **Assignee:** SD-QA
* **Story Points:** 3
* **Dependencies:** S6-SD-26
* **Description:** Automate startup checks verifying container statuses. This is a verification/testing ticket ensuring the Docker deployment environment starts correctly and services are reachable.
* **Detailed Steps:**
  1. Script that brings up docker containers, checks that `docker ps` returns active states, and runs curl queries. This is a deployment verification pass, not an infrastructure build.
* **Verification:** Automated verification script passes.
