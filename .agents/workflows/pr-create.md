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
   - **Rule Check**: Confirm that the branch name matches the pattern `<username>/<type>-<detail>` (e.g., `ekaramet/feat-s1-be-01-db-factory` or `dkotsi/fix-sqlite-busy-timeout`).
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
   - Locate the ticket in `docs/sprints/ticket-tracker.md` to confirm the sprint number and assignee group.
   - Open the specific sprint file: `docs/sprints/sprint-<number>.md` (e.g., `docs/sprints/sprint-1.md`).
   - Read the ticket's **Description**, **Detailed Steps**, and **Verification** sections.

2. **Verify Code Implementation**:
   - Run: `rtk git diff origin/main..HEAD` (or `rtk git diff main..HEAD`) to see all changes.
   - Cross-reference the diff against the **Detailed Steps** in the sprint ticket. Ensure every step has been fully addressed.
   - Check against `docs/sprints/general-instructions.md` and `.agents/rules/conventions.md`:
     - **Boundary Rules (D5)**: Feature roots, commands, queries, transport, and stores must not cross-import violating layers.
     - **Interface Strategy (D2)**: Narrow local interfaces for cross-slice behavior, duck typing.
     - **Cross-Slice Communication (D3)**: ID-only data references, async mutations via Event Bus.
     - **Database factory (D4)**: Store queries accept `platform/database.DB`.
     - **SQLite Limits**: Verify WAL mode is enabled, busy timeout is 5000, and connection pool is constrained (e.g., max 1 open for SQLite write paths to prevent lockups).
     - **TDD Requirement**: Ensure new commands or queries have corresponding unit/integration tests (`commands/*_test.go`, `queries/*_test.go`, `store/*_test.go`).
     - **Strangler Fig Steps (R1)**: For migration tickets, check that all Strangler Fig steps (Step 1-6) were fully executed (e.g., writing contract tests against old APIs, building new slice alongside, swapping routing in `bootstrap.go`, monitoring, and deleting old directories (`domain/`, `app/`, `infra/`)).
     - **Surgical Changes**: Ensure NO scope drift, unrelated formatting improvements, or pre-existing dead code cleanups. Remove imports/variables orphaned by your own changes.

3. **Run Verification Gates**:
   - Run full CI pipeline:
      - `rtk make ci`
   - Or run individually: `rtk make be-ci` (BE) / `rtk make fe-ci` (FE)
   - Do not proceed with PR creation if these validation gates fail.

---

## Phase 3: Draft the PR Message
Generate a clean, professional, and well-structured Markdown document containing the PR description. Save it locally as `.git/PR_DESCRIPTION.md` before publishing.

Follow premium typographic best practices:
- Use clean headings and dividers.
- Leverage details/summary dropdowns for verbose lists (e.g., file diffs or logs).
- Use tables for ticket metadata.
- Embed alerts for notable warnings, design choices, or migrations.

### PR Description Template

Copy `.github/PULL_REQUEST_TEMPLATE.md` into `.git/PR_DESCRIPTION.md` and fill in the details.

---

### Phase 4: Push and Create Pull Request
Once the PR message is written and saved to `.git/PR_DESCRIPTION.md`, execute the CLI commands to push the branch and open the PR in the Gitea platform.

1. **Verify Gitea CLI (tea) Connection**:
   - Run: `rtk tea whoami`
   - Confirm active credentials. If not logged in, ask the user to configure `tea`.

2. **Push the Branch**:
   - Run: `rtk git push -u origin <branch-name>`

3. **Fetch collaborators and create the Pull Request**:
   - First, dynamically fetch all repo collaborators via the Gitea API (excluding the PR author):
     ```bash
     COLLABORATORS=$(rtk curl -s \
       "https://platform.zone01.gr/git/api/v1/repos/dkotsi/social-network/collaborators" \
       -H "Authorization: token $(grep -A10 'zone01' ~/.config/tea/config.yml | grep token | awk '{print $2}')" \
       2>/dev/null | python3 -c "import json,sys; users=json.load(sys.stdin); print(','.join(u['login'] for u in users if u['login']!='$(rtk tea whoami 2>/dev/null | head -1)'))" 2>/dev/null)
     ```
   - Create the PR (note: `--reviewer` flag does NOT exist on `tea pulls create`):
     ```bash
     PR_OUTPUT=$(rtk tea pulls create \
       --title "[Ticket ID]: [Brief Title]" \
       --description "$(cat .git/PR_DESCRIPTION.md)" \
       --base main \
       --head [branch-name] \
       --output simple)
     echo "$PR_OUTPUT"
     ```
   - Extract the PR number and add reviewers via `tea pulls edit`:
     ```bash
     PR_NUMBER=$(echo "$PR_OUTPUT" | grep -oP '#\K\d+' | head -1)
     rtk tea pulls edit --add-reviewers "$COLLABORATORS" "$PR_NUMBER"
     ```
   - Print the generated PR URL and details to the user.
   - Clean up the temporary description file: `rm .git/PR_DESCRIPTION.md`.
