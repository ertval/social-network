# .opencode Agent Configuration — Change Plan

## Root Cause Fixed
- `{file:.opencode/agents/*.md}` paths resolved relative to config directory `.opencode/`, producing double-nested path `.opencode/.opencode/agents/*.md` → opencode crashed on startup.
- **Fix**: Removed `"agent"` block from `opencode.json`. All 5 agents now standalone markdown files in `.opencode/agents/` with YAML frontmatter.

## All Changes

### 1. Convert agents to standalone markdown frontmatter format
**Files**: `.opencode/agents/pr-create.md`, `pr-fix.md`, `pr-implement.md`, `pr-review.md`, `ticket-to-pr.md`

Each `.md` file now has YAML frontmatter with full config. Filename = agent name (opencode auto-discovers).

Config moved from `opencode.json` into frontmatter:
- `description`, `mode`, `model`, `color`, `steps`, `temperature`, `permission`

### 2. Removed explicit `"agent"` block from `opencode.json`
**File**: `.opencode/opencode.json`
- `"agent": { ... }` removed entirely — agents auto-discovered from `.opencode/agents/*.md`
- `"plugin"` key also removed — plugin auto-discovered from `.opencode/plugins/`

### 3. Subagents now deny spawning subagents
**Scope**: all 4 subagents (`pr-create`, `pr-fix`, `pr-implement`, `pr-review`)
```yaml
task:
  "*": deny
```
Prevents infinite recursion. Only `ticket-to-pr` (primary orchestrator) has `task: {"*": allow}`.

### 4. Distinct colors per agent type
| Agent | Color | Role |
|---|---|---|
| `pr-create` | `primary` | Push & PR creation |
| `pr-fix` | `warning` | Bugfix loop |
| `pr-implement` | `success` | Implementation |
| `pr-review` | `accent` | Code review |
| `ticket-to-pr` | `primary` | Orchestrator |

### 5. Added step limits (cost guardrails)
| Agent | Steps |
|---|---|
| `pr-create` | 30 |
| `pr-fix` | 25 |
| `pr-implement` | 50 |
| `pr-review` | 40 |
| `ticket-to-pr` | 60 |

### 6. `pr-review` edit permission — granular allow for report file
```yaml
edit:
  "*": deny
  "docs/reviews/PR_REVIEW_REPORT.md": allow
```
Agent needs `edit` to write the review report but should not modify code.

### 7. `pr-review` bash — added catch-all `"*": "ask"`
Missing `"*": "ask"` meant unknown bash commands had undefined permission. Added as first rule.

### 8. Removed stray Chinese characters in `ticket-to-pr.md`
Line 6 had `枪口` ("muzzle") — removed.

### 9. Graphify plugin — session reset for reminder
**File**: `.opencode/plugins/graphify.js`
- `reminded` flag now resets on `"session.created"` event, so each new session gets the graph hint.

### 10. `pr-implement.md` — removed stale "HumanLayer RPI" reference
Description now says "RPI framework" consistently; "HumanLayer" prefix dropped.

## Verification
- `opencode` starts with 0 config errors
- All 5 agents auto-discovered from `.opencode/agents/*.md`
- Prompt content loads correctly (1211-1577 chars each)
- Permissions, colors, steps, models match frontmatter config

## Files Modified
```
.opencode/opencode.json              — trimmed to just $schema (agents now .md)
.opencode/agents/pr-create.md        — standalone frontmatter agent
.opencode/agents/pr-fix.md           — standalone frontmatter agent
.opencode/agents/pr-implement.md     — standalone frontmatter agent, removed HumanLayer ref
.opencode/agents/pr-review.md        — standalone frontmatter agent, fine-grained edit perm
.opencode/agents/ticket-to-pr.md     — standalone frontmatter agent, removed stray chars
.opencode/plugins/graphify.js        — session.created resets reminded flag
.opencode/PLAN.md                    — this file
```
