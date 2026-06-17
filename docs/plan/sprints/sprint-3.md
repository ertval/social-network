# Sprint 3: Follow, Comment & Notification (Week 7–8)

**Outcome:** Social relationships (follow requests, accepts, lists), commenting capability with media validation, and the event-driven notification dispatch pipeline work end-to-end.

---

## Backend Track — Follow (`internal/follow/`)

### S3-BE-01: Follow: Entities & Repository Interface
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Define domain entity shapes for follower links and pending follow requests.
* **Detailed Steps:**
  1. Create `internal/follow/follow.go`.
  2. Define `Follow` (FollowerID, FolloweeID, CreatedAt) and `FollowRequest` (SenderID, ReceiverID, Status: pending/accepted/declined, CreatedAt).
  3. Define the `Repository` interface mapping required storage operations.
* **Verification:** Compile check `go build ./internal/follow/...`.

---

### S3-BE-02: Follow: SQLite Store
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S3-BE-01
* **Description:** Implement storage operations for relationships in SQLite.
* **Detailed Steps:**
  1. Create `internal/follow/store/sqlite.go`.
  2. Implement relationship inserts, removals, and lookups using standard SQL queries.
* **Verification:** Store integration tests using in-memory SQLite connections checking requests creation and link resolutions.

---

### S3-BE-03: Follow: Follow User Command
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S3-BE-01
* **Description:** Initiate relationship link. Perform auto-follow for public profiles, and follow request creation for private ones.
* **Detailed Steps:**
  1. Create `internal/follow/commands/follow_user.go`.
  2. Define a local `UserPrivacyChecker` interface to inspect if target profile is private.
  3. If public -> insert direct relationship and publish `follow.accepted` event to `platform/eventbus`.
  4. If private -> insert record to `follow_requests` and publish `follow.requested` event.
* **Verification:** Unit tests verifying: instant follow for public user, and request creation for private user.

---

### S3-BE-04: Follow: Unfollow User Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-01
* **Description:** Remove relationship links.
* **Detailed Steps:**
  1. Create `internal/follow/commands/unfollow_user.go`. Delete relationship records.
* **Verification:** Unit tests verifying that unfollow severs connection in database.

---

### S3-BE-05: Follow: Accept Request Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-03
* **Description:** Approve pending follow requests.
* **Detailed Steps:**
  1. Create `internal/follow/commands/accept_request.go`.
  2. Update status of request to accepted, insert direct relationship link, and publish `follow.accepted` event.
* **Verification:** Unit tests verifying acceptance state updates and event emissions.

---

### S3-BE-06: Follow: Decline Request Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 1
* **Dependencies:** S3-BE-03
* **Description:** Decline pending follow requests.
* **Detailed Steps:**
  1. Create `internal/follow/commands/decline_request.go`. Set request status to declined or delete record.
* **Verification:** Unit tests verifying decline actions.

---

### S3-BE-07: Follow: Get Followers Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-01
* **Description:** Retrieve follower listing.
* **Detailed Steps:**
  1. Create `internal/follow/queries/get_followers.go`.
* **Verification:** Test querying listing for mock user relationships.

---

### S3-BE-08: Follow: Get Following Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-01
* **Description:** Retrieve users followed listing.
* **Detailed Steps:**
  1. Create `internal/follow/queries/get_following.go`.
* **Verification:** Test querying correct listing.

---

### S3-BE-09: Follow: Get Pending Requests Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-01
* **Description:** Retrieve pending incoming follow requests.
* **Detailed Steps:**
  1. Create `internal/follow/queries/get_pending_requests.go`.
* **Verification:** Test retrieves all pending requests for receiver.

---

### S3-BE-10: Follow: Are Connected Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-01
* **Description:** Implement helper query check for cross-slice connection testing.
* **Detailed Steps:**
  1. Create `internal/follow/queries/are_connected.go`.
  2. Satisfies `FollowChecker` interface used by other slices.
