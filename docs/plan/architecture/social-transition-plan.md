# Social Network — Integrated Evolution & Compliance Plan

This document outlines the unified, deduplicated, and optimized roadmap to evolve the codebase into a fully compliant, production-ready, and secure Facebook-like social network. It merges the requirements, codebase guidelines, and audit checklists from `PICLE.md`, `plan-ds.md`, and `plan-flash.md`.

---

## Target Architecture

```
                      +-------------------+
                      |      Browser      |
                      |   (Next.js App)   |
                      +---------+---------+
                                | (HTTP/WS)
                                v
                      +-------------------+
                      |    Go Backend     |
                      |    (Port 8080)    |
                      +---------+---------+
                                |
                                v
                      +-------------------+
                      |   SQLite Database |
                      |    (WAL Enabled)  |
                      +-------------------+
```

- **Frontend**: A modern **Next.js** web application replacing the current Vanilla JS SPA. It will leverage Next.js App Router, using vanilla CSS for premium, custom styling (gradients, glassmorphism, smooth animations), and handle client-side WebSocket/SSE updates.
- **Backend**: Clean Architecture with CQRS commands/queries (`domain` -> `app` -> `infra`). We preserve the separation of HTTP handlers, CQRS commands/queries, and SQLite repositories.
- **Database**: Support through Factory pattern for PostgreSQL and SQLite with Write-Ahead Logging (`WAL`), `_busy_timeout=5000`, and parameterized queries.

---

## Phase 0: Critical Bug & Architectural Fixes

*Prerequisite tasks targeting high-signal, low-risk structural issues and bugs before starting feature development.*

### 0.1 SQL Migration Delimiter Correction
- **File**: `internal/infra/storage/sqlite/init.go`
- **Fix**: Change `strings.SplitSeq(string(content), ":")` to `strings.SplitSeq(string(content), ";")` to avoid parsing errors on schemas/timestamps containing colons.

### 0.2 SQLite DSN Configuration & Busy Timeout
- **Files**: `.env`, `docker-compose.yml`, `internal/infra/storage/sqlite/init.go`
- **Fix**: Ensure the SQLite connection DSN includes `_journal_mode=WAL` and `_busy_timeout=5000` to prevent database locks under concurrent write workloads.

### 0.3 Fix OAuth Repository Scan Bug
- **File**: `internal/infra/storage/sqlite/oauth/oauthRepo.go`
- **Fix**: Correct the `Scan()` signature on the DB query, removing the context (`ctx`) argument from the destination parameters to prevent runtime crashes during OAuth flows.

### 0.4 Harden WebSocket Origin Checking
- **File**: `internal/infra/http/ws/handler.go`
- **Fix**: Replace the unsafe `return true` check inside `CheckOrigin` with a validation check verifying that the request host matches the configured frontend origin.

### 0.5 Prevent SQL Injection in Ordering Whitelist
- **Files**: `internal/infra/storage/sqlite/topics/topicRepo.go`, `internal/infra/storage/sqlite/categories/categoryRepo.go`
- **Fix**: Validate the `order` parameter against `["ASC", "DESC"]` and target columns against an explicit field whitelist before interpolating them into SQL statements.

### 0.6 Fix Prepared Statement Execution
- **File**: `internal/infra/storage/sqlite/users/userRepo.go`
- **Fix**: Replace `r.DB.ExecContext` calls using prepared statement queries with direct `stmt.ExecContext` calls on the active prepared statement resource to ensure correct query execution.

### 0.7 WebSocket Panic Recovery
- **File**: `internal/infra/ws/client.go`
- **Fix**: Wrap `ReadPump` and `WritePump` goroutine loops in a `defer recover()` block to log errors and prevent user-level WebSocket panics from crashing the entire server process.

### 0.8 RateLimiter Ticker Cleanup Goroutine Leak
- **File**: `internal/infra/middleware/ratelimiter/rateLimiter.go`
- **Fix**: Add a `stop chan struct{}` field to the rate limiter and handle cleanup channel selection inside the ticker loop to properly dispose of goroutines when the service terminates.

---

## Phase 1: Database Migration System & Schema Overhaul

*Transitioning the database architecture to support new tables and column additions using numbered migrations.*

### 1.1 Migration Tracker System
- Update `init.go` to support a new migrations runner. The system should read files sequentially from `db/migrations/` and track applied versions in a `schema_migrations` table.
- Use the standard numbered format: `000001_initial_schema.up.sql`, `000001_initial_schema.down.sql`.

### 1.2 Alter Existing Tables
- **`users` Table**:
  - Add `date_of_birth DATE NOT NULL` (replaces `age` integer, migrate existing data if needed, then drop `age`).
  - Add `about_me TEXT DEFAULT ''`.
  - Add `is_private BOOLEAN DEFAULT 0`.
