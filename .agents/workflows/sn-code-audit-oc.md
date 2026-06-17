---
name: sn-code-audit-oc
description: |
  Social Network codebase audit workflow. Four-phase analysis tailored to the
  SN grading specification: allowed-package verification, Server/App/Database
  layer separation, SQLite WAL/pooling rules, follower request flows, post
  privacy scopes (public/almost private/private), group lifecycle with events,
  WebSocket auth handshake and chat authorization, notification triggers, and
  bonus capabilities (OAuth, seeding, popups, Docker). Incorporates 2026
  best practices: Chain-of-Verification, hallucination gates, separated
  Judge/Critic context, and evidence-citation quality gates.
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
6. **Spec-aligned** — every finding should reference the relevant grading requirement from `docs/readme.md`.

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
- `golang.org/x/crypto/bcrypt` — password hashing
- `github.com/gofrs/uuid` or `github.com/google/uuid` — UUID generation
- Authorized migration libraries: `golang-migrate/migrate`, `rubenv/sql-migrate`, `Boostport/migration`

Flag any dependency OUTSIDE this list as HIGH severity — the grading spec restricts allowed packages.

### Step 1.5 — Record baseline
Write raw tool output to `.audit/phase1-baseline.md`. This file is the source of truth that Phase 2 and Phase 3 MUST reference to avoid contradicting reality.

---

## Phase 2: Layered Codebase Analysis (Cognitive)

Analyze the codebase systematically across five layers. Each layer is independent — you MAY dispatch subagents for parallel analysis per layer, provided each subagent gets the Phase 1 baseline as context.

### Layer A — Software Design & Architecture (Clean Architecture / DDD)

1. **Layer separation** — verify the three-layer structure:
   - **Server layer** (`internal/infra/http/`): HTTP handlers, middleware, WebSocket upgrades. Entry points only — no business logic.
   - **App layer** (`internal/app/` or equivalent): use cases, listeners, orchestration. Depends on domain interfaces, not infrastructure.
   - **Database layer** (`internal/infra/storage/`): repository implementations, migrations, queries.
   - Flag any layer-skip violations (e.g., handler calling storage directly without going through app layer).
2. **Domain purity** — `internal/domain/` MUST NOT import `internal/infra/` or any external package beyond standard library. Flag violations as HIGH.
3. **Startup migrations** — verify that database migrations are applied automatically at application startup. Check `main.go`, bootstrap code, or an `init()` function for migration execution. Flag if migrations require manual invocation.
4. **Project structure** — verify migration files follow `db/migrations/sqlite/*.up.sql` / `*.down.sql` naming or an equivalent structured convention.

### Layer B — Allowed Packages & Dependency Audit

1. **Import scan** — grep all `.go` files for external imports. Verify every non-standard import falls in the allowed list from Phase 1.4.
2. **Migration library check** — confirm the chosen migration library is one of the three authorized options. Flag any custom or disallowed migration system.
3. **UUID library check** — confirm `gofrs/uuid` or `google/uuid` is used. Flag other UUID libraries.

### Layer C — Idiomatic Go Correctness

