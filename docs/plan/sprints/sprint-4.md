# Sprint 4: Group & Event Features (Week 5)

**Outcome:** Groups with membership, group feed, group chat via WebSocket, and the event RSVP voting system work end-to-end.

> **Missing migration DDL:** Architecture specifies `000005_groups.up.sql` and `000006_events.up.sql`. S1-BE006 created `000005` and `000006` stubs (empty or minimal). S4-BE061 and S4-BE076 must extend these files with actual Group and Event table DDL — or create replacement migration files if stubs were not created. If S1-BE006 was skipped, create `000005` and `000006` migration pairs as part of S4-BE061 and S4-BE076.
>
> **GroupPost entity:** Architecture lists GroupPost (GroupID, AuthorID, Content, ImagePath, CreatedAt) as a named entity distinct from Topic posts. Explicitly defined in S4-BE060. Group feed (S4-BE071) queries GroupPost, not Topic.

---

### S4-BE059: Wire Group & Event bootstrap routes
* **Priority:** P0
* **Assignee:** BE-A + BE-B
* **Story Points:** 3
* **Dependencies:** S4-BE073, S4-BE074, S4-BE080
* **Description:** Register new slice routes in `bootstrap.go` so endpoints are live immediately after this sprint.
* **Detailed Steps:**
  1. In `internal/bootstrap/bootstrap.go`, import group and event transport packages.
  2. Call their route registration functions on the HTTP mux and WS router.
  3. Register event bus consumers for group events.
* **Verification:** `go build ./...` passes, new endpoints respond 200/401/403 (not 404).

---

## BE-A (Backend A) Tickets

### S4-BE060: Group: Entities & Repository Interface
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Establish domain model entities mapping group lifecycles.
* **Detailed Steps:**
  1. Create `internal/group/group.go`.
   2. Define `Group` (ID, Title, Description, CreatorID, CreatedAt), `GroupMember` (GroupID, UserID, Role: owner/member, JoinedAt), `Invitation` (GroupID, InviterID, InviteeID, Status: pending/accepted/declined, CreatedAt), `JoinRequest` (GroupID, RequesterID, Status, CreatedAt), and `GroupPost` (ID, GroupID, AuthorID, Content, ImagePath, CreatedAt).
* **Verification:** Compile check `go build ./internal/group/...`.

---

### S4-BE061: Group: SQLite Store
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE060, S4-SD016
* **Description:** Implement SQLite storage mapping group structure.
* **Detailed Steps:**
   1. Create `internal/group/store/sqlite.go`. Implement queries checking group membership: `IsMember(ctx context.Context, groupID, userID string) (bool, error)`.
* **Verification:** Integration tests checking memberships writing against the tables created by the `000005` migration.

---

### S4-BE062: Group: Create Group Command
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE060
* **Description:** Create group record and automatically set the creator as the group owner.
* **Detailed Steps:**
  1. Create `internal/group/commands/create_group.go`.
  2. Validate parameters (title, description bounds). Insert group, insert creator into members.
* **Verification:** Unit tests validating correct creator promotion.

---

### S4-BE063: Group: Invite Member Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE060
* **Description:** Invite follower to group, firing event notifications. Architecture requires invite-gating: only users who follow the inviter can be invited.
* **Detailed Steps:**
   1. Create `internal/group/commands/invite_member.go`. Ensure requester is group member.
   2. Define a local `FollowChecker` interface (same pattern as S2-BE022, S2-BE030): `AreConnected(ctx context.Context, a, b string) (bool, error)`.
   3. Before inserting invitation, verify that invitee follows the inviter. Reject with 403 if not connected.
   4. Insert invitation, publish `group.invited` event.
* **Verification:** Unit tests verifying: successful invite when follower, rejected invite when not connected, event outputs.

---

### S4-BE064: Group: Respond Invite Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE063
* **Description:** Accept/decline group invitations.
* **Detailed Steps:**
  1. Create `internal/group/commands/respond_invite.go`. If accepted -> insert user to member table.
* **Verification:** Check members mapping after accepts.

---

### S4-BE065: Group: Request Join Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE060
* **Description:** Submit request to join group, notifying owner.
* **Detailed Steps:**
  1. Create `internal/group/commands/request_join.go`. Insert record, publish `group.join_requested` event.
* **Verification:** Unit tests checking double request validation bounds.

