# Social Network — Evolution & Compliance Plan

This document outlines the detailed roadmap to evolve the codebase into a fully compliant, production-ready, and secure Facebook-like social network. It addresses all functional and non-functional requirements from the specification (`readme.md`), the grading audit checklist (`audit.md`), and the codebase guidelines (`AGENTS.md`).

---

## Architecture Overview

```
                      +-------------------+
                      |      Browser      |
                      |   (Svelte SPA)    |
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

- **Frontend**: Migrating from the current Vanilla JS hand-rolled router to **Svelte**. Svelte compiles down to minimal vanilla-like JS, matching the project's light, reactive feel while satisfying the framework requirement.
- **Backend**: Clean Architecture with CQRS structure (`domain` -> `app` -> `infra`). We preserve the separation between the HTTP handlers, CQRS commands/queries, and storage repositories.
- **Database**: SQLite with Write-Ahead Logging (`WAL`), `_busy_timeout=5000`, and parameterized queries.

---

## User Review Required

> [!IMPORTANT]
> **1. JS Framework Choice (Svelte)**
> We recommend **Svelte** as the JS framework. It compiles components down to small, clean vanilla JavaScript, which aligns perfectly with the current SPA structure and minimizes build bloat. Svelte Router will replace the current hand-rolled router.
> 
> **2. Database Migrations Schema Overhaul**
> Evolving the schema requires a new numbered up/down migration format. We will create files like `000001_create_users_table.up.sql` to systematically build tables, instead of the flat `schema.sql` file.

---

## Open Questions

> [!NOTE]
> **1. Confirmation Popups Requirement**
> The audit checklist specifies: "+If you unfollow a user, do you get a confirmation pop-up?" and "+If you change your profile from public to private (or vice versa), do you get a confirmation pop-up?". These confirmation flows will be implemented in the frontend UI.
> 
> **2. Date of Birth Validation**
> The registration form must accept `Date of Birth` instead of the current `Age` integer. Users must be at least 13 years old. We will enforce this validation both in the frontend and the backend.

---

## Proposed Changes

### 1. Database & Migrations

#### [MODIFY] [sqlite/init.go](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/init.go)
- Append `&_busy_timeout=5000` to the SQLite connection DSN.
- Change the SQL execution statement splitting delimiter from `":"` to `";"` (line 118) to avoid parsing bugs with strings or timestamps containing colons.
- Update migration system to read numbered migrations from `internal/infra/storage/sqlite/migrations/` in up/down format.

#### [NEW] [000001_init_schema.up.sql](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/migrations/000001_init_schema.up.sql)
- Numbered SQL script containing core user, session, category, topic, comment, and notification tables.
- Updates `users` table to store `date_of_birth` (date type) and `about_me` text.

#### [NEW] [000002_follows.up.sql](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/migrations/000002_follows.up.sql)
- Schema for follows and follow requests:
  ```sql
  CREATE TABLE IF NOT EXISTS follows (
      follower_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      followee_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (follower_id, followee_id),
      CHECK (follower_id != followee_id)
  );
  CREATE TABLE IF NOT EXISTS follow_requests (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      receiver_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'declined')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      UNIQUE(sender_id, receiver_id)
  );
  ```

#### [NEW] [000003_groups_and_events.up.sql](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/migrations/000003_groups_and_events.up.sql)
- Schema for groups, memberships, invitations, events, RSVPs, and group chat messages:
  ```sql
  CREATE TABLE IF NOT EXISTS groups (
      id TEXT PRIMARY KEY,
      creator_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      description TEXT DEFAULT '',
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  CREATE TABLE IF NOT EXISTS group_members (
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      role TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('creator', 'admin', 'member')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (group_id, user_id)
  );
  CREATE TABLE IF NOT EXISTS group_invitations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      inviter_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      invitee_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'declined')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  CREATE TABLE IF NOT EXISTS group_join_requests (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      requester_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'declined')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      UNIQUE(group_id, requester_id)
  );
  CREATE TABLE IF NOT EXISTS events (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      creator_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      description TEXT DEFAULT '',
      event_time TIMESTAMP NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  CREATE TABLE IF NOT EXISTS event_rsvps (
      event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      response TEXT NOT NULL CHECK(response IN ('going', 'not_going')),
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      PRIMARY KEY (event_id, user_id)
  );
  CREATE TABLE IF NOT EXISTS group_chat_messages (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
      sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      content TEXT NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```

#### [NEW] [000004_post_privacy.up.sql](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/migrations/000004_post_privacy.up.sql)
- Mappings for private post targeted followers:
  ```sql
  CREATE TABLE IF NOT EXISTS topic_allowed_users (
      topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      PRIMARY KEY (topic_id, user_id)
  );
  ```

---

### 2. Backend Domain Models & App Services

#### [MODIFY] [user/user.go](file:///home/ertval/code/zone-modules/social-network/internal/domain/user/user.go)
- Add `DateOfBirth time.Time`, `AboutMe string`, and `IsPrivate bool` fields.

#### [MODIFY] [topic/topic.go](file:///home/ertval/code/zone-modules/social-network/internal/domain/topic/topic.go)
- Add `Visibility string` field with constants: `Public`, `AlmostPrivate`, and `Private`.

#### [NEW] [follow/follow.go](file:///home/ertval/code/zone-modules/social-network/internal/domain/follow/follow.go)
- Follow and FollowRequest models.
- Defines the `Repository` interface: `CreateFollow`, `DeleteFollow`, `GetFollowers`, `GetFollowing`, `IsFollowing`, `CreateFollowRequest`, `GetPendingFollowRequests`, `UpdateFollowRequest`.

#### [NEW] [group/group.go](file:///home/ertval/code/zone-modules/social-network/internal/domain/group/group.go)
- Group models: `Group`, `GroupMember`, `GroupInvitation`, `GroupJoinRequest`.
- Defines `Repository` interface for Group operations, including invitations, memberships, and join requests.

#### [NEW] [event/event.go](file:///home/ertval/code/zone-modules/social-network/internal/domain/event/event.go)
- Event and EventRSVP models.
- Defines the `Repository` interface for event creation and RSVPs.

#### [NEW] [follow/commands/](file:///home/ertval/code/zone-modules/social-network/internal/app/follow/commands/)
- Implement `SendFollowRequest`, `AcceptFollowRequest`, `DeclineFollowRequest`, and `Unfollow`.
- Triggers notifications in `AcceptFollowRequest` (type `follow_request_accepted`) and `SendFollowRequest` (type `follow_request`).

#### [NEW] [group/commands/](file:///home/ertval/code/zone-modules/social-network/internal/app/group/commands/)
- Implement `CreateGroup`, `InviteToGroup`, `AcceptGroupInvitation`, `RequestToJoinGroup`, `AcceptJoinRequest`, `LeaveGroup`.

#### [NEW] [event/commands/](file:///home/ertval/code/zone-modules/social-network/internal/app/event/commands/)
- Implement `CreateEvent` and `RsvpEvent`.
- `CreateEvent` notifies all group members of type `event_created`.

#### [MODIFY] [services.go](file:///home/ertval/code/zone-modules/social-network/internal/app/services.go)
- Register new follow, group, and event commands/queries into backend orchestrator wiring.

---

### 3. Infrastructure Layer (Repositories & Handlers)

#### [NEW] [follows/followRepo.go](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/follows/followRepo.go)
- Implements `domain/follow.Repository` interface with SQLite.

#### [NEW] [groups/groupRepo.go](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/groups/groupRepo.go)
- Implements `domain/group.Repository` interface with SQLite.

#### [NEW] [events/eventRepo.go](file:///home/ertval/code/zone-modules/social-network/internal/infra/storage/sqlite/events/eventRepo.go)
- Implements `domain/event.Repository` interface with SQLite.

#### [MODIFY] [ws/handler.go](file:///home/ertval/code/zone-modules/social-network/internal/infra/http/ws/handler.go)
- Harden `CheckOrigin` configuration to check matching `Host` headers rather than allowing all standard origins unconditionally (`return true`).

#### [MODIFY] [ws/client.go](file:///home/ertval/code/zone-modules/social-network/internal/infra/ws/client.go)
- Wrap `ReadPump` and `WritePump` loops in `defer func() { if r := recover(); r != nil { ... } }()` to prevent process crashes due to WebSocket panics.

#### [MODIFY] [ratelimiter/rateLimiter.go](file:///home/ertval/code/zone-modules/social-network/internal/infra/middleware/ratelimiter/rateLimiter.go)
- Add a `stop chan struct{}` field and select statements inside the `cleanup` ticker loop to prevent background goroutine leaks.

#### [MODIFY] [topics/filestorage.go](file:///home/ertval/code/zone-modules/social-network/internal/app/topics/filestorage.go)
- Add server-side magic byte MIME check (`http.DetectContentType`) to block disguised uploads.

---

### 4. Frontend (Vanilla JS SPA to Svelte Migration)

We will configure Svelte inside the `frontend/` directory to replace the hand-rolled templates and custom DOM routing.

#### [NEW] [App.svelte](file:///home/ertval/code/zone-modules/social-network/frontend/src/App.svelte)
- Main layout wrapper containing the navbar, notifications dropdown, chat sidebar, and main page content.

#### [NEW] [routes.js](file:///home/ertval/code/zone-modules/social-network/frontend/src/routes.js)
- Svelte-SPA-Router routes mapping paths to Svelte page views:
  - `/profile/:id`: conditional view (public profile vs private follow-only).
  - `/groups`: list all groups + create group modal.
  - `/group/:id`: single group details, including posts feed, event planner, and chat.
  - `/chat`: websocket chat dashboard.

#### [NEW] [Profile.svelte](file:///home/ertval/code/zone-modules/social-network/frontend/src/components/Profile.svelte)
- Renders profile card with user information, avatar, followers count, and follow activity.
- Profile privacy toggle with confirmation popup.
- Follow/unfollow buttons with confirmation popup for unfollows.

#### [NEW] [CreatePost.svelte](file:///home/ertval/code/zone-modules/social-network/frontend/src/components/CreatePost.svelte)
- Post editor supporting image upload and privacy selection:
  - Public
  - Almost Private (followers only)
  - Private (modal to search and choose specific followers).

---

### 5. Multi-Container Deployment

#### [MODIFY] [docker-compose.yml](file:///home/ertval/code/zone-modules/social-network/docker-compose.yml)
- Split into two isolated Docker services:
  - `backend`: Go REST API and WS Hub.
  - `frontend`: Svelte static build served via Nginx with reverse proxy to `http://backend:8080/api/`.

---

## Verification Plan

### Automated Tests
- `go test -race -coverprofile=coverage.out ./...`
- `go vet ./...`
- `$(go env GOPATH)/bin/golangci-lint run --config scratch/golangci.yml`
- `$(go env GOPATH)/bin/staticcheck ./...`

### Manual Verification
- Deploying using `docker-compose up --build`.
- Verify Registration: enter DOB, upload avatar, bio, verify Optional Nickname.
- Verify Follow flows:
  - Try following a private profile -> verify follow request notification received.
  - Accept request -> verify followers/following list updates.
  - Try unfollowing -> verify confirmation popup triggers.
- Verify Post privacy:
  - Create a private post selecting only User B -> verify User C cannot view the post.
- Verify Groups & Events:
  - Create group, send invitation -> verify invite notification.
  - Accept invite -> join group chat.
  - Create event, RSVP "Going" -> verify attendee count.
