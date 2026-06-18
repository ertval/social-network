## ticket-to-pr

End-to-end orchestrator that takes a ticket ID and sequentially spawns subagents to implement, review, fix, and publish the PR. Handles the review-fix loop with a 3-strike limit.

## Core Loop
枪口
1. **Locate** the ticket in `docs/sprints/ticket-tracker.md` and read its sprint spec.
2. **Implement**: spawn `pr-implement` subagent → code + tests on a feature branch.
3. **Review**: spawn `pr-review` subagent → run deterministic gates + full review pipeline → save report to `docs/reviews/PR_REVIEW_REPORT.md`.
4. **Fix loop**: if review is `🔴 CHANGES REQUESTED`, spawn `pr-fix` subagent → re-run review. Repeat up to 3 times.
5. **Create PR**: on `🟢 APPROVED` or `🟡 PASS WITH RECOMMENDATIONS`, spawn `pr-create` subagent → push branch + open PR via `tea`.

## Rules

- Run subagents **sequentially** (do not spawn in parallel). Each phase depends on the previous phase's output.
- Each subagent gets the full workflow file path and the ticket ID in its prompt.
- On review failure, read `docs/reviews/PR_REVIEW_REPORT.md` to confirm severity before deciding to loop.
- After 3 review failures, stop and present unresolved findings to the user.
- Do not skip phases. Do not combine subagent responsibilities.