---

### S4-BE066: Group: Respond Join Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE065
* **Description:** Allow group creator/owner to approve join requests.
* **Detailed Steps:**
  1. Create `internal/group/commands/respond_join.go`. Enforce that only the group creator can approve. If accepted -> add user to member list.
* **Verification:** Assert only owners can trigger, and member writes succeed.

---

### S4-BE067: Group: Create Post Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE060
* **Description:** Create post inside a group.
* **Detailed Steps:**
  1. Create `internal/group/commands/create_group_post.go`. Enforce group membership checks.
* **Verification:** Block creation for non-members, allow for members.

---

### S4-BE068: Group: Send Group Message Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE060
* **Description:** Post a message in the group chat, dispatching to WebSocket connections.
* **Detailed Steps:**
  1. Create `internal/group/commands/send_group_message.go`. Validate membership. Route message through WS coordinator.
* **Verification:** Test socket delivery payload verification.

---

### S4-BE069: Group: List Groups Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE060
* **Description:** Retrieve listing of all existing groups for browsing.
* **Detailed Steps:**
  1. Create `internal/group/queries/list_groups.go`.
* **Verification:** Test list pagination.

---

### S4-BE070: Group: Get Group Detail Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE060
* **Description:** Get specific group profile info.
* **Detailed Steps:**
  1. Create `internal/group/queries/get_group.go`.
* **Verification:** Tests asserting responses structures.

---

### S4-BE071: Group: Get Group Feed Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE060
* **Description:** Retrieve post list inside group. Enforce membership check.
* **Detailed Steps:**
  1. Create `internal/group/queries/get_group_feed.go`.
* **Verification:** Assert non-members cannot read feed details.

---

### S4-BE072: Group: Get Group Chat History Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE060
* **Description:** Get group message history log.
* **Detailed Steps:**
  1. Create `internal/group/queries/get_group_chat.go`.
* **Verification:** Retrieve log chronologically.

---

### S4-BE073: Group: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE062..S4-BE072
* **Description:** Bind HTTP routes. Every command and query must have at least one route.
* **Detailed Steps:**
   1. Create `internal/group/transport/http.go`.
   2. Route:
      - `POST /api/groups` — create_group (S4-BE062)
      - `GET /api/groups` — list_groups (S4-BE069)
      - `GET /api/groups/:id` — get_group (S4-BE070)
      - `GET /api/groups/:id/feed` — get_group_feed (S4-BE071)
      - `GET /api/groups/:id/chat` — get_group_chat (S4-BE072)
      - `POST /api/groups/:id/invite` — invite_member (S4-BE063)
      - `POST /api/groups/:id/invite/respond` — respond_invite (S4-BE064)
      - `POST /api/groups/:id/join` — request_join (S4-BE065)
      - `POST /api/groups/:id/join/respond` — respond_join (S4-BE066)
      - `POST /api/groups/:id/posts` — create_group_post (S4-BE067)
* **Verification:** Mock requests integration tests. Every command handler has a corresponding route.

---

### S4-BE074: Group: WS Transport Routing
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE073
* **Description:** Route real-time WS chat events.
* **Detailed Steps:**
  1. Create `internal/group/transport/ws.go`. Connect to core WebSocket.
* **Verification:** Test messaging over connections.

---

## BE-B (Backend B) Tickets

### S4-BE075: Event: Entities & Repository Interface
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Establish domain structures for events and RSVP votes. Use separate `Option` entity per the architecture spec for extensibility (e.g. adding "maybe" later).
* **Detailed Steps:**
   1. Create `internal/event/event.go`. Define `Event` (ID, GroupID, CreatorID, Title, Description, ScheduledTime, CreatedAt), `Option` (EventID, Label string, e.g. "going", "not_going"), and `EventRSVP` (EventID, UserID, OptionID, UpdatedAt).
* **Verification:** Build checks.

---

### S4-BE076: Event: SQLite Store
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S4-BE075, S4-SD016
* **Description:** Store query mappings for events.
* **Detailed Steps:**
   1. Create `internal/event/store/sqlite.go`.
* **Verification:** Integration tests checking storage against the tables created by the `000006` migration.

---

