# Social Network — Implementation Plan

## Architecture Overview

Current: `Browser (Vanilla JS SPA) <-> Go BFF (Port 3001) <-> Go API (Port 8080) <-> SQLite`

Target: Same architecture, with all missing spec features added. The codebase follows clean architecture (`domain/` → `app/` → `infra/`) with CQRS commands/queries. We preserve this.

---

## Phase 1: Critical Bug Fixes

### 1.1 Migration delimiter `":"` → `";"`
**File:** `internal/infra/storage/sqlite/init.go:118`
**Change:** `statements := strings.SplitSeq(string(content), ":")` → `statements := strings.SplitSeq(string(content), ";")`
**Risk:** None. Current SQL has no `;` inside strings. Verified safe.

### 1.2 `ctx` in `Scan()` args
**File:** `internal/infra/storage/sqlite/oauth/oauthRepo.go:182-183`
**Change:** Remove `ctx` from `Scan()` arg list. Pass correct pointers.
**Risk:** Runtime crash when `GetOAuthProvider()` is called.

### 1.3 WebSocket `CheckOrigin` hardening
**File:** `internal/infra/http/ws/handler.go`
**Change:** Replace `return true` with origin-whitelist check using configured frontend URL.
**Risk:** CSRF WebSocket vector.

### 1.4 Add `_busy_timeout=5000` to DSN
**File:** `config` (where DSN pragma is defined) + `internal/infra/storage/sqlite/init.go`
**Change:** Append `&_busy_timeout=5000` to pragma string.
**Risk:** Prevents "database is locked" errors under WAL contention.

### 1.5 Fix prepared statement in `userRepo.go`
**File:** `internal/infra/storage/sqlite/users/userRepo.go:70-76`
**Change:** Replace `r.DB.ExecContext(ctx, query, ...)` with `stmt.ExecContext(ctx, ...)`.
**Risk:** Performance only, no correctness bug.

### 1.6 Order-by SQL injection in `categoryRepo.go`
**File:** `internal/infra/storage/sqlite/categories/categoryRepo.go:68`
**Change:** Validate `order` (ASC/DESC) against `[]string{"ASC", "DESC"}` whitelist.
**Risk:** SQL injection in category listing.

### 1.7 RateLimiter goroutine leak
**File:** Wherever `go rl.cleanup()` runs without a stop channel.
**Change:** Add `stop chan struct{}` to ticker goroutine.
**Risk:** Permanent goroutine leak.

---

## Phase 2: Foundation Changes

### 2.1 Migration system upgrade
**Current:** Flat `schema.sql` + `indexes.sql` files with no tracking table, wrong delimiter.
**Target:** Numbered up/down migration files with `schema_migrations` tracking table.
- Create `000001_create_users_table.up.sql` / `.down.sql` from current schema (split into logical units)
- Create `000002_create_follows.up.sql` (follows + follow_requests)
- Create `000003_add_privacy_columns.up.sql` (is_private, date_of_birth, about_me, visibility)
- Create `000004_create_groups.up.sql` (groups, group_members, group_invitations)
- Create `000005_create_events.up.sql` (events, event_rsvps)
- Add `schema_migrations` tracking table
- Update `init.go` to use golang-migrate or similar

### 2.2 Domain model updates

#### `internal/domain/user/user.go`
- Replace `Age int` with `DateOfBirth time.Time`
- Add `AboutMe string`
- Add `IsPrivate bool`
- Add `Nickname string` (keep username as nickname, make truly optional)

#### `internal/domain/topic/topic.go` (or posts domain)
- Add `Visibility string` with constants: `public`, `almost_private`, `private`
- Add `AllowedUsers []string` for private posts

#### `internal/domain/notification/notification.go`
- Add notification types: `follow_request`, `follow_accept`, `group_invite`, `group_join_request`, `group_join_accepted`, `event_creation`

#### New domain: `internal/domain/follow/`
- `Follow` struct: ID, FollowerID, FolloweeID, Status (pending/accepted/rejected), CreatedAt
- `Repository` interface: Create, GetByID, UpdateStatus, GetFollowers, GetFollowing, IsFollowing, GetFollowRequests

#### New domain: `internal/domain/group/`
- `Group` struct: ID, Title, Description, CreatorID, AvatarURL, CreatedAt
- `GroupMember` struct: GroupID, UserID, Role (creator/admin/member), JoinedAt
- `GroupInvitation` struct: ID, GroupID, InviterID, InviteeID, Status, CreatedAt
- `GroupJoinRequest` struct: ID, GroupID, UserID, Status, CreatedAt
- `Repository` interface

