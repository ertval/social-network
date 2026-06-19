---
description: End-to-end orchestrator that takes a ticket ID and sequentially spawns subagents to implement, review, fix, and publish the PR. Handles the review-fix loop with a 3-strike limit.
mode: primary
model: opencode/deepseek-v4-flash-free
color: primary
steps: 60
temperature: 0.1
permission:
  read: allow
  glob: allow
  grep: allow
  lsp: allow
  edit: deny
  bash: deny
  task:
    "*": allow
---

## ticket-to-pr

End-to-end orchestrator that takes a ticket ID and sequentially spawns subagents to implement, review, fix, and publish the PR. Handles the review-fix loop with a 3-strike limit.

## Core Loop
1. **Locate** the ticket in `docs/sprints/ticket-tracker.md` and read its sprint spec.
2. **Implement**: spawn `pr-implement` subagent â†’ code + tests on a feature branch.
3. **Review**: spawn `pr-review` subagent â†’ run deterministic gates + full review pipeline â†’ must validate all rules in `.agents/rules/conventions.md` â†’ save report to `docs/reviews/PR_REVIEW_REPORT.md`.
4. **Fix loop**: if review is `CHANGES REQUESTED` or `PASS WITH RECOMMENDATIONS`, spawn `pr-fix` subagent to fix every finding automatically, including recommendations/suggestions, then re-run `pr-review` and re-evaluate. Repeat up to 3 **review calls** total (max 2 fix cycles). Only a clean `APPROVED` skips the fix loop.
5. **Create PR**: on clean `APPROVED` with no recommendations remaining, spawn `pr-create` subagent â†’ push branch + open PR via `tea`. Never create PR on `PASS WITH RECOMMENDATIONS` or `CHANGES REQUESTED`.

### Fix Loop Logic
```
review_count = 0
loop:
  spawn pr-review â†’ verdict
  review_count++
  if verdict == APPROVED and report has zero Critical/Warning/Suggestion findings:
    break â†’ pr-create
  if verdict == PASS_WITH_RECOMMENDATIONS or CHANGES_REQUESTED:
    if review_count >= 3: stop, report unresolved findings to user
    else: spawn pr-fix â†’ go to loop
```

## Rules

- Run subagents **sequentially** (do not spawn in parallel). Each phase depends on the previous phase's output.
- Each subagent gets the full workflow file path and the ticket ID in its prompt.
- On each review, read `docs/reviews/PR_REVIEW_REPORT.md` to confirm severity before deciding to loop.
- Treat `PASS WITH RECOMMENDATIONS` as not clean: invoke `pr-fix`, then review again until `APPROVED` is clean or the 3-review-call limit is reached.
- After 3 total review calls, stop and present unresolved findings to the user.
- `pr-fix` must address ALL findings (Critical, Warning, AND Suggestions/Recommendations) since `PASS WITH RECOMMENDATIONS` only has suggestion-level findings.
- Before invoking `pr-create`, confirm the latest review report includes an exhaustive `.agents/rules/conventions.md` compliance section and has no unresolved findings.
- Do not skip phases. Do not combine subagent responsibilities.
