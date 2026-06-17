# Sprint 4: Group & Event Features (Week 9–10)

**Outcome:** Groups with membership, group feed, group chat via WebSocket, and the event RSVP voting system work end-to-end.

> **Missing migration DDL:** Architecture specifies `000005_groups.up.sql` and `000006_events.up.sql`. S1-BE-04 created `000005` and `000006` stubs (empty or minimal). S4-BE-02 and S4-BE-17 must extend these files with actual Group and Event table DDL — or create replacement migration files if stubs were not created. If S1-BE-04 was skipped, create `000005` and `000006` migration pairs as part of S4-BE-02 and S4-BE-17.
>
> **GroupPost entity:** Architecture lists GroupPost (GroupID, AuthorID, Content, ImagePath, CreatedAt) as a named entity distinct from Topic posts. Explicitly defined in S4-BE-01. Group feed (S4-BE-12) queries GroupPost, not Topic.

---

### S4-BE-JOINT: Wire Group & Event bootstrap routes
* **Priority:** P0
* **Assignee:** BE-A + BE-B
* **Story Points:** 3
* **Dependencies:** S4-BE-14, S4-BE-15, S4-BE-21
* **Description:** Register new slice routes in `bootstrap.go` so endpoints are live immediately after this sprint.
* **Detailed Steps:**
  1. In `internal/bootstrap/bootstrap.go`, import group and event transport packages.
  2. Call their route registration functions on the HTTP mux and WS router.
  3. Register event bus consumers for group events.
* **Verification:** `go build ./...` passes, new endpoints respond 200/401/403 (not 404).

---

## BE-A (Backend A) Tickets

### S4-BE-01: Group: Entities & Repository Interface
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Establish domain model entities mapping group lifecycles.
* **Detailed Steps:**
  1. Create `internal/group/group.go`.
   2. Define `Group` (ID, Title, Description, CreatorID, CreatedAt), `GroupMember` (GroupID, UserID, Role: owner/member, JoinedAt), `Invitation` (GroupID, InviterID, InviteeID, Status: pending/accepted/declined, CreatedAt), `JoinRequest` (GroupID, RequesterID, Status, CreatedAt), and `GroupPost` (ID, GroupID, AuthorID, Content, ImagePath, CreatedAt).
* **Verification:** Compile check `go build ./internal/group/...`.

---

### S4-BE-02: Group: SQLite Store
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE-01
* **Description:** Implement SQLite storage mapping group structure. If `000005_groups.up.sql` does not yet contain Group table DDL, create it here (or extend the existing migration).
* **Detailed Steps:**
   1. Create `internal/group/store/sqlite.go`. Implement queries checking group membership: `IsMember(ctx context.Context, groupID, userID string) (bool, error)`.
   2. **If S1-BE-04 left `000005_groups.up.sql` as a stub:** replace with full DDL: `CREATE TABLE groups (...)`, `group_members`, `group_invitations`, `group_join_requests`, `group_chat_messages`, `group_posts`. Provide corresponding `.down.sql`.
* **Verification:** Integration tests checking memberships writing. Verify `000005` migration creates all Group tables.

---

### S4-BE-03: Group: Create Group Command
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE-01
* **Description:** Create group record and automatically set the creator as the group owner.
* **Detailed Steps:**
  1. Create `internal/group/commands/create_group.go`.
  2. Validate parameters (title, description bounds). Insert group, insert creator into members.
* **Verification:** Unit tests validating correct creator promotion.

---

### S4-BE-04: Group: Invite Member Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE-01
* **Description:** Invite follower to group, firing event notifications. Architecture requires invite-gating: only users who follow the inviter can be invited.
* **Detailed Steps:**
   1. Create `internal/group/commands/invite_member.go`. Ensure requester is group member.
   2. Define a local `FollowChecker` interface (same pattern as S2-BE-08, S2-BE-17): `AreConnected(ctx context.Context, a, b string) (bool, error)`.
   3. Before inserting invitation, verify that invitee follows the inviter. Reject with 403 if not connected.
   4. Insert invitation, publish `group.invited` event.
* **Verification:** Unit tests verifying: successful invite when follower, rejected invite when not connected, event outputs.

---

### S4-BE-05: Group: Respond Invite Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE-04
* **Description:** Accept/decline group invitations.
* **Detailed Steps:**
  1. Create `internal/group/commands/respond_invite.go`. If accepted -> insert user to member table.