#### New domain: `internal/domain/event/`
- `Event` struct: ID, GroupID, CreatorID, Title, Description, DateTime, CreatedAt
- `EventRSVP` struct: EventID, UserID, Option (going/not_going), CreatedAt
- `Repository` interface

### 2.3 Schema changes to `db/migrations/schema.sql` (and new migration files)

```sql
-- Add columns to users
ALTER TABLE users ADD COLUMN date_of_birth TEXT;
ALTER TABLE users ADD COLUMN about_me TEXT;
ALTER TABLE users ADD COLUMN is_private INTEGER DEFAULT 0;

-- Add visibility to topics
ALTER TABLE topics ADD COLUMN visibility TEXT DEFAULT 'public' CHECK(visibility IN ('public','almost_private','private'));

-- Create follows table
CREATE TABLE IF NOT EXISTS follows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    follower_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','accepted','rejected')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(follower_id, followee_id)
);

-- Create groups tables
CREATE TABLE IF NOT EXISTS groups (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    creator_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    avatar_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('creator','admin','member')),
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS group_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    inviter_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','accepted','rejected')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS group_join_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','accepted','rejected')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, user_id)
);

-- Create events tables
CREATE TABLE IF NOT EXISTS group_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    creator_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    event_time DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS event_rsvps (
    event_id INTEGER NOT NULL REFERENCES group_events(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    option TEXT NOT NULL CHECK(option IN ('going','not_going')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, user_id)
);
```

### 2.4 App layer (use cases)

New files in `internal/app/`:
- `commands/follow.go` — SendFollowRequest, AcceptFollowRequest, RejectFollowRequest, Unfollow
- `commands/group.go` — CreateGroup, InviteToGroup, AcceptInvitation, RejectInvitation, RequestToJoin, AcceptJoinRequest, RejectJoinRequest, LeaveGroup
- `commands/event.go` — CreateEvent, RSVPEvent
- `queries/follow.go` — GetFollowers, GetFollowing, GetFollowRequests, IsFollowing
- `queries/group.go` — GetGroup, GetUserGroups, GetGroupMembers, GetGroupInvitations, GetGroupJoinRequests, ListAllGroups
- `queries/event.go` — GetGroupEvents, GetEventRSVPs

---

## Phase 3: Followers & Profile Privacy

### 3.1 Domain
- Implement `internal/domain/follow/` with structs + repository interface
- Add `IsPrivate` to user domain model

### 3.2 Infrastructure
- `internal/infra/storage/sqlite/follows/followRepo.go` — implement `follow.Repository`
- Wire into bootstrap

### 3.3 HTTP handlers
- `POST /api/v1/follow/request` — send follow request
- `POST /api/v1/follow/accept` — accept follow request
- `POST /api/v1/follow/reject` — reject follow request
- `POST /api/v1/follow/unfollow` — unfollow
- `GET /api/v1/follow/followers?user_id=X` — get followers
- `GET /api/v1/follow/following?user_id=X` — get followees
- `GET /api/v1/follow/requests` — get pending follow requests for current user
- `GET /api/v1/follow/status?user_id=X` — check follow status between users

### 3.4 Profile handler
- Modify `/api/v1/user/activity` to:
  - Return `is_private` flag
  - If viewing user's own profile: return everything + option to toggle privacy
  - If viewing another user's profile:
    - Public profile: return all info
    - Private profile + viewer is follower: return all info
    - Private profile + viewer is NOT follower: return only basic info (no posts, no followers/following lists)
- Add `PUT /api/v1/user/privacy` — toggle profile privacy

### 3.5 Chat init check
- Modify `POST /api/v1/chat/init` to verify follow relationship exists
- Users can only chat if: one follows the other (either direction) OR recipient has public profile

### 3.6 Frontend: Profile pages
- Create `frontend/static/js/pages/profile.js` — user profile page route `/profile/:id`
- Show user info, posts, followers/following lists
- Privacy toggle on own profile
- Follow/unfollow buttons
- Display conditional content based on privacy/follow status
- Show follower/following counts

### 3.7 Frontend: Follow UI
- Add "Follow" / "Unfollow" button to user profile page
- For private profiles: shows "Request Sent" state
- Accept/Reject buttons for incoming follow requests (in notifications dropdown + dedicated page)
- Real-time notification via SSE when follow request arrives

---

## Phase 4: Post Privacy

### 4.1 Backend
- Add `visibility` column to topics table
- Add `allowed_users` table (topic_id, user_id) for private post targeting

```sql
CREATE TABLE IF NOT EXISTS topic_allowed_users (
    topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (topic_id, user_id)
);
```

- Modify topic create handler to accept `visibility` + `allowed_users[]`
- Modify topic retrieval to filter based on visibility:
  - Public: everyone can see
  - Almost private: only followers of post creator
  - Private: only specified allowed_users
