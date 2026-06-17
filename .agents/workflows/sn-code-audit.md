---
name: sn-code-audit
description: This agent workflow is configured to perform a deep software engineering, security, and performance audit specifically for the Social Network codebase, mapping directly to grading and design requirements.
---

Perform a deep and comprehensive software engineering, security, and performance audit of the Go backend and architecture in this repository. 

Since this is a multi-layered codebase audit, please break down your progress systematically into four sequential execution phases to ground analysis, minimize noise, and prevent hallucinations.

---

## 🛠️ Execution Phases

### Phase 1: Deterministic Grounding & Tool Scanning
Before beginning any cognitive analysis, run local static verification tools to establish a baseline of facts:
1. Run local linters (e.g., `golangci-lint run`) to check style, format, and static code quality rules defined in [.golangci.yml](file:///home/ertval/code/zone-modules/social-network/.golangci.yml).
2. Run standard dependency security checkers (e.g., `govulncheck` or the built-in security scanner skill) to trace known vulnerabilities in dependencies defined in [go.mod](file:///home/ertval/code/zone-modules/social-network/go.mod).
3. Record the exact output of these runs to serve as the source of truth for downstream phases.

### Phase 2: Layered Codebase Analysis
Audit the codebase systematically against these specific areas:

1. **Software Design & Architecture (Domain-Driven Design / Clean Architecture)**
   - **Responsibility Decoupling**: Check how clean the boundary is between the domain entities/interfaces in [internal/domain](file:///home/ertval/code/zone-modules/social-network/internal/domain) and infrastructure implementations in [internal/infra](file:///home/ertval/code/zone-modules/social-network/internal/infra). Confirm separation among Server (entry points), App (use cases & listener core), and Database (migrations & queries).
   - **Migrations System**: Check if the folder structure matches or behaves similarly to the target layout (e.g., `db/migrations/sqlite` with `*.up.sql` and `*.down.sql` migrations). Verify that migrations are applied automatically at startup.
   - **Dependency Management**: Evaluate the initialization and dependency injection setups in [internal/bootstrap](file:///home/ertval/code/zone-modules/social-network/internal/bootstrap) and [internal/infra/services.go](file:///home/ertval/code/zone-modules/social-network/internal/infra/services.go). Adhere to the allowed packages list from [docs/readme.md](file:///home/ertval/code/zone-modules/social-network/docs/readme.md):
     - `database/sql` & Standard Go packages
     - `github.com/gorilla/websocket`
     - `github.com/mattn/go-sqlite3`
     - `golang.org/x/crypto/bcrypt`
     - `github.com/gofrs/uuid` or `github.com/google/uuid`
     - Authorized migration libraries (`golang-migrate`, `sql-migrate`, `Boostport/migration`)

2. **Idiomatic Go Best Practices**
   - **Error Handling**: Check error handling patterns: Are errors wrapped correctly using `%w`? Are there naked/ignored errors? Is panic recovery handled properly in goroutines?
   - **Concurrency Safety**: Ensure mutexes (e.g., `sync.RWMutex`), channels, and waitgroups are correctly used. Verify `context.Context` is propagated all the way down to database calls and HTTP clients.
   - **Linter Compliance**: Verify alignment with [effective Go guidelines](https://go.dev/doc/effective_go) and standards configured in [.golangci.yml](file:///home/ertval/code/zone-modules/social-network/.golangci.yml).

3. **Code Security & Functional Specification Review**
   - **Authentication & Registration**:
     - Check if custom session/cookie handling (`internal/infra/http/authcookies`) uses strong secure parameters (HttpOnly, Secure, SameSite).
     - Verify the registration form handles: Email, Password, First Name, Last Name, Date of Birth, and the optional fields: Avatar/Image, Nickname, and About Me. Ensure passwords are securely hashed with `bcrypt`.
   - **SQL Injection**: Inspect all database queries in `internal/infra/storage` to ensure parameters are bound correctly and no raw string concatenation is used.
   - **SQLite Connection Safety**: Verify SQLite WAL mode is enabled (`_journal_mode=WAL`) and a busy timeout is configured (`_busy_timeout=5000`) in the DSN connection string to prevent locking issues.
   - **Access Control & Profile Privacy**:
     - Verify profile routes ensure non-followers are blocked from viewing private profiles, while allowing followers to view them.
     - Check that users can toggle their profile between public and private.
     - Verify follower requests flow: follow request and accept/decline logic for private profiles, and direct auto-follow for public profiles.
   - **Posts & Comments Privacy**:
     - Verify posts and comments can include attached images (JPG, PNG) or GIFs.
     - Validate post privacy scope rules:
       - `public`: Visible to all logged-in users.
       - `almost private`: Visible only to followers.
       - `private`: Visible only to specifically selected followers.
   - **Groups & Group Events**:
     - Verify group membership controls: invitation acceptance/declination by followers, join request approvals by the creator.
     - Ensure group browse capability is functional.
     - Validate that group posts and comments are only visible to group members.
     - Verify group event model: Title, Description, Day/Time, and options ("Going", "Not going").
   - **Websocket Chat Security & Real-Time**:
     - Validate real-time websocket connections in `internal/infra/ws` and `internal/infra/realtime`. Ensure authentication token verification happens *during the upgrade handshake*.
     - Verify chat authorization: prevent chat creation between users unless at least one follows the other.
     - Ensure support for sending emojis.
     - Verify group chat room logic: only group members can send and receive messages in the group room.
     - Check resource limits: ensure `conn.SetReadLimit` is set to block oversized payloads, and timeouts (`SetReadDeadline`/`SetWriteDeadline`) prevent dead connection hanging.
   - **Notifications Engine**:
     - Validate that notifications are available on all pages.
     - Verify triggers generate notifications for:
       - Follow request received (for private profiles).
       - Group invitation received.
       - Group join request received (sent to the creator).
       - New event created in a group.

4. **Performance & Concurrency Issues**
   - Identify database bottlenecks (e.g., N+1 query loops, missing indexes on SQLite tables, leaks in connection pools).
   - Verify that connection pool settings (`SetMaxOpenConns` and `SetMaxIdleConns`) are set conservatively (e.g., 1-10 connections) to avoid file-locking contention under high concurrency.
   - Look for goroutine leaks, potential race conditions in state tracking, or memory allocation overheads in API responses.

5. **Bonus Capabilities Audit**
   - Validate if OAuth authentication (GitHub/OAuthenticator) is integrated.
   - Check if there is an automated database seed migration to pre-fill content.
   - Verify presence of confirmation pop-ups for actions like unfollowing or profile privacy toggling.
   - Inspect build/deployment helper scripts (e.g. `entrypoint.sh` or docker scripts) to ensure correct container compilation.

### Phase 3: Adversarial Validation (The "Judge" Pass)
Evaluate the draft findings through a validation lens to eliminate false positives and hallucinations:
1. **Evidence Check**: For each issue found, check if it cites an exact file path and line number. Verify the claim by reading the code.
2. **Context Verification**: Check if the identified issue is actually mitigated elsewhere (e.g., input is already validated or sanitized in a middleware).
3. **Socratic Critique**: Challenge recommendations on trade-offs regarding readability, complexity, and performance.

### Phase 4: Aggregation & Synthesis
Aggregate verified findings into the final report.

---

## 📝 Output Report Schema

Output your findings as a markdown artifact named `docs/audit/codebase_audit_report.md` structured with:
- **Executive Summary**: Overall health status (Design, Idiomatic Go, Security, Performance).
- **Critical & High Severity Findings**: Code snippets, path links, impact explanation, and immediate remediation code blocks.
- **Medium & Low Findings**: Maintainability improvements and architectural suggestions.
- **Verification Plan**: Step-by-step verification methods (including test suites to run).