- **`topics` Table**:
  - Add `visibility TEXT DEFAULT 'public' CHECK(visibility IN ('public', 'almost_private', 'private'))`.

### 1.3 Create New Schema Tables
Create SQL scripts inside `db/migrations/` to establish:

- **`follows`**:
  ```sql
  CREATE TABLE IF NOT EXISTS follows (
      follower_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      followee_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (follower_id, followee_id),
      CHECK (follower_id != followee_id)
  );
  ```
- **`follow_requests`**:
  ```sql
  CREATE TABLE IF NOT EXISTS follow_requests (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      receiver_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'declined')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      UNIQUE(sender_id, receiver_id)
  );
  ```
- **`groups`**:
  ```sql
  CREATE TABLE IF NOT EXISTS groups (
      id TEXT PRIMARY KEY,
      creator_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      description TEXT DEFAULT '',
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```
- **`group_members`**:
  ```sql
  CREATE TABLE IF NOT EXISTS group_members (
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      role TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('creator', 'admin', 'member')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (group_id, user_id)
  );
  ```
- **`group_invitations`**:
  ```sql
  CREATE TABLE IF NOT EXISTS group_invitations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      inviter_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      invitee_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'declined')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```
- **`group_join_requests`**:
  ```sql
  CREATE TABLE IF NOT EXISTS group_join_requests (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      requester_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'declined')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      UNIQUE(group_id, requester_id)
  );
  ```
- **`events`**:
  ```sql
  CREATE TABLE IF NOT EXISTS events (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      creator_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      description TEXT DEFAULT '',
      event_time TIMESTAMP NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```
- **`event_rsvps`**:
  ```sql
  CREATE TABLE IF NOT EXISTS event_rsvps (
      event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      response TEXT NOT NULL CHECK(response IN ('going', 'not_going')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (event_id, user_id)
  );
  ```
- **`topic_allowed_users`** (Post Privacy Target Whitelist):
  ```sql
  CREATE TABLE IF NOT EXISTS topic_allowed_users (
      topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      PRIMARY KEY (topic_id, user_id)
  );
  ```
- **`group_chat_messages`**:
  ```sql
  CREATE TABLE IF NOT EXISTS group_chat_messages (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      content TEXT NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```

---

## Phase 2: Backend Domain & Application Services

*Defining interfaces, domain models, and CQRS commands/queries.*

### 2.1 Domain Layer Updates
- **User Domain (`internal/domain/user/user.go`)**: Update `User` entity to contain `DateOfBirth time.Time`, `AboutMe string`, and `IsPrivate bool`. Make Nickname field handling optional.
- **Topic Domain (`internal/domain/topic/topic.go`)**: Add `Visibility` field (constants: `Public`, `AlmostPrivate`, `Private`) and list of `AllowedUsers []string`.
- **Follow Domain (`internal/domain/follow/follow.go`)**:
  - Define `Follow` and `FollowRequest` structs.
  - Define `Repository` interface: `CreateFollow`, `DeleteFollow`, `GetFollowers`, `GetFollowing`, `IsFollowing`, `CreateFollowRequest`, `GetPendingFollowRequests`, `UpdateFollowRequest`.
- **Group Domain (`internal/domain/group/group.go`)**:
  - Define `Group`, `GroupMember`, `GroupInvitation`, and `GroupJoinRequest` structs.
  - Define `Repository` interface for groups CRUD, membership updates, invitations, and join requests.
- **Event Domain (`internal/domain/event/event.go`)**:
  - Define `Event` and `EventRSVP` structs.
  - Define `Repository` interface for events creation and RSVP operations.

### 2.2 CQRS Application Layer
- **Follow Commands/Queries (`internal/app/follow/`)**:
  - `commands/followUser.go` / `unfollowUser.go`
  - `commands/sendFollowRequest.go` / `respondToFollowRequest.go`
  - `queries/getFollowers.go` / `getFollowing.go` / `getFollowRequests.go`
- **Group Commands/Queries (`internal/app/group/`)**:
  - `commands/createGroup.go` / `updateGroup.go` / `deleteGroup.go`
  - `commands/inviteToGroup.go` / `respondToInvitation.go`
  - `commands/requestJoinGroup.go` / `respondToJoinRequest.go`
  - `commands/leaveGroup.go`
  - `queries/getGroup.go` / `getUserGroups.go` / `getGroupMembers.go` / `getGroupPosts.go` / `searchGroups.go`
- **Event Commands/Queries (`internal/app/event/`)**:
  - `commands/createEvent.go` / `rsvpEvent.go`
  - `queries/getGroupEvents.go` / `getEventRSVPs.go`