- Implement in topic query/repository level

### 4.2 Frontend
- Add privacy selector to create post form (public / almost private / private)
- When "private" selected, show user search/multi-select for allowed users
- Filter display of posts on home page, category page, profile page based on visibility

---

## Phase 5: Registration Enhancement

### 5.1 Domain
- Replace `Age int` with `DateOfBirth time.Time` in `User` struct
- Add `AboutMe string`
- Keep `Age` for backwards compat during migration, mark deprecated

### 5.2 Backend handler
- Modify register handler to accept: `date_of_birth`, `about_me`, `avatar` (file upload)
- Convert DOB to age on server side for existing age field, or store DOB directly
- Add avatar upload processing during registration (not just post-creation)

### 5.3 Frontend
- Modify registration form:
  - Date of Birth field (date type input) instead of Age number
  - About Me textarea
  - Avatar file upload with preview
  - Make Nickname optional (remove required validation, keep field)
- Update `register.js`

### 5.4 Profile display
- Update activity/profile page to show:
  - Date of Birth (formatted, not raw age)
  - About Me/bio
  - Avatar

---

## Phase 6: Groups

### 6.1 Domain
- Implement `internal/domain/group/` with all structs + repository interface

### 6.2 Infrastructure
- `internal/infra/storage/sqlite/groups/groupRepo.go`
- Wire into bootstrap

### 6.3 HTTP handlers
- `POST /api/v1/groups/create` — create group
- `GET /api/v1/groups/{id}` — get group details
- `GET /api/v1/groups` — list all groups (browse page)
- `POST /api/v1/groups/{id}/invite` — invite user to group
- `POST /api/v1/groups/invitations/{id}/accept` — accept invitation
- `POST /api/v1/groups/invitations/{id}/reject` — reject invitation
- `POST /api/v1/groups/{id}/request-join` — request to join
- `POST /api/v1/groups/requests/{id}/accept` — accept join request (creator only)
- `POST /api/v1/groups/requests/{id}/reject` — reject join request (creator only)
- `POST /api/v1/groups/leave` — leave group
- `GET /api/v1/groups/{id}/members` — list members
- `POST /api/v1/groups/{id}/posts` — create group post (topic with group_id)
- `GET /api/v1/groups/{id}/posts` — list group posts

### 6.4 Topic extension
- Add `group_id` column to topics table (nullable, FK to groups)
- Group posts are only visible to group members

### 6.5 Frontend
- Create `groups.js` page module — browse all groups
- Create `group.js` page module — single group view (members, posts, events)
- Create `createGroup.js` — group creation form
- Group navigation in sidebar/navbar
- Invitation handling UI (accept/reject in notifications)
- Join request UI
- Group posts appear in group feed only (not main feed unless member)

---

## Phase 7: Events

### 7.1 Domain
- Implement `internal/domain/event/` with structs + repository interface

### 7.2 Infrastructure
- `internal/infra/storage/sqlite/events/eventRepo.go`
- Wire into bootstrap

### 7.3 HTTP handlers
- `POST /api/v1/groups/{id}/events/create` — create event (group member only)
- `GET /api/v1/groups/{id}/events` — list group events
- `POST /api/v1/events/{id}/rsvp` — RSVP (going / not going)
- `GET /api/v1/events/{id}/rsvps` — get RSVP list

### 7.4 Frontend
- Event creation form in group page (title, description, day/time, going/not going options)
- Event listing in group page
- RSVP button with current status
- Event display with attendee counts

---

## Phase 8: Group Chat

### 8.1 Schema
- `group_chats` table: chat_id, group_id
- `group_chat_messages` table: id, chat_id, sender_id, content, created_at

### 8.2 Domain
- Extend chat domain to support group chats (or create separate group_chat domain)

### 8.3 WebSocket
- Extend WS router with `group_chat.send`, `group_chat.history`, `group_chat.typing` handlers
- Group chat messages broadcast to all online group members

### 8.4 Frontend
- Group chat tab/button in group page
- Real-time group messaging
- Emoji support (same as direct chat)

---

## Phase 9: Notification Types Enhancement

### 9.1 Backend
- Add new notification type constants in domain
- Extend notification creation to cover:
  - Follow request received (for private profile follows)
  - Follow request accepted
  - Group invitation received
  - Group join request received (for creator)
  - Group join request accepted
  - Event created in group

### 9.2 Frontend
- Update notification display in navbar dropdown to handle new types
- Action buttons in notification items (accept/reject follow requests, accept/reject group invitations, accept/reject join requests)

---

## Phase 10: Infrastructure

