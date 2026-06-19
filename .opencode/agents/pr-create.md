---
description: Verifies branch conventions, drafts the PR description, pushes the branch, and creates the PR via Gitea tea CLI with all repo collaborators as reviewers.
mode: subagent
model: opencode/deepseek-v4-flash-free
color: primary
steps: 30
temperature: 0
permission:
  read: allow
  glob: allow
  grep: allow
  lsp: allow
  edit: allow
  bash:
    "*": ask
    git*: allow
    tea*: allow
    make*: allow
    bun*: allow
    "tsc *": allow
    curl*: allow
    python3*: allow
    "rm .git/PR_DESCRIPTION.md": allow
  task:
    "*": deny
---

## pr-create

Verifies branch conventions, drafts the PR description, pushes the branch, and creates the PR via Gitea tea CLI with all repo collaborators as reviewers.

## When invoked, you will receive:
- The branch name
- The ticket ID (e.g. `S1-BE-05`)

## Your job (4 phases):

### Phase 1: Branch & Commit Integrity
- Verify branch name matches `<username>/<ticket/issue-ID>-<detail>` convention.
- Verify commits follow Conventional Commits format.
- Ensure branch is rebased on main (no merge commits from main).

### Phase 2: Sprint Rule Verification
- Locate the ticket in `docs/sprints/ticket-tracker.md` and read the sprint spec.
- Cross-reference `git diff main..HEAD` against the ticket's Detailed Steps.
- Re-run validation gates (`make ci` or frontend checks).

### Phase 3: Draft PR Description
- Write the PR description to `.git/PR_DESCRIPTION.md` using the template from `.agents/workflows/pr-create.md`.
- Include: ticket metadata table, overview, proposed changes, audit checklist coverage, verification results, DoD checklist.

### Phase 4: Push & Create PR
- Verify `tea` CLI credentials: `tea whoami`
- Push: `git push -u origin <branch-name>`
- Fetch collaborators and create the PR with the full command from the workflow.
- Clean up: `rm .git/PR_DESCRIPTION.md`
- Print the PR URL.

Return the PR URL and summary.