* **Verification:** Test check returns accurate boolean evaluations.

---

### S3-BE-11: Follow: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S3-BE-03..10
* **Description:** Bind follow actions to HTTP paths.
* **Detailed Steps:**
  1. Create `internal/follow/transport/http.go`.
  2. Map `POST /api/follow`, `POST /api/unfollow`, `POST /api/follow/accept`, `POST /api/follow/decline`, `GET /api/follow/requests`.
* **Verification:** Integration tests checking handler requests.

---

### S3-BE-12: Follow: Event Publishing Verification
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-03..06
* **Description:** Verify follow events are published onto the event bus correctly.
* **Detailed Steps:**
  1. Write tests subscribing mock event listeners to `follow.requested` and `follow.accepted` topics.
  2. Trigger commands and assert listeners receive expected structures.
* **Verification:** Tests pass without event drops.

---

## Backend Track — Comment & Notification

### S3-BE-13: Comment: Entity & Repository Interface
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 1
* **Description:** Setup Domain model for post comments.
* **Detailed Steps:**
  1. Create `internal/comment/comment.go`.
  2. Define `Comment` entity (ID, TopicID, AuthorID, Content, ImagePath, CreatedAt) and `Repository` interface.
* **Verification:** Verify package compilation.

---

### S3-BE-14: Comment: SQLite Store
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S3-BE-13
* **Description:** Implement SQLite CRUD queries for comments.
* **Detailed Steps:**
  1. Create `internal/comment/store/sqlite.go`.
* **Verification:** Integration tests checking comments insertion and query by TopicID.

---

### S3-BE-15: Comment: Create Comment Command
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-13
* **Description:** Write-use-case for creating comment with image attachments.
* **Detailed Steps:**
  1. Create `internal/comment/commands/create_comment.go`.
  2. Validate image MIME type (JPG, PNG, GIF) using `pkg/imgutil` if file upload is present.
  3. Validate content length bounds.
* **Verification:** Unit tests verifying valid inputs, image MIME types checking, and content validations.

---

### S3-BE-16: Comment: Get Comments Query
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S3-BE-13
* **Description:** Retrieve list of comments for a specific post.
* **Detailed Steps:**
  1. Create `internal/comment/queries/get_comments.go`.
* **Verification:** Test correct ordering of queries.

---

### S3-BE-17: Comment: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S3-BE-15, S3-BE-16
* **Description:** Bind HTTP handlers.
* **Detailed Steps:**
  1. Create `internal/comment/transport/http.go`.
  2. Route `POST /api/posts/:id/comments`, `GET /api/posts/:id/comments`.
* **Verification:** Integration tests testing json payload returns.

---

### S3-BE-18: Comment Slice: Contract Tests
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S3-BE-17
* **Description:** Ensure comments vertical slice compatibility with old domain.
* **Detailed Steps:**
  1. Create `internal/comment/store/sqlite_migration_test.go`.
* **Verification:** Assert matching outputs.

---

### S3-BE-19: Notification: Entity & Repository Interface
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Define Notification domain structures.
* **Detailed Steps:**
  1. Create `internal/notification/notification.go`.
  2. Define `Notification` entity (ID, ReceiverID, Type, SenderID, ResourceID, IsRead, CreatedAt).
  3. Define standard notification types: `follow_request`, `follow_accept`, `group_invite`, `group_join_request`, `event_created`.
* **Verification:** Build checks.

---

### S3-BE-20: Notification: SQLite Store
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S3-BE-19
* **Description:** Implement notification write/update SQLite actions.
* **Detailed Steps:**
  1. Create `internal/notification/store/sqlite.go`.
* **Verification:** Tests checking inserts and updating `is_read` parameters.

---

