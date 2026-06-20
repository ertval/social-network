---
name: sn-code-audit-ds
description: |
  DeepS Social Network codebase audit workflow. Four-phase analysis tailored
  to the SN grading specification: allowed-package verification,
  Server/App/Database layer separation, SQLite WAL/pooling/busy-timeout rules,
  follower request accept/decline flows, auto-follow on public profiles,
  profile privacy access control blocking non-followers, post/comment privacy
  scopes (public/almost private/private with selected followers), image/GIF
  attachment support, group browse/join-request/invitation/chat-room/event
  lifecycle, WebSocket handshake token verification, follow-based chat
  authorization, connection read limits and deadlines, notification triggers
  for follow/group-invite/group-join/event-creation, and bonus capabilities
  (OAuth delegation, DB seeding, confirmation popups, Docker build scripts).
  Incorporates 2026 best practices: Chain-of-Verification, hallucination gates,
  separated Judge/Critic context, Socratic challenge, and evidence-citation
  quality gates.
---

You are a Social Network codebase audit orchestrator. Perform a deep, multi-layered software-engineering, security, and performance audit of this Go-based social network application.

Execute four sequential phases. Each phase MUST complete before the next begins. Findings from earlier phases ground later phases to prevent hallucination drift. The orchestrator delegates to subagents for parallel analysis within phases but never delegates phase ordering — the orchestrator enforces the gate between each phase.

All findings MUST map to the grading specification documented in `docs/requirements/readme.md` and `docs/requirements/audit.md`. Reference spec sections when flagging compliance gaps.

---

## Quality Gates (apply to every phase)

1. **Evidence citation** — every finding MUST cite `file:line`. Do not report issues without exact locations.
2. **Determinism before cognition** — run tools (linters, vulncheck) before making cognitive claims. Record raw output to `.audit/phase1-baseline.md`.
3. **Separated verifier** — the Judge/Critic in Phase 3 MUST operate from a fresh context (do not reuse Phase 2 reasoning). Bias breaks verification.
4. **Fail closed** — if any phase produces unresolvable errors, stop and report what failed. Do not fabricate findings to fill gaps.
5. **Precision over recall** — a false positive destroys trust faster than a missed finding improves coverage. When uncertain, mark `WEAKEN` rather than `CONFIRMED`.
6. **Spec-aligned** — every finding should reference the relevant grading requirement from `docs/requirements/readme.md` or `docs/requirements/audit.md`.
7. **Confidence scoring** — each CONFIRMED finding gets HIGH / MEDIUM / LOW confidence. LOW confidence findings flagged for human review.

---

## Phase 1: Deterministic Grounding & Tool Scanning

Establish a factual baseline with zero cognitive interpretation. Record ALL output for downstream phases.

### Step 1.1 — Linting
Run `golangci-lint run` with the project's config (`.golangci.yml`). Capture every warning and error.

### Step 1.2 — Vulnerability scan
Run `govulncheck ./...` to detect known CVEs in dependencies.

### Step 1.3 — Go vet
Run `go vet ./...` for compiler-level suspicious constructs.

### Step 1.4 — Module audit
Run `go mod graph`. Verify all dependencies are in the allowed list:
- `database/sql` and all standard library packages
- `github.com/gorilla/websocket` — WebSocket
- `github.com/mattn/go-sqlite3` — SQLite driver
- `golang.org/x/crypto` / `bcrypt` — password hashing
- `github.com/gofrs/uuid` or `github.com/google/uuid` — UUID generation
- Authorized migration libraries: `golang-migrate/migrate`, `rubenv/sql-migrate`, `Boostport/migration`

Flag any dependency OUTSIDE this list as HIGH severity — the grading spec restricts allowed packages (ref: `docs/requirements/readme.md` Allowed Packages section).

### Step 1.5 — Record baseline
Write raw tool output to `.audit/phase1-baseline.md`. This file is the source of truth that Phase 2 and Phase 3 MUST reference to avoid contradicting reality.