### S4-BE077: Event: Create Event Command
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S4-BE075
* **Description:** Create event in group, executing group membership constraints, and publishing to notifications.
* **Detailed Steps:**
  1. Create `internal/event/commands/create_event.go`.
  2. Define local `GroupMemberChecker` interface. Enforce creator is group member.
  3. Validate fields: title, description, time in future, minimum of 2 RSVP options.
  4. Publish `event.created` event on EventBus.
* **Verification:** Unit tests verifying validator constraints, member checking, and notifications triggering.

---

### S4-BE078: Event: RSVP Command
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S4-BE075
* **Description:** Record RSVP choices (going/not going).
* **Detailed Steps:**
  1. Create `internal/event/commands/rsvp.go`. Upsert choices.
* **Verification:** Unit tests verifying updates to RSVP states.

---

### S4-BE079: Event: List Group Events Query
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S4-BE075
* **Description:** List events under a group with aggregated vote tallies.
* **Detailed Steps:**
  1. Create `internal/event/queries/list_group_events.go`.
* **Verification:** Test return outputs contain correct count aggregates.

---

### S4-BE080: Event: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S4-BE077..S4-BE079
* **Description:** Bind HTTP handlers.
* **Detailed Steps:**
  1. Create `internal/event/transport/http.go`.
  2. Route `POST /api/groups/:id/events`, `GET /api/groups/:id/events`, `POST /api/events/:id/rsvp`.
* **Verification:** Integration test endpoints mapping.

---

## FE-A (Frontend A) Tickets

### S4-FE019: Groups Directory Page
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Implement browse view lists and search filters.
* **Detailed Steps:**
  1. Create `/groups` path fetching all items. Add search text filter input.
* **Verification:** Test listing render.

---

### S4-FE020: Group Profile Page
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 5
* **Description:** Render detailed group info, member listings, and Join/Invite actions.
* **Detailed Steps:**
  1. Create `/groups/[id]` path. Show tabs: Feed, Events, Chat, Members.
* **Verification:** Interactive buttons clicks verification.

---

### S4-FE021: Group Posts Feed
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Implement publishing post inside group space.
* **Detailed Steps:**
  1. Add post submission layout that sends group parameters.
* **Verification:** Submissions test verification.

---

### S4-FE022: Group Chat Workspace
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 5
* **Description:** Build WebSocket messaging frame under group tabs.
* **Detailed Steps:**
  1. Open websocket channel, map incoming group payloads, display typing events.
* **Verification:** Check socket routing parameters.

---

## FE-B (Frontend B) Tickets

### S4-FE023: Event Creation Dialog
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 5
* **Description:** Render event creator box inside group.
* **Detailed Steps:**
  1. Inputs for Title, Description, Date/Time picker, and RSVP option toggles.
* **Verification:** Form validates correct inputs before postings.

---

### S4-FE024: Events List Component
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Display group events detailing description, time and current RSVP numbers.
* **Detailed Steps:**
  1. Render lists displaying stats counters.
* **Verification:** Verify counters load correctly.

---

### S4-FE025: RSVP Switch Actions
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 2
* **Description:** Instant updates when voting RSVP.
* **Detailed Steps:**
  1. Click going/not going button updates choice immediately via websocket/REST call.
* **Verification:** Unit tests verifying update tallies visual checks.

---

## SD-QA (System Design/QA) Tickets

### S4-SD017: E2E: Complete Groups Workspace Journey
* **Priority:** P0
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** Playwright testing group cycles.
* **Detailed Steps:**
  1. Script: Create Group -> Join members -> Post message -> Create Event -> Vote RSVP.
* **Verification:** Execution completes successfully in CI.

---

### S4-SD016: Platform: Group & Event Migrations (000005 & 000006)
* **Priority:** P0
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** S1-BE006
* **Description:** Create the database migration files for the Group and Event vertical slices (Phases 2.4 / 4).
* **Detailed Steps:**
  1. Create `db/migrations/000005_groups.up.sql` to create tables: `groups`, `group_members`, `group_invitations`, `group_join_requests`, `group_posts`, `group_chat_messages`.
  2. Create `db/migrations/000005_groups.down.sql` to reverse these changes.
  3. Create `db/migrations/000006_events.up.sql` to create tables: `events`, `event_options`, `event_rsvps`.
  4. Create `db/migrations/000006_events.down.sql` to reverse these changes.
* **Verification:** Run `make db-reset` or execute the migration runner and verify that these migrations apply and rollback cleanly.
