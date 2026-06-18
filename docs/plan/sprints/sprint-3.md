# Sprint 3: Follow, Comment & Notification (Week 7–8)

**Outcome:** Social relationships (follow requests, accepts, lists), commenting capability with media validation, and the event-driven notification dispatch pipeline work end-to-end.

> **Migration note:** New slices use `/api/` prefix. Old code uses `/api/v1/`. During Strangler Fig migration, both must coexist. Register new routes alongside old ones — old code stays active until Sprint 6 cleanup. Feature-flag the new routes behind a config toggle if needed. Old notification types (reply, mention, like, dislike) are replaced entirely by new types (follow_request, follow_accept, group_invite, group_join_request, event_created). Existing notification data with old types is kept for history but not migrated to new types.
>
> **Live notifications via Server-Sent Events (SSE):** To ensure a live and premium user experience, real-time notifications are delivered using Server-Sent Events (SSE) via `GET /api/notifications/stream` (implemented in S3-BE-57). The frontend (S3-FE-18) establishes an `EventSource` connection to this stream to render live alerts without manual polling, falling back to 15s interval polling only if SSE is unsupported or disconnected.
>
> **Old notification data migration:** S3-BE-58 migrates old rows (Title, Message, RelatedType) to the new schema (Type, SenderID, ResourceID).

---

### S3-BE-59: Wire Follow, Comment & Notification bootstrap routes
* **Priority:** P0
* **Assignee:** BE-A + BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-45, S3-BE-50, S3-BE-57
* **Description:** Register new slice routes in `bootstrap.go` so endpoints are live immediately after this sprint.
* **Detailed Steps:**
  1. In `internal/bootstrap/bootstrap.go`, import follow, comment, and notification transport packages.
  2. Call their route registration functions on the HTTP mux.
  3. Register event bus consumers for notification events.
* **Verification:** `go build ./...` passes, new endpoints respond 200/401 (not 404).

---

## BE-A (Backend A) Tickets

### S3-BE-35: Follow: Entities & Repository Interface
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Define domain entity shapes for follower links and pending follow requests.
* **Detailed Steps:**
  1. Create `internal/follow/follow.go`.
  2. Define `Follow` (FollowerID, FolloweeID, CreatedAt) and `FollowRequest` (FollowerID, FolloweeID, CreatedAt). **Note:** Remove the redundant `Status` field and use `FollowerID` and `FolloweeID` to align with the database schema which has no `status` column (row presence denotes pending).
  3. Define the `Repository` interface mapping required storage operations.
* **Verification:** Compile check `go build ./internal/follow/...`.

---

### S3-BE-36: Follow: SQLite Store
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S3-BE-35
* **Description:** Implement storage operations for relationships in SQLite.
* **Detailed Steps:**
  1. Create `internal/follow/store/sqlite.go`.
  2. Implement relationship inserts, removals, and lookups using standard SQL queries on `follows` and `follow_requests` tables.
* **Verification:** Store integration tests using in-memory SQLite connections checking requests creation and link resolutions.

---

### S3-BE-37: Follow: Follow User Command
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S3-BE-35
* **Description:** Initiate relationship link. Perform auto-follow for public profiles, and follow request creation for private ones.
* **Detailed Steps:**
  1. Create `internal/follow/commands/follow_user.go`.
  2. Define a local `UserPrivacyChecker` interface to inspect if target profile is private.
  3. If public -> insert direct relationship to `follows` and publish `follow.accepted` event to `platform/eventbus`.
  4. If private -> insert record to `follow_requests` and publish `follow.requested` event.
* **Verification:** Unit tests verifying: instant follow for public user, and request creation for private user.

---

### S3-BE-38: Follow: Unfollow User Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-35
* **Description:** Remove relationship links.
* **Detailed Steps:**
  1. Create `internal/follow/commands/unfollow_user.go`. Delete relationship records.
* **Verification:** Unit tests verifying that unfollow severs connection in database.

---