- **Services Registry (`internal/app/services.go`)**:
  - Register all new CQRS structures within the orchestrating `Services` object.

---

## Phase 3: Backend Infrastructure Layer

*Implementing database access logic, router setup, and API handlers.*

### 3.1 SQLite Repositories
Implement the repository interfaces using SQLite:
- `internal/infra/storage/sqlite/follows/followRepo.go`
- `internal/infra/storage/sqlite/groups/groupRepo.go`
- `internal/infra/storage/sqlite/events/eventRepo.go`
- Register the new repositories inside `internal/infra/storage/sqlite/repositories.go`.

### 3.2 HTTP Handlers & API Routes

#### Users / Profiles
- `GET /api/v1/users/:id` — Get profile information. Respects profile privacy (returns minimal profile summary if private and requester is not a follower).
- `PUT /api/v1/users/update` — Update avatar, biography, date of birth, nickname, or toggle privacy settings.
- `GET /api/v1/users/:id/activity` — Get post and comment history of the target user. Restricts output based on profile privacy.

#### Follow System
- `POST /api/v1/follow/:id` — Sends follow request (if profile is private) or registers direct follow (if public).
- `DELETE /api/v1/follow/:id` — Unfollows a user.
- `GET /api/v1/follow/requests` — Lists pending follow requests for the authenticated user.
- `PUT /api/v1/follow/requests/:id` — Accept or decline a follow request.
- `GET /api/v1/users/:id/followers` — List of user's followers.
- `GET /api/v1/users/:id/following` — List of users the target follows.

#### Group Management
- `POST /api/v1/groups` — Create a group.
- `GET /api/v1/groups` — Search and browse groups.
- `GET /api/v1/groups/:id` — Retrieve group details.
- `DELETE /api/v1/groups/:id/members/:userId` — Leave group or remove a member.
- `POST /api/v1/groups/:id/invite` — Invite users.
- `GET /api/v1/groups/:id/invitations` — List active invitations.
- `PUT /api/v1/groups/:id/invitations/:inviteId` — Accept/decline an invite.
- `POST /api/v1/groups/:id/join` — Request to join.
- `GET /api/v1/groups/:id/join-requests` — List active requests.
- `PUT /api/v1/groups/:id/join-requests/:reqId` — Accept/decline join request.

#### Group Content
- `POST /api/v1/groups/:id/posts` — Create a group post (saved in topics table with `group_id`).
- `GET /api/v1/groups/:id/posts` — Retrieve group posts feed (accessible by members only).
- `POST /api/v1/groups/:id/events` — Create a group event.
- `GET /api/v1/groups/:id/events` — List events scheduled for the group.
- `POST /api/v1/events/:eventId/rsvp` — RSVP (`going`, `not_going`).

#### Post Privacy Whitelist
- `GET /api/v1/topics/:id/allowed-users` — Retrieve target follower usernames allowed to view the private topic.

### 3.3 Domain Integrations & Adjustments
- **Chat Access Constraint**: Update `internal/app/chat/commands/initChat.go`. Users can only initialize a chat session if they follow each other (either follower/followee direction) or the recipient has a public profile.
- **Image Content Security**: Update `internal/app/topics/filestorage.go` and user registration/update profile routes. Read the first 512 bytes of uploads using `http.DetectContentType` to enforce true image type validation (JPEG, PNG, GIF) and prevent client MIME-spoofing.
- **Notification Dispatches**: Wire notification logic inside the services layer:
  - `follow_request` when sending a request to a private profile.
  - `follow_request_accepted` when follow request is accepted.
  - `group_invite` when a user is invited to a group.
  - `group_join_request` when a user requests to join a group.
  - `event_created` when an event is added to a group (dispatched to all members).
- **Google OAuth Wiring Check**: Verify that Google OAuth handlers in `internal/infra/http/server.go` correctly map to the Google authentication queries/services (e.g. `UserLoginGoogle` query service) rather than reusing the Github login handler explicitly if distinct parameters are required.

---

## Phase 4: Next.js Frontend Migration

*Replacing the Vanilla JS SPA with a high-fidelity Next.js application.*

### 4.1 Next.js Application Architecture
- Scaffold a Next.js application in `frontend/` (using App Router).
- Set up a clean directory structure:
  - `frontend/src/app` (routes and page structures).
  - `frontend/src/components` (reusable widgets: navbar, modals, buttons).
  - `frontend/src/styles` (global and modular Vanilla CSS).
- Styling & Aesthetics:
  - Keep standard Next.js default styles out. Apply a rich vanilla CSS aesthetic (glassmorphism overlays, custom dark modes, harmonious color palettes, micro-interactions, CSS custom transitions).
  - Typography: Implement Google Font loaders (e.g., Outfit or Inter).

