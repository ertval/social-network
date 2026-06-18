I want to create a robust, reusable codebase auditing system using agent workflows. Please execute the following tasks:

1. **Research 2026 Agentic Best Practices**:
   Perform web searches (20) through subagents to identify industry best practices (2026) regarding agent instructions for codebase audits, multi-agent orchestrator-judge setups, Socratic code reviews, and strategies to prevent LLM hallucinations during large codebase reviews.

2. **Generate a General Go Audit Workflow**:
   Create a general Go codebase audit workflow file at `.agents/workflows/go-code-audit-ds.md` using the correct agent frontmatter format (with name and description fields). This workflow must describe:
   - Phase 1: Deterministic grounding & tool scanning (e.g., `golangci-lint`, `govulncheck`).
   - Phase 2: Layered codebase analysis (Domain vs. Infrastructure decoupling, idiomatic Go concurrency/errors/context, SQL injection, WebSocket security, and performance/leaks).
   - Phase 3: Adversarial validation (using a "Judge/Critic" pass to verify findings against line references and screen for false positives).
   - Phase 4: Output synthesis to `docs/audit/codebase_audit_report.md`.

3. **Generate a Customized Social Network Audit Workflow**:
   Create a customized optimal workflow for this specific repository at `.agents/workflows/sn-code-audit-ds.md`. It must use the same frontmatter format and four-phase structure, but with detailed checks tailored to the Social Network codebase requirements:
   - Verification of only allowed packages (e.g., `gorilla/websocket`, `go-sqlite3`, `bcrypt`, `uuid`, migration tools).
   - Structural checks for the separation of Server, App, and Database layers, and startup database migrations.
   - SQLite WAL mode, connection pooling rules (1-10 max open connections), and busy timeouts in connection strings.
   - Follower request accept/decline flows, auto-follow on public profiles, and profile privacy access control (blocking non-followers).
   - Post/comment privacy scopes (`public`, `almost private` (followers only), and `private` (selected followers)), with image/GIF attachment support.
   - Group browse capabilities, creator approval for join requests, follower invitations, group-isolated chat rooms, and group event creation (Title, Description, Day/Time, Going/Not Going options).
   - WebSocket handshakes token verification, follow-based chat authorization, connection read limits, and deadlines.
   - Notification triggers for follow requests, group invites, group join requests, and event creations.
   - Bonus items checks (e.g., OAuth delegation, database seeding, confirmation popups, container build scripts).