### S3-BE-39: Follow: Accept Request Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-37
* **Description:** Approve pending follow requests.
* **Detailed Steps:**
  1. Create `internal/follow/commands/accept_request.go`.
  2. Delete follow request row from `follow_requests`, insert direct relationship link to `follows`, and publish `follow.accepted` event.
* **Verification:** Unit tests verifying acceptance state updates, row removals, and event emissions.

---

### S3-BE-40: Follow: Decline Request Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 1
* **Dependencies:** S3-BE-37
* **Description:** Decline pending follow requests.
* **Detailed Steps:**
  1. Create `internal/follow/commands/decline_request.go`. Delete request record from `follow_requests`.
* **Verification:** Unit tests verifying decline actions and row deletion.

---

### S3-BE-41: Follow: Get Followers Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-35
* **Description:** Retrieve follower listing.
* **Detailed Steps:**
  1. Create `internal/follow/queries/get_followers.go`.
* **Verification:** Test querying listing for mock user relationships.

---

### S3-BE-42: Follow: Get Following Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-35
* **Description:** Retrieve users followed listing.
* **Detailed Steps:**
  1. Create `internal/follow/queries/get_following.go`.
* **Verification:** Test querying correct listing.

---

### S3-BE-43: Follow: Get Pending Requests Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-35
* **Description:** Retrieve pending incoming follow requests.
* **Detailed Steps:**
  1. Create `internal/follow/queries/get_pending_requests.go`.
* **Verification:** Test retrieves all pending requests for receiver.

---

### S3-BE-44: Follow: Are Connected Query
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S3-BE-35
* **Description:** Implement helper query check for cross-slice connection testing.
* **Detailed Steps:**
  1. Create `internal/follow/queries/are_connected.go`. Query if users are followed in the database.
* **Verification:** Tests returning true for connected, false for unconnected profiles.

---

### S3-BE-45: Follow: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S3-BE-37..10
* **Description:** Bind HTTP REST handlers.
* **Detailed Steps:**
  1. Create `internal/follow/transport/http.go`.
  2. Route `POST /api/follow`, `POST /api/unfollow`, `POST /api/follow/accept`, `POST /api/follow/decline`, `GET /api/users/:id/followers`, `GET /api/users/:id/following`, `GET /api/follow/requests`.
* **Verification:** Mock HTTP handlers test integration.

---

## BE-B (Backend B) Tickets

### S3-BE-46: Comment: Entity & Repository Interface
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Define domain entity shapes for post comments.
* **Detailed Steps:**
  1. Create `internal/comment/comment.go`.
  2. Define `Comment` entity: ID (int), TopicID (int), UserID (string) [maps to user_id in DB], Content, ImagePath, and CreatedAt. Use `int` for IDs to match legacy SQLite auto-increment schemas.
  3. Define the `Repository` interface mapping required storage operations.
* **Verification:** Compile check `go build ./internal/comment/...`.

---

### S3-BE-47: Comment: SQLite Store
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-46
* **Description:** Implement SQLite storage for comment records.
* **Detailed Steps:**
  1. Create `internal/comment/store/sqlite.go`. Implement queries saving comments and loading comments by TopicID.
* **Verification:** Store integration tests using in-memory SQLite connections checking writes/reads.

---

### S3-BE-48: Comment: Create Comment Command
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-46
* **Description:** Create comment record with file attachments and MIME validation.
* **Detailed Steps:**
  1. Create `internal/comment/commands/create_comment.go`.
  2. Validate parameters. Enforce magic-byte image checks (JPG, PNG, GIF) on uploaded attachments using `pkg/imgutil`.
* **Verification:** Unit tests verifying commenting, invalid image format rejection, and database insertions.

---

### S3-BE-49: Comment: Get Comments Query
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S3-BE-46
* **Description:** Retrieve comment list for a specific post.
* **Detailed Steps:**
  1. Create `internal/comment/queries/get_comments.go`.
* **Verification:** Test querying correct mapping.

---

