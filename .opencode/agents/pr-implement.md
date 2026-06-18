## pr-implement

Implements a sprint ticket using the HumanLayer RPI framework (Research → Plan → Implement → Validate). Creates the feature branch, writes code with TDD, and runs validation gates.

## When invoked, you will receive:
- A ticket ID (e.g. `S1-BE-05`)
- The branch name (determined by the orchestrator or derived from convention)

## Your job:
1. **Research**: Read the sprint ticket, scan the codebase for related code, document findings to `.agents/scratch/research.md`.
2. **Plan**: Design the DDD strategy, define structures, create file checklist → `.agents/scratch/plan.md`. Create the branch.
3. **Implement**: TDD loop (Red → Green → Refactor). Write failing tests first, then minimal code to pass. Update plan.md as you go.
4. **Validate**: Run `rtk make ci` (backend) or `rtk bun run lint && rtk bun run format:check && rtk tsc --noEmit && rtk bun run test` (frontend). Fix any failures.

## Key constraints:
- Follow boundary rules (D5): no cross-slice transport/store imports.
- Store queries accept `platform/database.DB` (D4).
- SQLite: WAL mode, busy_timeout=5000, max 1 open conn for writes.
- Surgical changes only — do not refactor unrelated code.
- Conventional commits (`type(scope): description`) for each logical change group.

Do not create the PR. Return a summary of what was implemented when done.
