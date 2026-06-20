---
name: go-code-audit
description: A general agent workflow configured to perform a deep software engineering, security, and performance audit of any Go codebase.
---

Perform a deep and comprehensive software engineering, security, and performance audit of the Go backend and architecture in this repository. 

Since this is a multi-layered codebase audit, please break down your progress systematically into four sequential phases to ground analysis, minimize noise, and prevent hallucinations.

---

## 🛠️ Execution Phases

### Phase 1: Deterministic Grounding & Tool Scanning
Before beginning any cognitive analysis, run local static verification tools to establish a baseline of facts.
1. Run local linters (e.g., `golangci-lint run`) to check style, format, and static code quality rules.
2. Run standard dependency security checkers (e.g., `govulncheck`) to trace known vulnerabilities in dependencies.
3. Record the exact output of these runs to serve as the source of truth for downstream phases.

### Phase 2: Layered Codebase Analysis
Analyze the codebase systematically, focusing on the following core areas:

1. **Software Design & Architecture**
   - **Decoupling**: Verify clean boundaries between the core domain logic/interfaces and infrastructure implementations (adapters, database drivers, third-party services).
   - **Dependency Management**: Evaluate the initialization and dependency injection setup. Ensure configurations are loaded securely and components are mockable for testing.
   - **SOLID Principles**: Check for tight coupling, clean interface segmentation, and adherence to clean design guidelines.

2. **Idiomatic Go Best Practices**
   - **Error Handling**: Check if errors are wrapped correctly using `%w` and never silently ignored. Ensure panics are recovered in goroutines.
   - **Concurrency Safety**: Verify correct usage of synchronization primitives (`sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, channels). Ensure there are no potential data races or goroutine leaks.
   - **Context Propagation**: Ensure `context.Context` is propagated correctly down to HTTP handlers, database queries, and downstream calls for timeout and cancellation handling.
   - **Code Style**: Check for consistency with idiomatic Go style and linter configurations.

3. **Code Security**
   - **SQL Injection**: Inspect all database queries (SQL/NoSQL) to ensure user-supplied parameters are bound securely and not concatenated raw into query strings.
   - **Authentication & Sessions**: Review how session tokens/cookies are configured (e.g. Secure, HttpOnly, SameSite, expiration) and verify password hashing (e.g. bcrypt) parameters.
   - **Network & API Handlers**: Validate input parameters on all entry points. For real-time communication (e.g., WebSockets), verify origin checks, validation on incoming payloads, read/write message size limits, and timeouts.

4. **Performance & Resource Management**
   - **Database Efficiency**: Identify query bottlenecks (e.g., N+1 loops, missing indices, or transaction/connection leaks).
   - **Memory/Allocation**: Review hot paths for unnecessary allocations. Suggest pre-allocating slices or maps where appropriate.
   - **Resource Lifecycles**: Ensure files, network sockets, database connections, and channels are closed correctly using `defer`.

### Phase 3: Adversarial Validation (The "Judge" Pass)
Evaluate the draft findings through a validation lens to eliminate false positives and hallucinations:
1. **Evidence Check**: For each issue found, check if it cites an exact file path and line number. Verify the claim by reading the code.
2. **Context Verification**: Check if the identified issue is actually mitigated elsewhere (e.g., input is already validated or sanitized in a middleware).
3. **Socratic Critique**: Challenge recommendations on trade-offs regarding readability, complexity, and performance.

### Phase 4: Aggregation & Synthesis
Aggregate verified findings into the final report.

---

## 📝 Output Report Schema (`codebase_audit_report.md`)

Output your findings as a markdown artifact named `codebase_audit_report.md` structured with:
- **Executive Summary**: Overall health status (Design, Idiomatic Go, Security, Performance).
- **Critical & High Severity Findings**: Code snippets, path links, impact explanation, and immediate remediation code blocks.
- **Medium & Low Findings**: Maintainability improvements and architectural suggestions.
- **Verification Plan**: Step-by-step verification methods (including test commands to run).