### S3-BE-50: Comment: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-48, S3-BE-49
* **Description:** Setup HTTP routing handlers for comment endpoints.
* **Detailed Steps:**
   1. Create `internal/comment/transport/http.go`.
   2. Wire up `POST /api/comments`, `GET /api/posts/:id/comments`.
* **Verification:** HTTP mock integration tests verifying correct endpoint codes.

---

### S3-BE-52: Notification: Entity & Repository Interface
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Define domain model entities mapping notifications.
* **Detailed Steps:**
  1. Create `internal/notification/notification.go`.
  2. Define `Notification` (ID, UserID, Type, SourceID, Content, IsRead, CreatedAt) and `Repository` interface.
* **Verification:** Compile checks.

---

### S3-BE-53: Notification: SQLite Store
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-52
* **Description:** Create SQLite storage queries mapping notifications.
* **Detailed Steps:**
  1. Create `internal/notification/store/sqlite.go`.
* **Verification:** Store integration tests checks.

---

### S3-BE-54: Notification: Event Bus Consumer
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-52
* **Description:** Subscribe to Event Bus events and compile user notification alerts.
* **Detailed Steps:**
  1. Create `internal/notification/commands/consume_events.go`.
  2. Subscribe to: `follow.requested`, `follow.accepted`, `group.invited`, `group.join_requested`, `event.created`.
  3. Create and store a new notification row in SQLite for each incoming event.
* **Verification:** Mock publishers test check.

---

### S3-BE-55: Notification: Mark Read Command
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 1
* **Dependencies:** S3-BE-52
* **Description:** Mark notifications as read.
* **Detailed Steps:**
  1. Create `internal/notification/commands/mark_read.go`. Flip `is_read` boolean field in database.
* **Verification:** Unit tests asserting that updates succeed.

---

### S3-BE-56: Notification: List Notifications Query
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S3-BE-52
* **Description:** Retrieve notifications history listing.
* **Detailed Steps:**
  1. Create `internal/notification/queries/list_notifications.go`.
* **Verification:** Test querying listing.

---

### S3-BE-57: Notification: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-54..23
* **Description:** Bind HTTP routes. Add unread count endpoint and a real-time SSE stream endpoint.
* **Detailed Steps:**
   1. Create `internal/notification/transport/http.go`.
   2. Route `GET /api/notifications` (lists all notifications), `GET /api/notifications/unread-count` (returns unread count), `POST /api/notifications/:id/read`.
   3. **Server-Sent Events Stream:** Route `GET /api/notifications/stream`. Establish a persistent HTTP stream chunking notifications in real-time when new eventbus alerts fire.
* **Verification:** HTTP calls assert.

---

### S3-BE-58: Notification: Old Schema→New Schema Migration
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-52
* **Description:** One-time data migration converting old notification rows to the new schema.
* **Detailed Steps:**
   1. Create `db/migrations/000008_migrate_notifications.up.sql`. Map old rows: `Title` → stored as metadata, `Message` → stored as metadata, `RelatedType` → mapped to new `Type` enum where possible (reply→comment, mention→follow, like/dislike→dropped). Rows with unmappable types get `Type = "legacy"`.
   2. Create `000008_migrate_notifications.down.sql` to reverse.
   3. Old rows keep their IDs; new notifications use the new schema going forward.
* **Verification:** Run migration up/down. Verify old rows are converted without data loss.

---

### S3-BE-51: Comment: Cast Vote Command & Queries (Gap Fix)
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S3-BE-46
* **Description:** Implement upvoting/downvoting on comments by leveraging the unified votes storage schema.
* **Detailed Steps:**
  1. Create `internal/comment/commands/cast_comment_vote.go` (similar logic to post voting: checks duplicate, inserts/updates `comment_id` link in the `votes` table).
  2. Create `internal/comment/queries/get_comment_votes.go` to sum and query voting directions for comment card displays.
  3. Map routing to `POST /api/comments/:id/vote`.
* **Verification:** Unit tests verifying upvoting and downvoting comments.

---

## FE-A (Frontend A) Tickets

