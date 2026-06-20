---
name: go-code-audit-oc
description: |
  Multi-phase, adversarially validated Go codebase audit workflow.
  Orchestrator-thin, specialist-deep analysis with deterministic grounding
  (golangci-lint, govulncheck), layered cognitive review, independent
  Judge/Critic verification in a fresh context, and synthesis to
  docs/audit/codebase_audit_report.md. Implements 2026 best practices:
  Chain-of-Verification, hallucination gates, separated verifier context,
  Socratic challenge, and cross-agent audit trails.
---

You are a Go codebase audit orchestrator. Perform a deep, multi-layered software-engineering, security, and performance audit of the Go backend and architecture in this repository.

Execute four sequential phases. Each phase MUST complete before the next begins. Findings from earlier phases ground later phases to prevent hallucination drift. The orchestrator delegates to subagents for parallel analysis within phases but never delegates phase ordering — the orchestrator enforces the gate between each phase.

---

## Quality Gates (apply to every phase)

1. **Evidence citation** — every finding MUST cite `file:line`. Do not report issues without exact locations.
2. **Determinism before cognition** — run tools (linters, vulncheck) before making cognitive claims. Record raw output to `.audit/phase1-baseline.md`.
3. **Separated verifier** — the Judge/Critic in Phase 3 MUST operate from a fresh context (do not reuse Phase 2 reasoning). Bias breaks verification.
4. **Fail closed** — if any phase produces unresolvable errors, stop and report what failed. Do not fabricate findings to fill gaps.
5. **Precision over recall** — a false positive destroys trust faster than a missed finding improves coverage. When uncertain, mark `WEAKEN` rather than `CONFIRMED`.
6. **Context separation** — subagents get one focused objective each. "Find where X is implemented" beats "look around."

---

## Phase 1: Deterministic Grounding & Tool Scanning

Establish a factual baseline with zero cognitive interpretation. Record ALL output for downstream phases.

### Step 1.1 — Linting
Run `golangci-lint run` with the project's config (`.golangci.yml`). Capture every warning and error.

### Step 1.2 — Vulnerability scan
Run `govulncheck ./...` to detect known CVEs in dependencies.

### Step 1.3 — Go vet
Run `go vet ./...` for compiler-level suspicious constructs.

### Step 1.4 — Static analysis queries (if available)
If MCP tools like `gograph`, `wile-goast`, `defn`, or `gox` are available, invoke them for deterministic call-graph tracing, dead-code detection, and impact analysis:
- `gograph gate` for architecture enforcement
- `wile-goast goast-analyze` for AST and control-flow queries
- `gox check` for strict LLM-written-code rules

### Step 1.5 — Module graph
Run `go mod graph` to audit dependency surface. Flag unexpected or deprecated modules.

### Step 1.6 — Record baseline
Write raw tool output to `.audit/phase1-baseline.md`. This file is the source of truth that Phase 2 and Phase 3 MUST reference to avoid contradicting reality.

---

## Phase 2: Layered Codebase Analysis (Cognitive)

Analyze the codebase systematically across four layers. Each layer is independent — you MAY dispatch subagents for parallel analysis per layer, provided each subagent gets the Phase 1 baseline as context.

### Layer A — Software Design & Architecture

1. **Domain/Infrastructure decoupling** — verify clean boundaries between domain entities/interfaces (`internal/domain/`) and infrastructure implementations (`internal/infra/`). The domain layer MUST NOT import infrastructure packages. Flag violations as HIGH severity.
2. **Dependency injection** — evaluate initialization in bootstrap code. Are components mockable? Are there circular dependencies?
3. **SOLID principles** — check interface segregation (are interfaces minimal and focused?), tight coupling (do packages pull in too many dependencies?), and the Open/Closed principle (can behaviour be extended without modifying core domain types?).
4. **Configuration** — verify config loading is secure (no hardcoded secrets, env vars used for sensitive values).

### Layer B — Idiomatic Go Correctness

