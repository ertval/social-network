# PICLE — Plan for Iterative Codebase Evolution

> Social Network — Full Spec Compliance Roadmap
> Derived from: readme.md (requirements), audit.md (audit checklist), AGENTS.md (spec compliance), codebase exploration

---

## Phase 0: Architecture & Foundation Fixes

> High-signal, low-risk structural issues and bugs. Prerequisite for feature work.

### 0.1 Migration delimiter: `":"` → `";"`
- **File**: `internal/infra/storage/sqlite/init.go:118`
- **Fix**: Change `strings.SplitSeq(string(content), ":")` to `strings.SplitSeq(string(content), ";")`
- **Risk**: Ensure no existing SQL contains an unprotected `";"` in strings (schema.sql uses no semicolons inside strings — safe)
- **Test**: Run server, verify migrations apply without error

### 0.2 Add `_busy_timeout=5000` to SQLite DSN
- **Files**: `.env`, `docker-compose.yml`, `internal/infra/storage/sqlite/init.go`
- **Fix**: Append `&_busy_timeout=5000` to the DSN pragma
- **Side effect**: Prevents `database is locked` errors under WAL contention

### 0.3 Fix OAuth repo `Scan(ctx)` bug
- **File**: `internal/infra/storage/sqlite/oauth/oauthRepo.go:182`
- **Fix**: Find the `Scan(ctx, ...)` call and replace `ctx` with the actual target variable
- **Test**: Run OAuth flow or unit test the repository method

### 0.4 Fix WebSocket `CheckOrigin` — remove dangerous wildcard
- **File**: `internal/infra/http/ws/handler.go`
- **Fix**: Replace `return true` with origin check against allowed origins from config or against the `Origin` header matching `Host`
- **Risk**: WS connections during dev may break if origin mismatch — allow configurable origins

### 0.5 Fix `order`/`orderBy` SQL injection vector
- **File**: `internal/infra/storage/sqlite/topics/topicRepo.go:414-420`
- **Fix**: Validate `order` against `["ASC", "DESC"]` whitelist, validate `orderBy` against known column whitelist
- **Test**: Verify topic listing with various sort params, ensure invalid params are rejected

### 0.6 Fix `userRepo.go` prepared statement — use `stmt.ExecContext`
- **File**: `internal/infra/storage/sqlite/users/userRepo.go:70-76`
- **Fix**: Change `r.DB.ExecContext(ctx, ...)` to `stmt.ExecContext(ctx, ...)`
- **Test**: Register a user and verify it persists correctly

### 0.7 Add `recover()` to WebSocket ReadPump/WritePump goroutines
- **File**: `internal/infra/ws/client.go`
- **Fix**: Wrap goroutine bodies in `defer func() { recover() }()` with log
- **Test**: Simulate panic inside WS handler, verify process survives

### 0.8 Fix RateLimiter cleanup goroutine leak
- **File**: `internal/infra/middleware/ratelimiter/rateLimiter.go`
- **Fix**: Add `stop chan struct{}` channel, select on ticker and stop in cleanup loop
- **Test**: Verify goroutine count doesn't grow after rate limiter creation

---

## Phase 1: Schema & Migration System Overhaul

> Database schema changes that underpin all new features. Requires migration system rework.

### 1.1 Migration system: numbered up/down format
- **Create**: `db/migrations/` with numbered files:
  - `000001_create_users.up.sql` / `000001_create_users.down.sql`
  - `000002_create_sessions.up.sql` / `000002_create_sessions.down.sql`
  - ...one per table for all current tables
- **Update**: `init.go` `migrateDB()` to read numbered files in order, track applied versions
- **Note**: Use `golang-migrate` or a custom tracker table (`schema_migrations`)

### 1.2 Add missing columns to existing tables
- **`users` table**:
  - `date_of_birth DATE NOT NULL` (replace `age INTEGER`)
  - `about_me TEXT DEFAULT ''`
  - `is_private BOOLEAN DEFAULT 0`
  - Keep `age INTEGER` initially, migrate data, then drop in down migration (or just add DOB — drop age later)