### 10.1 docker-compose — 2 services
- Split `docker-compose.yml` into two services:
  - `backend`: builds from `Dockerfile.backend`, exposes port 8080
  - `frontend`: builds from `Dockerfile.frontend` (nginx serving static files OR the BFF Go server), exposes port 3001
- Update `Dockerfile` into two separate Dockerfiles:
  - `Dockerfile.backend` — Go API server binary
  - `Dockerfile.frontend` — static file server (nginx) or BFF Go server
- Frontend container connects to backend via `BACKEND_URL` env var
- Both containers share `forum-network`

### 10.2 JS Framework evaluation
- Current: Vanilla JS SPA (hand-rolled router)
- If strict spec compliance is required: Migrate to a listed framework
- Recommended: Vue.js (lightest migration path for template-based rendering)
  - Or: Keep current SPA but wrap in Mithril (lightweight, same rendering style)
- Decision needed: is framework strictly required by auditors?

---

## Phase 11: Existing Bug Fixes from Codebase

### 11.1 OAuth handler uses `UserLoginGithub` for Google too
**File:** `internal/infra/http/server.go:197`
**Change:** Pass `UserLoginGoogle` query service for Google OAuth handler (if it exists), or verify the generic `UserLoginGithub` service actually works generically via `Provider` interface.

### 11.2 Notification `actor_id` not in schema
**File:** `db/migrations/schema.sql` + `notificationsRepo.go`
**Check:** The `Notification` struct has `ActorID` but the schema doesn't have this column. May cause silent data loss.

---

## Dependency Graph

```
Phase 1 (Bug Fixes) ─────────────────── no dependencies
       │
Phase 2 (Foundation) ────────────────── depends on Phase 1
       │
       ├── Phase 3 (Followers/Privacy) ─ depends on Phase 2
       │       │
       │       ├── Phase 4 (Post Privacy) ─ depends on Phase 3 (followers exist)
       │       │
       │       └── affects Chat init check (Phase 3.5)
       │
       ├── Phase 5 (Registration) ───── depends on Phase 2 (schema changes)
       │
       ├── Phase 6 (Groups) ─────────── depends on Phase 2
       │       │
       │       ├── Phase 7 (Events) ─── depends on Phase 6
       │       │
       │       ├── Phase 8 (Group Chat) ─ depends on Phase 6
       │       │
       │       └── affects Phase 9 (Notifications)
       │
       └── Phase 9 (Notifications) ──── depends on Phase 3, Phase 6, Phase 7
                          │
               Phase 10 (Infrastructure) ─ no dependencies, can run parallel
```

## Execution Order (Recommended)

1. **Phase 1** — Fix critical bugs (safe, independent)
2. **Phase 2.1** — Migration system upgrade (foundation for all schema changes)
3. **Phase 2.2—2.3** — Domain models + schema changes (add columns, new tables)
4. **Phase 3** — Followers & Profile Privacy (core feature, needed by other features)
5. **Phase 5** — Registration Enhancement (DOB, About Me, Avatar)
6. **Phase 4** — Post Privacy (depends on followers)
7. **Phase 6** — Groups (independent of followers)
8. **Phase 7** — Events (depends on groups)
9. **Phase 8** — Group Chat (depends on groups)
10. **Phase 9** — Notification Types (depends on all features above)
11. **Phase 10** — Infrastructure (docker-compose, JS framework)
12. **Phase 11** — Existing codebase bug fixes

## Effort Estimate

| Phase | Description | Files Changed | Est. Effort |
|-------|-------------|--------------|-------------|
| 1 | Critical Bug Fixes | 6 | 0.5 day |
| 2 | Foundation (domain + schema) | 25+ | 1.5 days |
| 3 | Followers & Privacy | 30+ | 2 days |
| 4 | Post Privacy | 15+ | 1 day |
| 5 | Registration Enhancement | 10+ | 0.5 day |
| 6 | Groups | 35+ | 2 days |
| 7 | Events | 15+ | 1 day |
| 8 | Group Chat | 10+ | 1 day |
| 9 | Notifications | 8+ | 0.5 day |
| 10 | Infrastructure | 5+ | 0.5 day |
| 11 | Existing bug fixes | 5+ | 0.5 day |
| **Total** | | | **~10 days** |

## Verification Strategy

After each phase:
1. `go vet ./...` — type correctness
2. `go build ./...` — compilation
3. `go test -race -coverprofile=coverage.out ./...` — unit tests
4. Manual testing of new endpoints via curl/httpie
5. Frontend: visual verification + console error checking

End-to-end:
- Full registration → follow → private profile → post with privacy → group creation → event → chat flow
- Docker build + compose up verification
