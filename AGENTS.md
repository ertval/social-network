# AGENTS.md

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

## 5. Documentation Conventions

- **Use relative paths** in all plans, proposals, and documentation — never absolute/full file paths.
  - ✅ `internal/user/service.go`, `docs/plan/arch-proposals.md`
  - ❌ `/home/user/code/social-network/internal/user/service.go`
- Reference paths from the project root (where `go.mod` lives).

## 6. Git Branch Naming

All branches must follow the format `<username>/<type>-<detail>`.

- **username**: GitHub/Gitea username (e.g. `ekaramet`, `arnald`)
- **type**: Short category — `feat`, `fix`, `chore`, `refactor`, `docs`, `arch`
- **detail**: kebab-case description of the change

Examples:
- `ekaramet/feat-arch-proposal`
- `arnald/fix-sqlite-busy-timeout`
- `ekaramet/chore-update-deps`

## Audit 2026-06-14: Social Network Codebase Audit

### Patterns
- SQLite DSN must include `_journal_mode=WAL` and `_busy_timeout=5000`
- SQL queries must use `?` placeholders (never `fmt.Sprintf` or `+` concat)
- Password hashing must use `bcrypt` via `golang.org/x/crypto/bcrypt`
- All SQL execution files must split on `";"` not `":"`
- Session/handler lifecycle: `defer` for cleanup, `recover()` in goroutines
- Domain/ never imports infra/ — enforce clean architecture boundary

### Spec Compliance Checklist (every time)
- [ ] Followers table exists? (follows, follow_requests)
- [ ] Profile privacy column? (is_private on users)
- [ ] Post privacy column? (visibility enum on topics)
- [ ] Groups table? (with members, invitations)
- [ ] Events table? (with RSVP)
- [ ] Notification types include follow-request/group-invite/group-join/event-creation
- [ ] Registration form: Email, Password, First Name, Last Name, Date of Birth, Avatar (opt), Nickname (opt), About Me (opt)
- [ ] Chat init checks follow relationship
- [ ] WebSocket upgrade checks auth token
- [ ] docker-compose has 2 services (backend + frontend)
- [ ] Migrations follow numbered up/down format

### Tool Invocations
- `go vet ./...` — catches type errors, shadowing
- `go mod graph` — verify allowed packages
- `golangci-lint run` — full lint (install via `make tools`)
- `govulncheck ./...` — CVE scan
- `go test -race -coverprofile=coverage.out ./...` — full test suite

### Known False Positives
- gorilla/websocket: allowed per spec, don't flag as unauthorized dep
- Custom migration system: allowed per spec ("or other package"), but up/down format expected
- Domain cross-imports (category→topic, topic→comment): acceptable in domain layer (only domain→domain)

### Spec Items Frequently Misread
- "two Docker images" = 2 services in docker-compose, not 2 binaries in 1 image
- "JS framework" — spec lists Next.js/Vue/Svelte/Mithril. Vanilla JS SPA may not qualify
- "Nickname optional" — spec says optional, many impls make it required
- "Date of Birth" — must be date type, not age int
- "Session persistence" — spec requires refresh-token rotation; checking the cookie alone is insufficient
- "Follow private profile" — requires notification (follow-request), not silent creation

### Post-Audit 2026-06-14: Patterns Discovered
- Always check `Scan()` arg list — `ctx` is frequently accidentally passed as a scan target (found in `oauthRepo.go:182`)
- SQLite migration delimiter: use `";"` NEVER `":"` — colon breaks SQL with `'\:'` or timestamps
- WebSocket `CheckOrigin` must NOT be `return true` — this is a CSRF WebSocket vector
- StateManager/Ticker goroutines must have `stop chan struct{}` — no `ctx` = guaranteed leak
- Prepared statements must use `stmt.Exec*` not `db.Exec*` (found `userRepo.go:70-76`)
- `order` (ASC/DESC) concatenation is injection — validate against `["ASC", "DESC"]` whitelist
- Image MIME validation must check magic bytes, not just `Content-Type` header

### Common False Positives to Suppress
- `gorilla/websocket // indirect` annotation: just a `go mod tidy` issue, not a bug
- Migration file naming: spec's "other package" clause can justify custom migration (but up/down and numbered format still expected)
- Google OAuth handler using `UserLoginGithub` service: service is generic via `Provider` interface, so it works — misleading name, not a logic bug
- Unused imports in `infra/services.go`: dead code due to the server wiring through `bootstrap/` directly

### Tool Invocations That Worked Well
- `govulncheck ./...` — caught 28 stdlib CVEs, zero noise
- `go vet ./...` — caught build errors and test signature mismatches
- `grep -rn '<pattern>' --include='*.go' internal/ cmd/ db/` — effective for feature presence/absence checks
- `$(go env GOPATH)/bin/golangci-lint run` — needs v2 for the project's `.golangci.yml` config (v2 format)
- Judge agent with fresh context (no prior Phase 2 context) — caught `CheckOrigin` (JUDGE-FOUND-1) that Phase 2 missed

### Socratic Challenges Worth Documenting
- Missing `_busy_timeout`: fix is trivial one-liner, prevents `database is locked` errors under WAL contention
- `ctx` in `Scan()`: trivial fix, prevents runtime crash — highest priority
- Follow/Group greenfield: high cost but unavoidable for spec compliance — these are core features, not optional
- Migration delimiter `:`→`;`: low risk currently (SQL files happen to lack colons), but fixing prevents future silent data corruption
- WebSocket `CheckOrigin`: disabling origin check is common in development but critical to fix for production deployment
- Remove Age/Add DOB: requires schema migration, handler change, and domain model change — moderate effort, medium priority since existing Age int doesn't crash

## Audit Run 2 2026-06-14: Deeper Scans & Concurrency Audit

### Additional Patterns Discovered
- **RateLimiter Goroutine Leaks**: Tickers must have a `stop` channel. Running `go rl.cleanup()` without a cancellation trigger creates permanent goroutine leaks.
- **WebSocket Panic Recovery**: Spawning long-lived goroutines like `ReadPump` / `WritePump` requires `recover()` block. Standard HTTP handlers are recovered by the net/http server, but user-spawned goroutines will crash the process if they panic.
- **Strict Framework Checks**: A project that uses standard DOM APIs and hand-rolled router instead of an allowed JS framework (e.g. Next.js, Vue, Svelte, Mithril) is a critical spec compliance failure.
- **Client-Side MIME Spoofing**: Validate uploaded images using `http.DetectContentType` on file headers rather than relying solely on the client's `Content-Type` header.

### Tool Invocations That Worked Well
- `~/go/bin/golangci-lint run --config scratch/golangci.yml` works around config v2 schema incompatibilities of older linters.
- `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` compiles cleanly under Go 1.25.1.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, invoke the `skill` tool with `skill: "graphify"` before doing anything else.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