1. **Error handling** — verify `%w` wrapping on all error returns from `fmt.Errorf`. Flag silent discards. Ensure `defer` + `recover` in every goroutine.
2. **Context propagation** — trace `context.Context` from HTTP handlers through service calls to database queries. Missing propagation is a bug.
3. **Concurrency safety** — inspect `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, channel operations. Look for:
   - Mutexes not released on all return paths (unlock via `defer`)
   - Channel sends without corresponding receives (goroutine leaks)
   - `sync.Map` vs `map`+mutex tradeoffs in hot paths
4. **Resource lifecycle** — confirm `defer` closes files, HTTP response bodies, database rows, and network connections.

### Layer D — Security & Functional Specification

This layer is the core of the SN audit. Every requirement below maps to the grading spec.

#### D1. Registration & Authentication
1. **Registration form** — verify handler accepts: Email, Password, First Name, Last Name, Date of Birth (required), plus Avatar/Image, Nickname, About Me (optional).
2. **Password hashing** — confirm passwords are hashed with `bcrypt` (cost >= 10). Flag plaintext or weak hashing.
3. **Session cookies** — review cookie config for `HttpOnly`, `Secure`, `SameSite`, `Expires`/`MaxAge`. Missing hardening flags are HIGH severity.
4. **OAuth (bonus)** — check if OAuth (GitHub / OAuthenticator) integration exists. Tag as BONUS-PRESENT or BONUS-MISSING without severity.

#### D2. SQL Injection & SQLite Configuration
1. **Parameter binding** — inspect EVERY query in storage packages. Ensure `?` or `$N` placeholders, not `fmt.Sprintf` or `+` concatenation.
2. **WAL mode** — verify SQLite DSN includes `_journal_mode=WAL`. Flag if absent.
3. **Busy timeout** — verify DSN includes `_busy_timeout=5000` (or equivalent). Flag if absent.
4. **Connection pooling** — verify `SetMaxOpenConns` (1-10), `SetMaxIdleConns`, and `SetConnMaxLifetime` are configured. Flag if unlimited.

#### D3. Profile Privacy & Follower Flows
1. **Public/private toggle** — verify users can switch profile between public and private.
2. **Auto-follow on public** — verify that following a public profile auto-follows (no request needed).
3. **Follow request flow** — verify private profiles generate follow requests that the target can accept or decline.
4. **Access control** — verify non-followers are blocked from viewing private profile content while followers have access.
5. **Confirmation popups (bonus)** — check for confirmation dialogs on actions like unfollowing or privacy toggling.

#### D4. Posts & Comments
1. **Privacy scopes** — verify posts support three visibility levels:
   - `public`: visible to all logged-in users
   - `almost private` (followers only): visible only to followers
   - `private` (selected followers): visible only to specifically selected followers
2. **Media attachments** — verify posts/comments can attach images (JPG, PNG) or GIFs.
3. **Content access enforcement** — verify the backend enforces privacy scope at query time (not just in the UI).

#### D5. Groups & Events
1. **Group browse** — verify group discovery / listing endpoint exists.
2. **Join requests** — verify users can request to join a group; creator must approve.
3. **Invitations** — verify followers can be invited to a group; they can accept or decline.
4. **Group posts visibility** — verify group posts/comments are visible ONLY to group members.
5. **Group chat rooms** — verify each group has an isolated chat room; only members can send/receive.
6. **Group events** — verify events have: Title, Description, Day/Time, "Going" / "Not going" RSVP options.

#### D6. WebSocket Chat Security
1. **Handshake token verification** — verify the WebSocket upgrade handler checks authentication tokens BEFORE completing the handshake.
2. **Chat authorization** — verify chat creation is only allowed when at least one user follows the other.
3. **Group chat authorization** — verify only group members can send/receive in group chat rooms.
4. **Read limits** — verify `conn.SetReadLimit` is set to block oversized messages.
5. **Deadlines** — verify `SetReadDeadline` / `SetWriteDeadline` are configured to prevent dead connections.
6. **Emoji support** — verify emoji characters are handled in messages.

#### D7. Notifications Engine
Verify notification triggers exist for ALL of:
1. Follow request received (private profile)
2. Group invitation received by user
3. Group join request received (sent to group creator)
4. New event created in a group

### Layer E — Performance

1. **N+1 queries** — scan handler/service loops making DB calls per iteration. Flag missing JOINs.
2. **Connection pooling** — verify SQLite `SetMaxOpenConns` is set conservatively (1-10) to avoid `database is locked` errors.
3. **Goroutine leaks** — look for goroutines in WebSocket handlers, notification pushers, or event listeners without lifecycle management.
4. **SQLite write contention** — if WAL mode is enabled, verify concurrent write patterns don't overwhelm SQLite's single-writer limitation.

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
- WebSocket handshake security that Phase 2 may have missed
- SQLite connection string parameters not checked in Phase 2
- Notification triggers that may be missing

### Step 3.3 — Mitigation check
For each severity-2+ finding, check if the issue is mitigated elsewhere (middleware, wrapper, validation layer). If mitigated, reduce severity or drop.

### Step 3.4 — Socratic challenge
For every HIGH/CRITICAL finding, ask: "Is the proposed fix worth the complexity? Does it introduce new failure modes? Is there a simpler approach?" Downgrade findings where the trade-off is net-negative.

### Step 3.5 — Spec compliance double-check
Re-verify each finding against the grading requirements in `docs/readme.md`. Findings that misread the spec should be DROPPED with explanation.

---

## Phase 4: Aggregation & Synthesis

Produce a structured report at `docs/audit/codebase_audit_report.md`.

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
| Requirement | Status | Location | Notes |
|---|---|---|---|
| Allowed packages only | ✅ / ❌ | file:line | |
| WAL mode + busy timeout | ✅ / ❌ | file:line | |
| Follower request flow | ✅ / ❌ | file:line | |
| Post privacy scopes | ✅ / ❌ | file:line | |
| ... | | | |

## Critical & High Findings
Each entry:
- **ID**: AUDIT-001
- **Severity**: CRITICAL
- **Location**: `path/to/file.go:42`
- **Spec ref**: section in `docs/readme.md`
- **Observation**: concise description of the issue
- **Evidence**: excerpt from code or tool output
- **Risk**: what an attacker or user would experience
- **Remediation**: exact code change (diff block or snippet)
- **Judge Verdict**: CONFIRMED / WEAKENED / JUDGE-FOUND

## Medium & Low Findings
- Maintainability suggestions, architectural observations, style deviations
- Each MUST still cite `file:line`

## Bonus Feature Inventory
| Feature | Status | Location |
|---|---|---|
| OAuth (GitHub) | ✅ PRESENT / ❌ MISSING | file:line |
| DB seeding | ✅ PRESENT / ❌ MISSING | file:line |
| Confirmation popups | ✅ PRESENT / ❌ MISSING | file:line |
| Docker / build scripts | ✅ PRESENT / ❌ MISSING | file:line |

## Verification Plan
- Commands to run to verify each fix
- Test suites to execute
- Manual testing steps for WebSocket, notifications, and privacy flows
```

### Phase 4 quality gates
- Every CRITICAL finding MUST be CONFIRMED by the Judge
- No finding appears without a `file:line` citation
- Remediation blocks MUST be syntactically valid Go (parse before writing)
- Spec compliance matrix MUST cover every grading requirement from `docs/readme.md`

---

## Post-Audit: AGENTS.md Update

After the report is written, append a summary entry to `AGENTS.md` (or `CLAUDE.md`) capturing:
- Patterns discovered during this audit (e.g., "always check SQLite DSN parameters")
- Common false positives to suppress next run
- Tool invocations that worked well
- Spec items that are frequently misread (to narrow prompts in future runs)
