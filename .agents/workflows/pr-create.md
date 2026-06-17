---
name: pr-create
description: Pull Request verification and creation workflow (utilizing token-saving rtk command prefixes).
---

This workflow guides the agent through verifying branch conventions, checking the implemented ticket from sprint files, auditing commits/diffs, drafting a well-structured PR description, and publishing the PR using CLI tools wrapped in `rtk`.

---

## 🛠️ Execution Phases

### Phase 1: Branch and Commit Integrity Verification
Before creating the PR, verify that local branch naming and commit history adhere to project rules.

1. **Verify Current Branch Name**:
   - Run: `rtk git branch --show-current`
   - **Rule Check**: Confirm that the branch name matches the pattern `<username>/<type>-<detail>` (e.g., `ekaramet/feat-s1-be-01-db-factory` or `arnald/fix-sqlite-busy-timeout`).
   - If the branch does not match this convention, report the discrepancy immediately.

2. **Check Commit Messages**:
   - Run: `rtk git log -n 10 --oneline`
   - **Rule Check**: Verify that commits follow the Conventional Commits format `type(scope): description` (e.g., `feat(user): add register command with age validation`).
   - Ensure there are no merge commits from `main` (the branch should be rebased on `main`).

3. **Identify the Source Ticket**:
   - Find the ticket identifier (e.g., `S1-BE-01`, `S2-FE-04`) from the branch name or commit messages.
   - If not found, search the files to match modified code to a ticket or prompt the user.

---

## Phase 2: Verification Against Sprint Rules & Conventions
Verify that the implemented changes strictly comply with sprint tickets and core conventions.

1. **Locate the Ticket Metadata**:
   - Locate the ticket in `docs/plan/sprints/ticket-tracker.md` to confirm the sprint number and assignee group.
   - Open the specific sprint file: `docs/plan/sprints/sprint-<number>.md` (e.g., `docs/plan/sprints/sprint-1.md`).
   - Read the ticket's **Description**, **Detailed Steps**, and **Verification** sections.

2. **Verify Code Implementation**:
   - Run: `rtk git diff origin/main..HEAD` (or `rtk git diff main..HEAD`) to see all changes.
   - Cross-reference the diff against the **Detailed Steps** in the sprint ticket. Ensure every step has been fully addressed.
   - Check against `docs/plan/sprints/general-instructions.md` and `.agents/rules/conventions.md`:
     - **Boundary Rules (D5)**: Feature roots, commands, queries, transport, and stores must not cross-import violating layers.
     - **Interface Strategy (D2)**: Narrow local interfaces for cross-slice behavior, duck typing.
     - **Cross-Slice Communication (D3)**: ID-only data references, async mutations via Event Bus.
     - **Database factory (D4)**: Store queries accept `platform/database.DB`.
     - **SQLite Limits**: Verify WAL mode is enabled, busy timeout is 5000, and connection pool is constrained (e.g., max 1 open for SQLite write paths to prevent lockups).
     - **TDD Requirement**: Ensure new commands or queries have corresponding unit/integration tests (`commands/*_test.go`, `queries/*_test.go`, `store/*_test.go`).
     - **Surgical Changes**: Ensure NO scope drift, unrelated formatting improvements, or pre-existing dead code cleanups. Remove imports/variables orphaned by your own changes.

3. **Run Verification Gates**:
   - Run standard backend validation:
     - `rtk make ci` or `rtk make test`
   - Run standard frontend validation:
     - `rtk npm run lint` and `rtk npm run format:check`
   - Do not proceed with PR creation if these validation gates fail.

---

## Phase 3: Draft the PR Message
Generate a clean, professional, and well-structured Markdown document containing the PR description. Save it locally as `.git/PR_DESCRIPTION.md` before publishing.

Follow premium typographic best practices:
- Use clean headings and dividers.
- Leverage details/summary dropdowns for verbose lists (e.g., file diffs or logs).
- Use tables for ticket metadata.
- Embed alerts for notable warnings, design choices, or migrations.

### PR Description Template:
```markdown
# 🚀 Pull Request: [Ticket ID] — [Brief Title]

## 📋 Ticket Metadata
| Field | Value |
|---|---|
| **Ticket ID** | `[Ticket ID]` |
| **Assignee** | `[Name]` |
| **Sprint** | Sprint `[N]` |
| **Branch** | `[branch-name]` |

> [!NOTE]
> Resolves ticket: [Ticket Details](file:///docs/plan/sprints/sprint-[N].md#[Ticket-Anchor])

## 🔍 Overview & Rationale
*Describe high-level context of why this change was made, how it solves the ticket requirements, and any technical decisions.*

## 🛠️ Proposed Changes
### [Component / Slice Name]
- **[NEW / MODIFY / DELETE]** `[path/to/file.go](file:///path/to/file.go)`
  - *Detailed bullet points of specific additions or changes.*

### DB Migrations (if applicable)
- Added sequential migrations:
  - `[00000X_migration.up.sql](file:///db/migrations/00000X_migration.up.sql)`
  - `[00000X_migration.down.sql](file:///db/migrations/00000X_migration.down.sql)`

## ✅ Verification & Testing Results
*Provide evidence that the implementation works and satisfies the verification criteria.*

### Automated Test Output
```bash
rtk make test
# Paste short, successful test summary here
```

### Manual Smoke Tests
- [ ] Checked scenario `[e.g. A1 / B2]` from `general-instructions.md` → Result: `[Passed]`

## 🏁 Definition of Done (DoD) Checklist
- [x] Code conforms to D5 boundary rules (no cross-slice transport/store imports).
- [x] Concurrency and SQLite WAL, busy timeout, and pooling rules followed.
- [x] Unit/integration tests written and verified passing.
- [x] Format & Lint gates pass cleanly (`make ci`).
- [x] Branch named correctly and commits follow conventional naming.

---

### Phase 4: Push and Create Pull Request
Once the PR message is written and saved to `.git/PR_DESCRIPTION.md`, execute the CLI commands to push the branch and open the PR in the Gitea platform.

1. **Verify Gitea CLI (tea) Connection**:
   - Run: `rtk tea whoami`
   - Confirm active credentials. If not logged in, ask the user to configure `tea`.

2. **Push the Branch**:
   - Run: `rtk git push -u origin <branch-name>`

3. **Create the Pull Request**:
   - Run the non-interactive Gitea CLI command to publish the PR:
     ```bash
     rtk tea pulls create \
       --title "[Ticket ID]: [Brief Title]" \
       --description "$(cat .git/PR_DESCRIPTION.md)" \
       --base main \
       --head [branch-name] \
       --output simple
     ```
   - Print the generated PR URL and details to the user.
   - Clean up the temporary description file: `rm .git/PR_DESCRIPTION.md`.