### S3-BE-21: Notification: Event Bus Consumer
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 5
* **Dependencies:** S3-BE-19
* **Description:** Event subscriber handler mapping events to database notifications.
* **Detailed Steps:**
  1. Create `internal/notification/commands/consume_events.go`.
  2. Subscribe to event types: `follow.requested`, `follow.accepted`, `group.invited`, `group.join_requested`, `event.created`.
  3. When an event fires, insert a record into the notifications table for the target user.
* **Verification:** Write event delivery test, triggering an event on the bus and inspecting database notifications update.

---

### S3-BE-22: Notification: Mark Read Command
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 1
* **Dependencies:** S3-BE-19
* **Description:** Update notifications state to read.
* **Detailed Steps:**
  1. Create `internal/notification/commands/mark_read.go`.
* **Verification:** Unit tests asserting `is_read` flag transitions to true.

---

### S3-BE-23: Notification: List Notifications Query
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S3-BE-19
* **Description:** Retrieve notification feed for active user.
* **Detailed Steps:**
  1. Create `internal/notification/queries/list_notifications.go`.
* **Verification:** Query matches correct chronological ordering.

---

### S3-BE-24: Notification: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S3-BE-21..23
* **Description:** Bind HTTP routes.
* **Detailed Steps:**
  1. Create `internal/notification/transport/http.go`.
  2. Route `GET /api/notifications`, `POST /api/notifications/:id/read`.
* **Verification:** HTTP calls assert.

---

## Frontend Track

### S3-FE-01: Follow Button with Popup
* **Priority:** P0
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Interactive toggle button with popup confirmation before unfollowing a user.
* **Detailed Steps:**
  1. If following -> click triggers Dialog popup confirming "Are you sure you want to unfollow?".
  2. Confirm -> posts to `/api/unfollow`.
  3. If not following -> click posts to `/api/follow`.
* **Verification:** Vitest asserting click behavior, state tracking, and API calls trigger.

---

### S3-FE-02: Followers List Pages
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Build lists of followers and following links on the profile view.
* **Detailed Steps:**
  1. Create sub-views showing user cards with action buttons.
* **Verification:** Assert visual rendering correctness.

---

### S3-FE-03: Follow Request Notifications
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Render inline accept/decline action items for follow requests.
* **Detailed Steps:**
  1. List item displaying user name, avatar, and buttons to Accept or Decline.
* **Verification:** Confirm accept triggers target state.

---

### S3-FE-04: Comment Section Components
* **Priority:** P0
* **Assignee:** FE-B
* **Story Points:** 5
* **Description:** Build comment listing and text entry component featuring image attachment upload.
* **Detailed Steps:**
  1. Fetch comments under target post cards.
  2. Form with file selector allowing JPG/PNG/GIF upload checking.
* **Verification:** Visual validation and Playwright submission testing.

---

### S3-FE-05: Notifications Panel
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Build UI panel displaying unread count badges in navigation bars.
* **Detailed Steps:**
  1. Badge icon displaying count. Clicking drops down notification listing drawer.
* **Verification:** Verify badge matches response totals.

---

### S3-FE-06: Notifications Live Stream
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Establish real-time notifications loading utilizing SSE or WebSockets.
* **Detailed Steps:**
  1. Connect to websocket hub / realtime server channels.
  2. Dispatch incoming event messages to update global notification badge count dynamically.
* **Verification:** Trigger mock WebSocket messages and check panel response.

---

### S3-FE-07: E2E: Relationships Notifications Flow
* **Priority:** P0
* **Assignee:** FE-A + FE-B
* **Story Points:** 3
* **Description:** E2E testing of relationship workflows.
* **Detailed Steps:**
  1. User A follows User B (private) -> B receives notification -> B accepts -> relationship established.
* **Verification:** Running Playwright test passes.

---

### S3-FE-08: E2E: Posts Comments Notification Flow
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 2
* **Description:** E2E testing of commenting actions.
* **Detailed Steps:**
  1. User A creates post -> User B comments -> verify comments listing updating.
* **Verification:** Running Playwright test passes.