* **Verification:** Check members mapping after accepts.

---

### S4-BE-06: Group: Request Join Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE-01
* **Description:** Submit request to join group, notifying owner.
* **Detailed Steps:**
  1. Create `internal/group/commands/request_join.go`. Insert record, publish `group.join_requested` event.
* **Verification:** Unit tests checking double request validation bounds.

---

### S4-BE-07: Group: Respond Join Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE-06
* **Description:** Allow group creator/owner to approve join requests.
* **Detailed Steps:**
  1. Create `internal/group/commands/respond_join.go`. Enforce that only the group creator can approve. If accepted -> add user to member list.
* **Verification:** Assert only owners can trigger, and member writes succeed.

---

### S4-BE-08: Group: Create Post Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE-01
* **Description:** Create post inside a group.
* **Detailed Steps:**
  1. Create `internal/group/commands/create_group_post.go`. Enforce group membership checks.
* **Verification:** Block creation for non-members, allow for members.

---

### S4-BE-09: Group: Send Group Message Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE-01
* **Description:** Post a message in the group chat, dispatching to WebSocket connections.
* **Detailed Steps:**
  1. Create `internal/group/commands/send_group_message.go`. Validate membership. Route message through WS coordinator.
* **Verification:** Test socket delivery payload verification.

---

### S4-BE-10: Group: List Groups Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE-01
* **Description:** Retrieve listing of all existing groups for browsing.
* **Detailed Steps:**
  1. Create `internal/group/queries/list_groups.go`.
* **Verification:** Test list pagination.

---

### S4-BE-11: Group: Get Group Detail Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE-01
* **Description:** Get specific group profile info.
* **Detailed Steps:**
  1. Create `internal/group/queries/get_group.go`.
* **Verification:** Tests asserting responses structures.

---

### S4-BE-12: Group: Get Group Feed Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE-01
* **Description:** Retrieve post list inside group. Enforce membership check.
* **Detailed Steps:**
  1. Create `internal/group/queries/get_group_feed.go`.
* **Verification:** Assert non-members cannot read feed details.

---

### S4-BE-13: Group: Get Group Chat History Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S4-BE-01
* **Description:** Get group message history log.
* **Detailed Steps:**
  1. Create `internal/group/queries/get_group_chat.go`.
* **Verification:** Retrieve log chronologically.

---

### S4-BE-14: Group: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE-03..13
* **Description:** Bind HTTP routes. Every command and query must have at least one route.
* **Detailed Steps:**
   1. Create `internal/group/transport/http.go`.
   2. Route:
      - `POST /api/groups` — create_group (S4-BE-03)
      - `GET /api/groups` — list_groups (S4-BE-10)
      - `GET /api/groups/:id` — get_group (S4-BE-11)
      - `GET /api/groups/:id/feed` — get_group_feed (S4-BE-12)
      - `GET /api/groups/:id/chat` — get_group_chat (S4-BE-13)
      - `POST /api/groups/:id/invite` — invite_member (S4-BE-04)
      - `POST /api/groups/:id/invite/respond` — respond_invite (S4-BE-05)
      - `POST /api/groups/:id/join` — request_join (S4-BE-06)
      - `POST /api/groups/:id/join/respond` — respond_join (S4-BE-07)
      - `POST /api/groups/:id/posts` — create_group_post (S4-BE-08)
* **Verification:** Mock requests integration tests. Every command handler has a corresponding route.

---

### S4-BE-15: Group: WS Transport Routing
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE-14
* **Description:** Route real-time WS chat events.
* **Detailed Steps:**
  1. Create `internal/group/transport/ws.go`. Connect to core WebSocket.
* **Verification:** Test messaging over connections.

---

## BE-B (Backend B) Tickets

### S4-BE-16: Event: Entities & Repository Interface
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Establish domain structures for events and RSVP votes. Use separate `Option` entity per the architecture spec for extensibility (e.g. adding "maybe" later).
* **Detailed Steps:**
   1. Create `internal/event/event.go`. Define `Event` (ID, GroupID, CreatorID, Title, Description, ScheduledTime, CreatedAt), `Option` (EventID, Label string, e.g. "going", "not_going"), and `EventRSVP` (EventID, UserID, OptionID, UpdatedAt).