- **`topics` table**:
  - `visibility TEXT DEFAULT 'public' CHECK(visibility IN ('public','almost_private','private'))`
- **`direct_chats` table**: ensure user_low_id != user_high_id constraint

### 1.3 Create new tables for missing features

**follows**:
```sql
CREATE TABLE IF NOT EXISTS follows (
    follower_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followee_id),
    CHECK (follower_id != followee_id)
);
```

**follow_requests**:
```sql
CREATE TABLE IF NOT EXISTS follow_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','accepted','declined')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(sender_id, receiver_id)
);
```

**groups**:
```sql
CREATE TABLE IF NOT EXISTS groups (
    id TEXT PRIMARY KEY,
    creator_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**group_members**:
```sql
CREATE TABLE IF NOT EXISTS group_members (
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('creator','admin','member')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id)
);
```

**group_invitations**:
```sql
CREATE TABLE IF NOT EXISTS group_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    inviter_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','accepted','declined')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**group_join_requests**:
```sql
CREATE TABLE IF NOT EXISTS group_join_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    requester_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','accepted','declined')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, requester_id)
);
```

**events**:
```sql
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    creator_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    event_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**event_rsvps**:
```sql
CREATE TABLE IF NOT EXISTS event_rsvps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    response TEXT NOT NULL CHECK(response IN ('going','not_going')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, user_id)
);
```

### 1.4 Extend notification types
- **No schema change** — notification types are strings stored in `type` column
- Add handlers that create notifications with types: `follow_request`, `follow_request_accepted`, `group_invite`, `group_join_request`, `event_created`

---

## Phase 2: Backend Domain + App Layer (New Features)

> Domain models, repository interfaces, and application services for all new features.

### 2.1 Followers domain (`internal/domain/follow/`)

**`follow.go`**:
```go
type Follow struct {
    FollowerID string
    FolloweeID string
    CreatedAt  time.Time
}

