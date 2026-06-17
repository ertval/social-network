# Sprint 4: Group & Event Features (Week 9–10)

**Outcome:** Groups with membership, group feed, group chat via WebSocket, and the event RSVP voting system work end-to-end.

---

## Backend Track — Group (`internal/group/`)

### S4-BE-01: Group: Entities & Repository Interface
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Establish domain model entities mapping group lifecycles.
* **Detailed Steps:**
  1. Create `internal/group/group.go`.
  2. Define `Group` (ID, Title, Description, CreatorID, CreatedAt), `GroupMember` (GroupID, UserID, Role: owner/member, JoinedAt), `Invitation` (GroupID, InviterID, InviteeID, Status: pending/accepted/declined, CreatedAt), and `JoinRequest` (GroupID, RequesterID, Status, CreatedAt).
* **Verification:** Compile check `go build ./internal/group/...`.

---

### S4-BE-02: Group: SQLite Store
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S4-BE-01
* **Description:** Implement SQLite storage mapping group structure.
* **Detailed Steps:**
  1. Create `internal/group/store/sqlite.go`. Implement queries checking group membership: `IsMember(ctx context.Context, groupID, userID string) (bool, error)`.
* **Verification:** Integration tests checking memberships writing.

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
* **Story Points:** 2
* **Dependencies:** S4-BE-01
* **Description:** Invite follower to group, firing event notifications.
* **Detailed Steps:**
  1. Create `internal/group/commands/invite_member.go`. Ensure requester is group member.
  2. Insert invitation, publish `group.invited` event.
* **Verification:** Unit tests verifying constraints and event outputs.

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
* **Description:** Bind HTTP routes.
* **Detailed Steps:**
  1. Create `internal/group/transport/http.go`.
  2. Route `POST /api/groups`, `GET /api/groups`, `POST /api/groups/:id/invite`, `POST /api/groups/:id/join`.
* **Verification:** Mock requests integration tests.

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

## Backend Track — Event (`internal/event/`)

### S4-BE-16: Event: Entities & Repository Interface
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Establish domain structures for events and RSVP votes.
* **Detailed Steps:**
  1. Create `internal/event/event.go`. Define `Event` (ID, GroupID, CreatorID, Title, Description, ScheduledTime, CreatedAt) and `EventRSVP` (EventID, UserID, Option: going/not_going, UpdatedAt).
* **Verification:** Build checks.

---

### S4-BE-17: Event: SQLite Store
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S4-BE-16
* **Description:** Store query mappings for events.
* **Detailed Steps:**
  1. Create `internal/event/store/sqlite.go`.
* **Verification:** Integration tests checking storage.

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

## Frontend Track

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

---

### S4-FE-08: E2E: Complete Groups Workspace Journey
* **Priority:** P0
* **Assignee:** FE-A + FE-B
* **Story Points:** 3
* **Description:** Playwright testing group cycles.
* **Detailed Steps:**
  1. Script: Create Group -> Join members -> Post message -> Create Event -> Vote RSVP.
* **Verification:** Execution completes successfully in CI.