* **Verification:** Build checks.

---

### S4-BE-17: Event: SQLite Store
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S4-BE-16
* **Description:** Store query mappings for events. If `000006_events.up.sql` does not yet contain Event table DDL, create it here (or extend the existing migration).
* **Detailed Steps:**
   1. Create `internal/event/store/sqlite.go`.
   2. **If S1-BE-04 left `000006_events.up.sql` as a stub:** replace with full DDL: `CREATE TABLE events (...)`, `event_options`, `event_rsvps`. Provide corresponding `.down.sql`.
* **Verification:** Integration tests checking storage. Verify `000006` migration creates all Event tables.

---

### S4-BE-18: Event: Create Event Command
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S4-BE-16
* **Description:** Create event in group, executing group membership constraints, and publishing to notifications.
* **Detailed Steps:**
  1. Create `internal/event/commands/create_event.go`.
  2. Define local `GroupMemberChecker` interface. Enforce creator is group member.
  3. Validate fields: title, description, time in future, minimum of 2 RSVP options.
  4. Publish `event.created` event on EventBus.
* **Verification:** Unit tests verifying validator constraints, member checking, and notifications triggering.

---

### S4-BE-19: Event: RSVP Command
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S4-BE-16
* **Description:** Record RSVP choices (going/not going).
* **Detailed Steps:**
  1. Create `internal/event/commands/rsvp.go`. Upsert choices.
* **Verification:** Unit tests verifying updates to RSVP states.

---

### S4-BE-20: Event: List Group Events Query
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S4-BE-16
* **Description:** List events under a group with aggregated vote tallies.
* **Detailed Steps:**
  1. Create `internal/event/queries/list_group_events.go`.
* **Verification:** Test return outputs contain correct count aggregates.

---

### S4-BE-21: Event: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S4-BE-18..20
* **Description:** Bind HTTP handlers.
* **Detailed Steps:**
  1. Create `internal/event/transport/http.go`.
  2. Route `POST /api/groups/:id/events`, `GET /api/groups/:id/events`, `POST /api/events/:id/rsvp`.
* **Verification:** Integration test endpoints mapping.

---

## SD-QA (System Design/QA) Tickets

### S4-FE-08: E2E: Complete Groups Workspace Journey
* **Priority:** P0
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** Playwright testing group cycles.
* **Detailed Steps:**
  1. Script: Create Group -> Join members -> Post message -> Create Event -> Vote RSVP.
* **Verification:** Execution completes successfully in CI.

---

## FE-A (Frontend A) Tickets

### S4-FE-01: Groups Directory Page
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Implement browse view lists and search filters.
* **Detailed Steps:**
  1. Create `/groups` path fetching all items. Add search text filter input.
* **Verification:** Test listing render.

---

### S4-FE-02: Group Profile Page
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 5
* **Description:** Render detailed group info, member listings, and Join/Invite actions.
* **Detailed Steps:**
  1. Create `/groups/[id]` path. Show tabs: Feed, Events, Chat, Members.
* **Verification:** Interactive buttons clicks verification.

---

### S4-FE-03: Group Posts Feed
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Implement publishing post inside group space.
* **Detailed Steps:**
  1. Add post submission layout that sends group parameters.
* **Verification:** Submissions test verification.

---

### S4-FE-04: Group Chat Workspace
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 5
* **Description:** Build WebSocket messaging frame under group tabs.
* **Detailed Steps:**
  1. Open websocket channel, map incoming group payloads, display typing events.
* **Verification:** Check socket routing parameters.

---

## FE-B (Frontend B) Tickets

### S4-FE-05: Event Creation Dialog
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 5
* **Description:** Render event creator box inside group.
* **Detailed Steps:**
  1. Inputs for Title, Description, Date/Time picker, and RSVP option toggles.
* **Verification:** Form validates correct inputs before postings.

---

### S4-FE-06: Events List Component
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Display group events detailing description, time and current RSVP numbers.
* **Detailed Steps:**
  1. Render lists displaying stats counters.
* **Verification:** Verify counters load correctly.

---

### S4-FE-07: RSVP Switch Actions
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 2
* **Description:** Instant updates when voting RSVP.
* **Detailed Steps:**
  1. Click going/not going button updates choice immediately via websocket/REST call.
* **Verification:** Unit tests verifying update tallies visual checks.