type FollowRequest struct {
    ID         int64
    SenderID   string
    ReceiverID string
    Status     string // pending, accepted, declined
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

**`repository.go`** — interface:
- `CreateFollow(ctx, followerID, followeeID) error`
- `DeleteFollow(ctx, followerID, followeeID) error`
- `GetFollowers(ctx, userID) ([]*Follow, error)`
- `GetFollowing(ctx, userID) ([]*Follow, error)`
- `IsFollowing(ctx, followerID, followeeID) (bool, error)`
- `CreateFollowRequest(ctx, senderID, receiverID) error`
- `GetFollowRequest(ctx, senderID, receiverID) (*FollowRequest, error)`
- `GetPendingFollowRequests(ctx, userID) ([]*FollowRequest, error)`
- `UpdateFollowRequest(ctx, id, status) error`
- `GetFollowRequestByID(ctx, id) (*FollowRequest, error)`

### 2.2 Groups domain (`internal/domain/group/`)

**`group.go`**: `Group`, `GroupMember`, `GroupInvitation`, `GroupJoinRequest` models

**`repository.go`**: interfaces for CRUD + membership + invitations + join requests

### 2.3 Events domain (`internal/domain/event/`)

**`event.go`**: `Event`, `EventRSVP` models

**`repository.go`**: interfaces for CRUD + RSVP management

### 2.4 Application services

**`internal/app/follow/`**:
- `commands/followUser.go`, `commands/unfollowUser.go`
- `commands/sendFollowRequest.go`, `commands/respondToFollowRequest.go`
- `queries/getFollowers.go`, `queries/getFollowing.go`, `queries/getFollowRequests.go`

**`internal/app/group/`**:
- `commands/createGroup.go`, `commands/updateGroup.go`, `commands/deleteGroup.go`
- `commands/inviteToGroup.go`, `commands/respondToInvitation.go`
- `commands/requestJoinGroup.go`, `commands/respondToJoinRequest.go`
- `commands/leaveGroup.go`, `commands/removeMember.go`
- `queries/getGroup.go`, `queries/getUserGroups.go`, `queries/getGroupMembers.go`
- `queries/getGroupPosts.go`, `queries/searchGroups.go`

**`internal/app/event/`**:
- `commands/createEvent.go`, `commands/updateEvent.go`, `commands/deleteEvent.go`
- `commands/rsvpEvent.go`
- `queries/getGroupEvents.go`, `queries/getEventRSVPs.go`

### 2.5 Update `app/services.go`
- Wire all new commands/queries into the `Services` struct

### 2.6 Update user domain
- Add `DateOfBirth time.Time`, `AboutMe string`, `IsPrivate bool` to `internal/domain/user/user.go`
- Update `UserRegister` command to accept new fields
- Add `UpdateProfile` command (privacy toggle, bio, avatar)

### 2.7 Update topic domain
- Add `Visibility string` to `internal/domain/topic/topic.go`
- Update topic create/update to handle visibility + selected followers for `private` visibility

---

## Phase 3: Backend Infra Layer (New Features)

> SQLite repositories and HTTP handlers for all new features.

### 3.1 Follow repository (`internal/infra/storage/sqlite/follows/followRepo.go`)
- Implement `domain/follow/repository.go` interface
- Queries for follows CRUD + follow_requests CRUD
- Notifications created as side effects in `respondToFollowRequest`

### 3.2 Group repository (`internal/infra/storage/sqlite/groups/groupRepo.go`)
- Implement `domain/group/repository.go` interface
- Queries for groups, members, invitations, join requests

### 3.3 Event repository (`internal/infra/storage/sqlite/events/eventRepo.go`)
- Implement `domain/event/repository.go` interface
- Queries for events + RSVPs

### 3.4 Wire repositories
- Update `internal/infra/storage/sqlite/repositories.go` to include new repos

### 3.5 HTTP handlers — Follows

| Method | Route | Handler |
|--------|-------|---------|
| POST | `/api/v1/follow/:id` | Send follow request (if private) or follow directly (if public) |
| DELETE | `/api/v1/follow/:id` | Unfollow |
| GET | `/api/v1/follow/requests` | List incoming follow requests |
| PUT | `/api/v1/follow/requests/:id` | Accept/decline follow request |
| GET | `/api/v1/users/:id/followers` | Get followers of user |
| GET | `/api/v1/users/:id/following` | Get who user follows |

### 3.6 HTTP handlers — Groups

| Method | Route | Handler |
|--------|-------|---------|
| POST | `/api/v1/groups` | Create group |
| GET | `/api/v1/groups` | List/search groups |
| GET | `/api/v1/groups/:id` | Get group details |
| PUT | `/api/v1/groups/:id` | Update group (creator only) |
| DELETE | `/api/v1/groups/:id` | Delete group (creator only) |
| GET | `/api/v1/groups/:id/members` | List members |
| POST | `/api/v1/groups/:id/invite` | Invite user |
| GET | `/api/v1/groups/:id/invitations` | List invitations |
| PUT | `/api/v1/groups/:id/invitations/:inviteId` | Accept/decline invitation |
| POST | `/api/v1/groups/:id/join` | Request to join |
| GET | `/api/v1/groups/:id/join-requests` | List join requests |
| PUT | `/api/v1/groups/:id/join-requests/:reqId` | Accept/decline join request |
| DELETE | `/api/v1/groups/:id/members/:userId` | Remove member / leave |
| POST | `/api/v1/groups/:id/posts` | Create group post (with visibility=group) |
| GET | `/api/v1/groups/:id/posts` | List group posts |

### 3.7 HTTP handlers — Events

| Method | Route | Handler |
|--------|-------|---------|
| POST | `/api/v1/groups/:id/events` | Create event |
| GET | `/api/v1/groups/:id/events` | List group events |
| GET | `/api/v1/groups/:id/events/:eventId` | Get event details |
| PUT | `/api/v1/groups/:id/events/:eventId` | Update event |
| DELETE | `/api/v1/groups/:id/events/:eventId` | Delete event |
| POST | `/api/v1/events/:eventId/rsvp` | RSVP (going/not_going) |

### 3.8 HTTP handlers — Profile

| Method | Route | Handler |
|--------|-------|---------|
| GET | `/api/v1/users/:id` | Get user profile (respects privacy) |
| PUT | `/api/v1/users/:id` | Update profile (own only) |
| PUT | `/api/v1/users/:id/privacy` | Toggle profile privacy |

### 3.9 Update topic handlers
- Add `visibility` and `allowed_users` to topic create/update
- Filter topic listing by visibility (public for all, almost_private for followers, private for selected)
- Add endpoint: `GET /api/v1/topics/:id/allowed-users`

### 3.10 Update chat init
- In `internal/app/chat/commands/initChat.go`, add follow-relationship check:
  - `isFollowing(a, b) || isFollowing(b, a)` must be true, OR one has a public profile
- Return error `"cannot chat: no follow relationship"` if not allowed

### 3.11 Notification integration
- New notification types and triggers:
  - `follow_request` → when user sends follow request to private profile
  - `follow_request_accepted` → when follow request is accepted
  - `group_invite` → when invited to group
  - `group_join_request` → when someone requests to join your group
  - `event_created` → when event is created in a group you're a member of

### 3.12 Add image MIME validation
- **File**: `internal/app/topics/filestorage.go` or storage implementation
- **Fix**: Before saving, read first 512 bytes, call `http.DetectContentType()`, validate against `["image/jpeg", "image/png", "image/gif"]`
- Also add avatar upload endpoint + validation

### 3.13 Add `/api/v1/users/update` endpoint for profile editing
- Support avatar upload, about_me, date_of_birth, nickname, privacy toggle
- Move avatar upload to user handler (currently only topic has image upload)

---

## Phase 4: Frontend — Vanilla JS to Vue.js Migration

> The spec requires one of: Next.js, Vue, Svelte, Mithril. Migrating to Vue.js is recommended for pragmatic alignment with the existing SPA-like architecture.

### 4.1 Bootstrap Vue.js project
- Install via `npm create vue@latest` in a new `frontend-vue/` directory
- Set up Vite as build tool
- Configure Vue Router, Pinia for state management
- Set up proxy to backend BFF for development

### 4.2 Migrate existing features to Vue
- **Auth**: Login/register/logout forms with validation, OAuth callback
- **Home**: Category cards overview
- **Topics**: List with search/sort/filter/pagination
- **Topic Detail**: Full view with comments, voting, image
- **Create Post**: Form with multi-category, image upload, visibility selector
- **Activity**: User's posts/comments/votes activity stream
- **Account Settings**: Profile editing, avatar upload, OAuth linking
- **Notifications**: Real-time notification badge (SSE stream)
- **Chat**: WebSocket-based messaging panel (global/layout-level)
- **Navbar**: User menu, notification badge, logout

### 4.3 Build new features in Vue
- **Profile Page**: User profile with followers/following display, privacy toggle, post list
- **Follow UI**: Follow/unfollow buttons, follow-request accept/decline
- **Groups**: Create/list/join groups, membership management, invitations
- **Group Posts**: Posts within groups (separate feed)
- **Events**: Create events with title/description/datetime, RSVP buttons
- **Group Chat**: Chat room per group
- **Post Visibility**: Public/Almost Private/Private selector in post creation
- **Select Followers**: Modal to pick allowed viewers for private posts

### 4.4 Update registration form
- Add `date_of_birth` (date picker), `avatar` (file upload), `about_me` (textarea)
- Make `nickname` optional (as spec requires)
- Remove `gender` (not in spec — keep if desired but not required)
- Remove `age` (replaced by DOB)

### 4.5 Update frontend Docker setup
- Dockerfile for Vue.js SPA served via nginx (production build)
- Or serve via the existing Go BFF during migration period

---

## Phase 5: Docker & Deployment Rework

> Two-container architecture per spec requirement.

### 5.1 Split docker-compose into 2 services
- **`backend`**: Go API server on port 8080
- **`frontend`**: Vue.js SPA served via nginx on port 3000 (or the existing Go client on 3001)
- Update `docker-compose.yml` to define two services
- Remove the multi-binary entrypoint.sh approach

### 5.2 Create frontend Dockerfile
- **Production**: Multi-stage with `node:alpine` build + `nginx:alpine` runtime
- **Development**: Hot-reload with Vite dev server on port 5173

### 5.3 Fix entrypoint/healthcheck
- Entrypoint per-service (clean startup per container)
- Healthcheck for backend: `http://localhost:8080/api/v1/health`
- Healthcheck for frontend: `http://localhost:3000/`

### 5.4 Network configuration
- Both services on same bridge network
- Frontend proxies API calls to `http://backend:8080`

---

## Phase 6: Quality Assurance

> Testing, linting, and verification.

### 6.1 Backend tests for new features
- Follow repository tests (CRUD, request workflow)
- Group repository tests (CRUD, invitations, join requests)
- Event repository tests (CRUD, RSVP)
- Follow service tests (send request, accept, follow public)
- Group service tests (create, invite, accept, join request)
- Event service tests (create, RSVP)
- Chat init follow-check test
- Topic visibility filter tests

### 6.2 Run full test suite
```bash
go mod verify
go vet ./...
~/go/bin/golangci-lint run
govulncheck ./...
go test -race -coverprofile=coverage.out ./...
```

### 6.3 Audit verification checklist
- [ ] Followers table exists (follows, follow_requests)
- [ ] Profile privacy column (is_private on users)
- [ ] Post privacy column (visibility enum on topics)
- [ ] Groups table (with members, invitations, join_requests)
- [ ] Events table (with RSVP)
- [ ] Notification types include: follow-request, group-invite, group-join, event-creation
- [ ] Registration form: Email, Password, First Name, Last Name, Date of Birth, Avatar (opt), Nickname (opt), About Me (opt)
- [ ] Chat init checks follow relationship
- [ ] WebSocket upgrade checks auth token (✓ already done)
- [ ] docker-compose has 2 services (backend + frontend)
- [ ] Migrations follow numbered up/down format
- [ ] SQLite DSN: `_journal_mode=WAL` + `_busy_timeout=5000`
- [ ] SQL queries use `?` placeholders (✓ already done)
- [ ] Password hashing uses bcrypt (✓ already done)

---

## Dependency Graph (Implementation Order)

```
Phase 0 ──────────────────────────────────────────────┐
        │                                              │
        ▼                                              │
Phase 1: Schema + Migrations ─────────────────────────┤
        │                                              │
        ▼                                              │
Phase 2: Domain + App Layer (new features) ───────────┤
        │                                              │
        ▼                                              │
Phase 3: Infra Layer (repos + handlers) ──────────────┤
        │                                              │
        ▼                                              ▼
Phase 4: Frontend (Vue.js) ←───────────────── Phase 3 complete
    
        │                                              │
        ▼                                              │
Phase 5: Docker (2 services) ─────────────────────────┤
        │                                              │
        ▼                                              │
Phase 6: QA + Audit ──────────────────────────────────┘
```

**Parallelizable work**:
- Phase 0 items 0.1–0.8 can be done in parallel (independent bug fixes)
- Phase 4.1 (Vue scaffolding) can start early, independent of backend work
- Phase 5 can start after Phase 4 is functional

**Dependencies**:
- Phase 2 depends on Phase 1 (tables must exist)
- Phase 3 depends on Phase 2 (repos implement domain interfaces)
- Phase 4.2–4.4 depends on Phase 3 (API endpoints must exist)
- Phase 4.5 depends on Phase 5 (Docker split)
- Phase 6 is last (testing all features)

---

## Risk Register

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Vanilla JS → Vue migration is large | Schedule slip | Medium | Build incrementally, feature-flag old SPA |
| SQLite schema changes break existing data | Data loss | Low | Back up DB before migration, test with copy |
| Group posts intersect with existing topic system | Design complexity | Medium | Use same `topics` table with `group_id` FK + `visibility='group'` |
| OAuth flow breaks with new user fields | Auth regression | Low | Add new fields as nullable, test OAuth register |
| Docker 2-service migration breaks CI/CD | Deployment failure | Medium | Test locally first, keep old compose as fallback |