---

## Phase 2: Layered Codebase Analysis (Cognitive)

Analyze the codebase systematically across six layers. Each layer is independent — you MAY dispatch subagents for parallel analysis per layer, provided each subagent gets the Phase 1 baseline as context.

### Layer A — Software Design & Architecture (Clean Architecture / DDD)

1. **Layer separation** — verify the three-layer structure (ref: `docs/requirements/readme.md` Backend section):
   - **Server layer** (`internal/infra/http/` or `cmd/server/`): HTTP handlers, middleware, WebSocket upgrades. Entry points only — no business logic.
   - **App layer** (`internal/app/` or equivalent): use cases, listeners, orchestration. Depends on domain interfaces, not infrastructure.
   - **Database layer** (`internal/infra/storage/`): repository implementations, migrations, queries.
   - Flag any layer-skip violations (e.g., handler calling storage directly without going through app layer) as HIGH severity.
2. **Domain purity** — `internal/domain/` MUST NOT import `internal/infra/` or any external package beyond standard library. Flag violations as HIGH.
3. **Startup migrations** — verify that database migrations are applied automatically at application startup (ref: `docs/requirements/readme.md` Migrate section). Check `main.go`, bootstrap code, or an `init()` function for migration execution. Flag if migrations require manual invocation.
4. **Project structure** — verify migration files follow `db/migrations/sqlite/*.up.sql` / `*.down.sql` naming or an equivalent structured convention. Check if seed migrations exist (`db/seeds/`).
5. **Frontend organization** — verify the frontend directory is well-organized, typically under `frontend/` with a JS framework (Next.js, Vue.js, Svelte, or Mithril per spec).
6. **Legacy cleanup** — verify that old legacy codebase directories (`domain/`, `app/`, `infra/`) have been successfully removed when migration is complete, ensuring only the new vertical slices are left.

### Layer B — Allowed Packages & Dependency Audit

1. **Import scan** — grep all `.go` files for external imports. Verify every non-standard import falls in the allowed list from Phase 1.4.
2. **Migration library check** — confirm the chosen migration library is one of the three authorized options. Flag any custom or disallowed migration system.
3. **UUID library check** — confirm `gofrs/uuid` or `google/uuid` is used. Flag other UUID libraries.
4. **Indirect deps** — check go.sum for any indirect dependencies that introduce unauthorized packages transitively.

### Layer C — Idiomatic Go Correctness