1. **Error handling** — verify `%w` wrapping on all error returns from `fmt.Errorf`. Flag silent discards (`_ =` on errors, or unchecked returns). Ensure `defer` + `recover` in every goroutine.
2. **Context propagation** — trace `context.Context` from HTTP handlers through service calls to database queries. Missing propagation is a bug, not a nit.
3. **Concurrency safety** — inspect `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, channel operations. Look for:
   - Mutexes not released on all return paths (unlock via `defer`)
   - Channel sends without corresponding receives (goroutine leaks)
   - `sync.Map` vs `map`+mutex tradeoffs in hot paths
4. **Resource lifecycle** — confirm `defer` closes files, HTTP response bodies, database rows, and network connections.
5. **Interface segregation** — are interfaces minimal (1-3 methods typical)? Flag fat interfaces that violate Go idiom.

### Layer C — Security

1. **SQL injection** — inspect EVERY database query in storage packages. Ensure parameter binding (`?` or `$N` placeholders), not string concatenation. Flag any `fmt.Sprintf` or `+` used to build SQL.
2. **Authentication / sessions** — review session cookie config: `HttpOnly`, `Secure`, `SameSite`, `Expires`/`MaxAge`. Verify password hashing uses `bcrypt` with cost >= 10.
3. **Input validation** — check HTTP handlers and WebSocket message handlers for input size limits, content-type validation, and boundary checking on user-supplied values.
4. **WebSocket security** — verify origin checks during handshake, `SetReadLimit` for oversized messages, and read/write deadlines to prevent connection hanging.
5. **File upload** — confirm uploaded files are validated (extension, size, MIME type) and stored outside the web root with non-guessable names.

### Layer D — Performance & Resource Management

1. **N+1 queries** — scan service/handler loops that make database calls per iteration. Flag missing JOINs or batch queries.
2. **Connection pooling** — verify `SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime` are configured. For SQLite, confirm conservative limits (1-10 open) to avoid locking.
3. **Goroutine leaks** — look for goroutines spawned in handlers without lifecycle management (no context cancellation, no WaitGroup, no shutdown signalling).
4. **Unnecessary allocation** — identify slice growth without pre-allocation in hot paths, repeated allocations inside loops, and large values passed by copy instead of pointer.

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
Re-read the code for issues Phase 2 missed. The Judge can raise NEW findings (tagged `JUDGE-FOUND`).

### Step 3.3 — Mitigation check
For each severity-2+ finding, check if the issue is mitigated elsewhere (middleware, wrapper, validation layer). If mitigated, reduce severity or drop.

### Step 3.4 — Socratic challenge
For every HIGH/CRITICAL finding, ask: "Is the proposed fix worth the complexity? Does it introduce new failure modes? Is there a simpler approach?" Downgrade findings where the trade-off is net-negative.

### Step 3.5 — Confidence scoring
For each CONFIRMED finding, assign a confidence level (HIGH / MEDIUM / LOW). Findings with LOW confidence MUST be flagged for human review rather than silently included.

---

## Phase 4: Aggregation & Synthesis

Produce a structured report at `docs/audit/codebase_audit_report.md`.

### Report structure

```markdown
# Codebase Audit Report

## Executive Summary
- Overall health: ✅ Good / ⚠️ Fair / ❌ Poor (per layer: Architecture, Go Idiom, Security, Performance)
- Scope: packages analyzed, total LOC, lines of audit trail
- Tool findings: linter warnings, vulncheck results, vet issues

## Severity Legend
- CRITICAL: exploitable vulnerability or guaranteed misbehaviour
- HIGH: likely bug or policy violation
- MEDIUM: best-practice gap or latent risk
- LOW: style / maintainability suggestion

## Critical & High Findings
Each entry:
- **ID**: AUDIT-001
- **Severity**: CRITICAL
- **Location**: `path/to/file.go:42`
- **Observation**: concise description of the issue
- **Evidence**: excerpt from code or tool output
- **Risk**: what an attacker or user would experience
- **Remediation**: exact code change (diff block or snippet)
- **Judge Verdict**: CONFIRMED / WEAKENED / JUDGE-FOUND
- **Confidence**: HIGH / MEDIUM / LOW

## Medium & Low Findings
- Maintainability suggestions, architectural observations, style deviations
- Each MUST still cite `file:line`

## Verification Plan
- Commands to run to verify each fix (e.g., tests, lint, manual curl)
- Ordering: fixes that unblock testing first
```

### Phase 4 quality gates
- Every CRITICAL finding MUST be CONFIRMED by the Judge
- No finding appears without a `file:line` citation
- Remediation blocks MUST be syntactically valid Go (parse before writing)

---

## Post-Audit: AGENTS.md Update

After the report is written, append a summary entry to `AGENTS.md` (or `CLAUDE.md`) capturing:
- Patterns discovered during this audit (e.g., "always check context propagation in handler wrappers")
- Common false positives to suppress next run
- Tool invocations that worked well
