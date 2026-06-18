# PR: [Ticket ID] — [Brief Title]

**Ticket ID** `[Ticket ID]` · **Sprint** `[N]` · **Branch** `[branch-name]`

Resolves [Ticket Details](file://docs/sprints/sprint-[N].md#[Ticket-Anchor])

---

#### Why

What problem does this solve? Brief context and technical rationale.

---

#### What

- **[NEW / MODIFY / DELETE]** `path/to/file.go` — what changed and why

##### DB Migrations

- `00000X_migration.up.sql`
- `00000X_migration.down.sql`

---

#### Audit Coverage

| Requirement | Status | Component | Notes |
|---|---|---|---|
| `/register` (G1) | | `RegisterForm` | |
| `/login` | | `LoginForm` | |
| `/profile/[id]` (G2/G10) | | `ProfileCard` | |
| `FollowButton` (G6/G10) | | `FollowButton` | |
| `/post/new` (G4/G5) | | `PostForm` | |
| `/groups` (G7) | | `GroupDirectory` | |
| `/groups/[id]` (G7) | | `JoinRequestButton` | |
| `/groups/[id]/events` (G7) | | `EventForm` | |
| `/groups/[id]/chat` (G6) | | `GroupChatWindow` | |
| `/chat/[userId]` (G6/G8) | | `ChatWindow` | |
| `NotificationBell` (G3) | | `NotificationBell` | |

---

#### Verification

```bash
# test output
```

- [ ] Smoke test scenario `[A1 / B2]` → Result: `[Passed]`

#### DoD

- [x] D5 boundary rules followed
- [x] Concurrency & SQLite rules followed
- [x] Tests passing (Vitest / Go test)
- [x] Type checking clean (`tsc --noEmit` / `go vet`)
- [x] Lint clean (`make ci` / Biome)
- [x] Branch name & commit convention correct