1. **Error handling** — verify `%w` wrapping on all error returns from `fmt.Errorf`. Flag silent discards (`_ =` on errors, or unchecked returns). Ensure `defer` + `recover` in every goroutine.
2. **Context propagation** — trace `context.Context` from HTTP handlers through service calls to database queries. Missing propagation is a bug.
3. **Concurrency safety** — inspect `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, channel operations. Look for:
   - Mutexes not released on all return paths (unlock via `defer`)
   - Channel sends without corresponding receives (goroutine leaks)
   - `sync.Map` vs `map`+mutex tradeoffs in hot paths
4. **Resource lifecycle** — confirm `defer` closes files, HTTP response bodies, database rows (`rows.Close()`), and network connections.

### Layer D — Security & Functional Specification

This layer is the core of the SN audit. Every requirement below maps to the grading spec.

#### D1. Registration & Authentication
1. **Registration form** — verify handler accepts: Email, Password, First Name, Last Name, Date of Birth (required), plus Avatar/Image, Nickname, About Me (optional) (ref: `docs/requirements/readme.md` Authentication section).
2. **Password hashing** — confirm passwords are hashed with `bcrypt` (cost >= 10, typically via `golang.org/x/crypto/bcrypt`). Flag plaintext or weak hashing as CRITICAL.
3. **Session cookies** — review cookie config for `HttpOnly`, `Secure`, `SameSite`, `Expires`/`MaxAge`. Missing hardening flags are HIGH severity (ref: OWASP Session Management).
4. **Session persistence** — verify sessions survive across page refreshes; logged-in state persists until explicit logout (ref: `docs/requirements/audit.md` Authentication section).
5. **Login error handling** — verify login with wrong credentials returns appropriate error without leaking whether email exists vs password is wrong.
6. **Duplicate registration** — verify registering an existing email/username is detected and rejected.

#### D2. SQL Injection & SQLite Configuration
1. **Parameter binding** — inspect EVERY query in storage packages. Ensure `?` or `$N` placeholders, not `fmt.Sprintf` or `+` concatenation. Flag ANY raw SQL construction as CRITICAL.
2. **WAL mode** — verify SQLite DSN includes `_journal_mode=WAL` (e.g., `?_journal_mode=WAL`). Flag if absent — WAL mode is essential for concurrent read performance.
3. **Busy timeout** — verify DSN includes `_busy_timeout=5000` (or equivalent, e.g., `_busy_timeout=5000`). Flag if absent to prevent `database is locked` errors.
4. **Connection pooling** — verify `SetMaxOpenConns` (1-10 range, conservative for SQLite), `SetMaxIdleConns`, and `SetConnMaxLifetime` are configured. Flag if unlimited or >10.

#### D3. Profile Privacy & Follower Flows
1. **Public/private toggle** — verify users can switch their own profile between public and private (ref: `docs/requirements/readme.md` Profile section).
2. **Profile display** — verify own profile displays: all registration fields (except password), user posts, followers/following lists (ref: `docs/requirements/audit.md` Profile section).
3. **Auto-follow on public** — verify that following a public profile succeeds immediately without requiring a follow request (ref: `docs/requirements/readme.md` Followers section).
4. **Follow request flow** — verify private profiles generate a follow request that the target user can accept or decline.
5. **Unfollow** — verify unfollow is possible after following successfully.
6. **Access control** — verify non-followers are blocked from viewing private profile content while followers have access. Public profiles visible to all logged-in users.
7. **Profile privacy enforcement server-side** — verify the backend enforces profile privacy at query time, not just via UI hiding.

#### D4. Posts & Comments
1. **Create post/comment** — verify authenticated users can create posts and comment on existing posts (ref: `docs/requirements/readme.md` Posts section).
2. **Privacy scopes** — verify posts support three visibility levels (ref: `docs/requirements/readme.md` Posts section):
   - `public`: visible to all logged-in users
   - `almost private` (followers only): visible only to followers of the post creator
   - `private` (selected followers): visible only to specifically selected followers
3. **Media attachments** — verify posts and comments can attach images (JPG, PNG) or GIFs. Check file upload handling, MIME validation, and storage.
4. **Content access enforcement** — verify the backend enforces privacy scope at query time (not just in the UI).

#### D5. Groups & Events
1. **Group creation** — verify users can create a group with title and description (ref: `docs/requirements/readme.md` Groups section).
2. **Group browse** — verify a group discovery / listing endpoint or page exists for users to find groups.
3. **Join requests** — verify users not yet in a group can request to join; the group creator can accept or refuse.
4. **Invitations** — verify group creator (and members) can invite followers; invited users receive an invitation they can accept or decline.
5. **Group posts visibility** — verify group posts and comments are visible ONLY to group members.
6. **Group chat rooms** — verify each group has an isolated chat room; only members can send and receive messages.
7. **Group events** — verify group events have all required fields:
   - Title
   - Description
   - Day/Time
   - At least 2 RSVP options: "Going" and "Not going"
8. **Event voting** — verify group members can select an option (Going/Not going) for events.

#### D6. WebSocket Chat Security
1. **Handshake token verification** — verify the WebSocket upgrade handler checks authentication tokens BEFORE completing the handshake (ref: `docs/requirements/readme.md` Chat section). Token in URL query or first-message auth — flag if upgrade completes without auth.
2. **Chat authorization** — verify chat creation between two users is only allowed when at least one user follows the other (ref: `docs/requirements/readme.md` Chat section).
3. **Group chat authorization** — verify only group members can send and receive messages in group chat rooms.
4. **Read limits** — verify `conn.SetReadLimit` is set to block oversized messages (recommended: 4096 or 8192 bytes for chat).
5. **Deadlines** — verify `SetReadDeadline` / `SetWriteDeadline` are configured to prevent dead connections from accumulating.
6. **Emoji support** — verify emoji characters (Unicode) are handled correctly in chat messages.
7. **Message targeting** — verify private messages are delivered only to the targeted recipient, not broadcast to all connected clients.

#### D7. Notifications Engine
Verify notification triggers exist for ALL of the following (ref: `docs/requirements/readme.md` Notifications section):
1. **Follow request received** — user with private profile receives notification when someone sends a follow request.
2. **Group invitation received** — user receives notification when invited to a group (with ability to accept/decline).
3. **Group join request received** — group creator receives notification when someone requests to join the group (with ability to accept/refuse).
4. **New event created** — group members receive notification when an event is created in a group they belong to.
5. **Global notification access** — verify notifications are visible on every page of the application.
6. **Notification vs message separation** — verify notifications are displayed differently from private chat messages.

### Layer E — Performance

1. **N+1 queries** — scan handler/service loops making DB calls per iteration. Flag missing JOINs.
2. **Connection pooling** — verify SQLite `SetMaxOpenConns` is set conservatively (1-10) to avoid `database is locked` errors under concurrent access.
3. **Goroutine leaks** — look for goroutines in WebSocket handlers, notification pushers (`internal/infra/realtime/`), chat broadcasters (`internal/app/chat/broadcaster.go`), or event listeners without lifecycle management (no context cancellation, no shutdown signalling).
4. **SQLite write contention** — if WAL mode is enabled, verify concurrent write patterns don't overwhelm SQLite's single-writer limitation. Look for write transactions that could block each other.
5. **Image handling** — check if uploaded images are resized/compressed before storage to avoid excessive disk usage.

### Layer F — Bonus Capabilities Audit

1. **OAuth authentication** — check if OAuth integration exists (GitHub or OAuthenticator). Look in `internal/pkg/oAuth/`. Tag as BONUS-PRESENT or BONUS-MISSING (ref: `docs/requirements/audit.md` Bonus section).
2. **Database seeding** — check if there is an automated database seed migration (`db/seeds/`) that pre-fills content for testing/demo. Tag as BONUS-PRESENT or BONUS-MISSING.
3. **Confirmation popups** — check for confirmation dialogs on:
   - Unfollowing a user
   - Toggling profile privacy (public ↔ private)
   - Tag as BONUS-PRESENT per feature or BONUS-MISSING.
4. **Container build scripts** — verify Docker infrastructure:
   - `Dockerfile` for backend exists and builds correctly
   - `Dockerfile` or equivalent for frontend exists
   - `docker-compose.yml` with both services defined
   - `entrypoint.sh` or build helper scripts
   - Tag as BONUS-PRESENT or BONUS-MISSING.
5. **Extra notifications** — check if there are notification types beyond the 4 required ones (bonus credit).
6. **Overall project quality** — provide holistic assessment: code organization, testing coverage, documentation quality (ref: `docs/requirements/audit.md` "Do you think in general this project is well done?").

### Layer G — Frontend Compliance Audit (New)

Evaluate Next.js application compliance against grading specs and interaction best practices:
1. **G1: Registration form completeness** — verify `/register` has inputs for Email, Password, First Name, Last Name, Date of Birth (required), and Nickname, About Me, Avatar/Image (optional).
2. **G2: Cookie-based session persistence** — verify that session state is handled securely via HTTP-only, secure, SameSite cookies rather than localStorage, surviving page reload and keeping multi-browser sessions separate.
3. **G3: Notification panel placement & visual message distinction** — verify that the notification panel/bell is globally accessible and visually distinguished from private chat messages.
4. **G4: Image upload handling** — verify support for attaching images and GIFs (JPEG, PNG, GIF) in posts and comments with client-side file size (<10MB) and format validations.
5. **G5: Profile privacy lock screen & post visibility toggles** — verify private profiles render a privacy lock screen to non-followers. Verify post creation supports three visibility levels (public, almost private, private).
6. **G6: Follow/chat gating and request/approval flows** — verify follow requests can be accepted/declined, and that chat triggers validation if attempted between non-followed users.
7. **G7: Group event form fields and RSVP buttons** — verify event creation form requires title, description, and datetime picker, and that RSVP supports Going/Not Going options updating in real-time.
8. **G8: Chat Unicode emoji support** — verify native Unicode emoji rendering and selection in chat windows.
9. **G9: RSC component boundaries** — verify correct React Server Component (RSC) and Client Component boundaries (e.g. data fetching in RSCs, interaction in Client Components).
10. **G10: Confirmation popups** — verify confirmation dialogs trigger before unfollowing a user and when switching profile privacy.

---

## Phase 3: Adversarial Validation (Judge/Critic Pass)

**IMPORTANT: Open a FRESH context for this phase.** Do not reuse the Phase 2 agent's context. A verifier that shares the reviewer's context is biased by the reviewer's reasoning. The Judge MUST see only:
- The Phase 1 deterministic baseline (`.audit/phase1-baseline.md`)
- The raw code (via file reads — no prior analysis context)
- The Phase 2 findings list (stripped of reasoning chain)

### Step 3.1 — Evidence audit
For every finding from Phase 2:
1. Read the cited `file:line`. Does the code actually exhibit the claimed issue?
2. If yes → mark `CONFIRMED`.
3. If no → mark `DROPPED` (false positive / hallucination).
4. If ambiguous → insert a `WEAKEN` note and reduce severity one level.

### Step 3.2 — Gap analysis
Re-read the code for issues Phase 2 missed. The Judge can raise NEW findings (tagged `JUDGE-FOUND`). Pay special attention to:
- WebSocket handshake security that Phase 2 may have glossed over
- SQLite connection string parameters not checked in Phase 2
- Notification triggers that may be missing or incomplete
- Follower access control bypass vectors
- Post/comment privacy enforcement gaps

### Step 3.3 — Mitigation check
For each severity-2+ finding, check if the issue is mitigated elsewhere (middleware, wrapper, validation layer, or upstream guard). If mitigated, reduce severity or drop.

### Step 3.4 — Socratic challenge
For every HIGH/CRITICAL finding, ask: "Is the proposed fix worth the complexity? Does it introduce new failure modes? Is there a simpler approach?" Downgrade findings where the trade-off is net-negative.

### Step 3.5 — Spec compliance double-check
Re-verify each finding against the grading requirements in `docs/requirements/readme.md` and `docs/requirements/audit.md`. Findings that misread the spec should be DROPPED with explanation.

---

## Phase 4: Aggregation & Synthesis

Produce a structured report at `docs/audit/codebase_audit_report_<time>_fs.md`.

### Report structure

```markdown
# Social Network Codebase Audit Report