### 4.2 Registration Form Updates
- Add `date_of_birth` datepicker and validate that the user is at least 13 years old.
- Add `about_me` biography textbox.
- Support optional `nickname` handling.
- Support dynamic avatar file uploading with instant image previews.

### 4.3 Post Creation Privacy Controls
- Integrate a dropdown to choose post visibility (`public`, `almost_private` [followers only], `private`).
- When `private` is selected, show a modal to search/filter active followers and select who is whitelisted.

### 4.4 Profile Page Features
- URL mapping: `/profile/[id]`.
- Public/Private profile selector with confirmation popups.
- Show follow/unfollow buttons. Unfollowing triggers a confirmation popup.
- Conditional feed rendering: If the profile is private and the current user is not following them, hide posts, comments, and followers/following lists. Show a lock screen placeholder instead.

### 4.5 Group & Event Views
- **/groups**: Create group modal and directory of all available groups.
- **/groups/[id]**: Group main portal:
  - Members directory.
  - Invitation sender and active join requests management dashboard (accessible by group owner).
  - Group posts feed (restricted to members).
  - Events scheduler card containing descriptions, event times, and RSVP switches (`Going` / `Not Going`).
  - Chat channel integration using WebSockets to send and fetch history in the group's chat room.

### 4.6 Global Components
- **Navbar**: Show notification counts dynamically.
- **Notifications Panel**: Render follow requests, group invites, and event notifications with quick action buttons (`Accept`/`Decline`) that execute immediate POST requests.
- **SSE/WebSocket Handler**: Maintain connection lifecycle to receive live notifications and feed events.

---

## Phase 5: Multi-Container Deployment

*Split the single-container execution model into isolated backend and frontend services.*

### 5.1 Docker Compose Configuration
- Expose **`backend`** service on port `8080` (Go rest api).
- Expose **`frontend`** service on port `3000` (Next.js server).
- Define backend environment variables inside the frontend service to routing calls cleanly.

### 5.2 Frontend Dockerfile
- Create `frontend/Dockerfile` for Next.js:
  - Multi-stage build process using Node.js image.
  - Installs npm dependencies, builds Next.js production files, and starts the Next.js server runner (`npm run start`).

### 5.3 Backend Dockerfile
- Create `Dockerfile` (or `Dockerfile.backend` in the root):
  - Multi-stage build compiling Go binaries into a minimal distroless or alpine container.
  - Mounts/attaches host volumes for SQLite data files to ensure database persistence.

---

## Phase 6: Quality Assurance & Checklist Verification

*Detailed procedures to confirm spec compliance, security parameters, and performance.*

### 6.1 Automated Verification
Run the following test procedures regularly:
- Go testing with race detection: `go test -race -coverprofile=coverage.out ./...`
- Code validation and syntax check: `go vet ./...`
- Lint checks (using local custom ruleset): `golangci-lint run --config scratch/golangci.yml`
- Vulnerability scanners: `govulncheck ./...`

### 6.2 Manual Test Scenarios

#### Scenario A: Registration & Login Validation
1. Register user under 13 years old -> Assert rejection.
2. Register user without Nickname and About Me -> Assert successful registration (nickname is optional).
3. Upload non-image text file renamed to `.png` as avatar -> Assert rejection.

#### Scenario B: Follow Workflow & Profile Locks
1. Set User B to private profile.
2. User A clicks follow B -> Assert follow request popup appears and B receives a follow request notification.
3. User B declines request -> Assert no relationship is formed.
4. User A attempts to view B's activity -> Assert page states "Profile is private" (no posts or follower lists visible).
5. User A clicks follow B again -> User B accepts -> Assert A can now view B's full activity and start a chat session.
6. User A clicks unfollow B -> Assert confirmation popup triggers. Confirm -> Assert relationship is severed and profile locks again.

#### Scenario C: Post Privacy Scopes
1. User A creates a post with visibility set to "Almost Private" (followers only).
2. User B (follower of A) visits A's profile or home feed -> Assert post is visible.
3. User C (not a follower of A) visits A's profile or home feed -> Assert post is invisible.
4. User A creates a post with visibility "Private" selecting only User B.
5. User B -> Assert post is visible.
6. User D (follower of A, but not selected) -> Assert post is invisible.

#### Scenario D: Group & Event Planner Lifecycle
1. User A creates a group.
2. User A sends group invitation to User B -> Assert User B receives group invite notification.
3. User B accepts -> B gains access to group chat room and posts.
4. User C requests to join the group -> Assert creator A receives join request notification.
5. User A accepts -> User C joins.
6. User A creates an event for the group scheduled next week -> Assert all members receive event creation notification.
7. User B RSVPs "Going" -> Assert count of attendees updates in real-time.