### S3-FE-13: Follow Button with Popup
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

### S3-FE-14: Followers List Pages
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Build lists of followers and following links on the profile view.
* **Detailed Steps:**
  1. Create sub-views showing user cards with action buttons.
* **Verification:** Assert visual rendering correctness.

---

### S3-FE-15: Follow Request Notifications
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Render inline accept/decline action items for follow requests.
* **Detailed Steps:**
  1. List item displaying user name, avatar, and buttons to Accept or Decline.
* **Verification:** Confirm accept triggers target state.

---

## FE-B (Frontend B) Tickets

### S3-FE-16: Comment Section Components
* **Priority:** P0
* **Assignee:** FE-B
* **Story Points:** 5
* **Description:** Build comment listing and text entry component featuring image attachment upload.
* **Detailed Steps:**
  1. Fetch comments under target post cards.
  2. Form with file selector allowing JPG/PNG/GIF upload checking.
* **Verification:** Visual validation and Playwright submission testing.

---

### S3-FE-17: Notifications Panel
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Build UI panel displaying unread count badges in navigation bars.
* **Detailed Steps:**
  1. Badge icon displaying count. Clicking drops down notification listing drawer.
* **Verification:** Verify badge matches response totals.

---

### S3-FE-18: Notifications Live Stream
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Implement Server-Sent Events (SSE) notification streaming connection.
* **Detailed Steps:**
   1. Establish connection to `GET /api/notifications/stream` using the browser `EventSource` API.
   2. When events arrive, update the global notification badge state and append notifications to the panel view.
   3. Gracefully fall back to polling `GET /api/notifications/unread-count` on a 15-second interval if the SSE connection drops or fails.
* **Verification:** Start server, trigger notification via another user, confirm badge count updates instantly without page reload.

---

### S3-FE-19: Comment Card Vote Buttons (Gap Fix)
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 2
* **Dependencies:** S3-FE-16
* **Description:** Implement upvote/downvote action items on comments.
* **Detailed Steps:**
  1. Render buttons on comment card components. Trigger POST calls to `/api/comments/:id/vote` and dynamically update local vote tally state.
* **Verification:** Verify interactive click increments tally.

---

## SD-QA (System Design/QA) Tickets

### S3-SD-11: Follow: Event Publishing Verification
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** S3-BE-37..06
* **Description:** Verify follow events are published onto the event bus correctly.
* **Detailed Steps:**
  1. Write tests subscribing mock event listeners to `follow.requested` and `follow.accepted` topics.
  2. Trigger commands and assert listeners receive expected structures.
* **Verification:** Tests pass without event drops.

---

### S3-SD-12: Comment Slice: Contract Tests
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** S3-BE-50
* **Description:** Ensure comments vertical slice compatibility with old domain.
* **Detailed Steps:**
  1. Create `internal/comment/store/sqlite_migration_test.go`.
* **Verification:** Assert matching outputs.

---

### S3-SD-13: Platform: Follow System Migrations (000004)
* **Priority:** P0
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** S1-BE-06
* **Description:** Create the database migration files for the Follow vertical slice.
* **Detailed Steps:**
  1. Create `db/migrations/000004_follow_system.up.sql` to create `follows` and `follow_requests` tables.
  2. Create `db/migrations/000004_follow_system.down.sql` to reverse these changes.
* **Verification:** Run `make db-reset` or execute the migration runner and verify that this migration applies and rolls back cleanly.

---

### S3-SD-14: E2E: Relationships Notifications Flow
* **Priority:** P0
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** E2E testing of relationship workflows.
* **Detailed Steps:**
  1. User A follows User B (private) -> B receives notification -> B accepts -> relationship established.
* **Verification:** Running Playwright test passes.

---

### S3-SD-15: E2E: Posts Comments Notification Flow
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 2
* **Description:** E2E testing of commenting actions.
* **Detailed Steps:**
  1. User A creates post -> User B comments -> verify comments listing updating.
* **Verification:** Running Playwright test passes.