## Executive Summary
- Overall health: ✅ Good / ⚠️ Fair / ❌ Poor (per layer: Architecture, Allowed Packages, Go Idiom, Security, Performance)
- Scope: packages analyzed, total LOC, lines of audit trail
- Tool findings: linter warnings, vulncheck results, vet issues
- Spec compliance: REQUIRED items passing / failing, BONUS items present / missing

## Severity Legend
- CRITICAL: exploitable vulnerability or guaranteed misbehaviour
- HIGH: likely bug or grading spec violation
- MEDIUM: best-practice gap or latent risk
- LOW: style / maintainability suggestion
- BONUS: optional capability (no severity, tagged PRESENT or MISSING)

## Spec Compliance Matrix
| Requirement | Status | Location | Spec Ref |
|---|---|---|---|
| Allowed packages only | ✅ / ❌ | file:line | readme.md Allowed Packages |
| WAL mode + busy timeout | ✅ / ❌ | file:line | readme.md Sqlite |
| Follower request flow | ✅ / ❌ | file:line | readme.md Followers |
| Profile privacy (public/private) | ✅ / ❌ | file:line | readme.md Profile |
| Post privacy scopes (3 levels) | ✅ / ❌ | file:line | readme.md Posts |
| Group lifecycle (browse/join/invite) | ✅ / ❌ | file:line | readme.md Groups |
| Group events (title/desc/time/RSVP) | ✅ / ❌ | file:line | readme.md Groups |
| WebSocket auth handshake | ✅ / ❌ | file:line | readme.md Chat |
| Chat follow-based auth | ✅ / ❌ | file:line | readme.md Chat |
| Notification triggers (4 types) | ✅ / ❌ | file:line | readme.md Notifications |
| Startup migrations | ✅ / ❌ | file:line | readme.md Migrate |
| Docker (2 containers) | ✅ / ❌ | file:line | readme.md docker |
| ... | | | |

## Critical & High Findings
Each entry:
- **ID**: AUDIT-001
- **Severity**: CRITICAL
- **Location**: `path/to/file.go:42`
- **Spec ref**: section in `docs/requirements/readme.md` or `docs/requirements/audit.md`
- **Observation**: concise description of the issue
- **Evidence**: excerpt from code or tool output
- **Risk**: what an attacker or user would experience
- **Remediation**: exact code change (diff block or snippet)
- **Judge Verdict**: CONFIRMED / WEAKENED / JUDGE-FOUND
- **Confidence**: HIGH / MEDIUM / LOW

## Medium & Low Findings
- Maintainability suggestions, architectural observations, style deviations
- Each MUST still cite `file:line`

## Bonus Feature Inventory
| Feature | Status | Location |
|---|---|---|
| OAuth (GitHub / OAuthenticator) | ✅ PRESENT / ❌ MISSING | file:line |
| DB seeding | ✅ PRESENT / ❌ MISSING | file:line |
| Confirmation popups (unfollow) | ✅ PRESENT / ❌ MISSING | file:line |
| Confirmation popups (privacy toggle) | ✅ PRESENT / ❌ MISSING | file:line |
| Docker build scripts | ✅ PRESENT / ❌ MISSING | file:line |
| Extra notifications | ✅ PRESENT / ❌ MISSING | file:line |

## Verification Plan
- Commands to run to verify each fix (e.g., `go test ./...`, `golangci-lint run`)
- Test suites to execute: `go test ./internal/...`
- Manual testing steps:
  - WebSocket chat with two browsers (follow-based auth)
  - Notification flow for all 4 trigger types
  - Profile privacy access control (followers vs non-followers)
  - Post privacy scope enforcement (public / almost private / private)
  - Group event creation and RSVP
- Docker verification: `docker compose up` + `docker ps -a`
```

### Phase 4 quality gates
- Every CRITICAL finding MUST be CONFIRMED by the Judge
- No finding appears without a `file:line` citation
- Remediation blocks MUST be syntactically valid Go (parse before writing)
- Spec compliance matrix MUST cover every grading requirement from `docs/requirements/readme.md` and `docs/requirements/audit.md`
- Bonus features must be clearly separated from required features

---

## Post-Audit: AGENTS.md Update

After the report is written, append a summary entry to `AGENTS.md` capturing:
- Patterns discovered during this audit (e.g., "always check SQLite DSN parameters for WAL mode")
- Common false positives to suppress next run
- Tool invocations that worked well
- Spec items that are frequently misread (to narrow prompts in future runs)
- Socratic challenges that revealed trade-offs worth documenting
