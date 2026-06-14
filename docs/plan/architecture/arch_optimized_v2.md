# Optimized Architecture Plan — Vertical Slices with Plug-in Infrastructure Services 

## Guiding Principle

**One pattern, everywhere.** When there are two ways to do something, pick one and use it consistently. Simplicity and readability beat optimization.

---

## System Overview

```
                       +-------------------+
                       |      Browser      |
                       |   (Next.js App)   |
                       |    Port 3000      |
                       +---------+---------+
                                 | (HTTP / WebSocket)
                                 v
                       +-------------------+
                       |    Go Backend     |
                       |    Port 8080      |
                       +---------+---------+
                          /      |      \
                         v       v       v
              +----------+  +---------+  +----------+
              |  SQLite  |  |  Redis  |  | RabbitMQ |
              |  (store) |  | (cache) |  | (events) |
              +----------+  +---------+  +----------+
```

### Components

**Frontend (Next.js)** — Serves the client-side UI on port 3000. Uses the App Router for server-side and client-side rendering. Communicates with the backend via HTTP REST for CRUD operations and WebSocket for real-time chat and notifications. Built with shadcn/ui components + Tailwind CSS styling + Biome for linting/formatting.

**Backend (Go)** — HTTP server on port 8080. Entry point for all API requests. Organized as **vertical feature slices** under `internal/<feature>/`, each encapsulating domain entities, CQRS commands/queries, HTTP transport handlers, and a SQLite store implementation. Cross-cutting concerns (auth, sessions, WebSocket hub, middleware) live in `internal/core/`. Platform abstractions (database factory, event bus, cache) live in `internal/platform/`.

**Infrastructure Services** — Pluggable behind interfaces in `internal/platform/`:
- **SQLite** (required): Primary storage with Write-Ahead Logging (`_journal_mode=WAL`) and busy timeout (`_busy_timeout=5000`).
- **PostgreSQL** (optional): Demonstrates database portability via the factory pattern — swap implementations without touching feature code.
- **Redis** (optional): In-memory cache, session store, rate limiter backend, and real-time pub/sub.
- **RabbitMQ** (optional): Async event bus for cross-feature notifications, replacing the in-process channel implementation.

The backend starts with in-memory infrastructure (channels, maps) and swaps to Redis/RabbitMQ later by reimplementing the same `platform` interfaces — zero changes to feature code.

### Feature Overview

| Feature | Description | New/Migrated |
|---------|-------------|--------------|
| User | Registration, login, profile, privacy toggle, avatar | Migrate from old layers |
| Topic | Posts with visibility (public/almost_private/private), images | Migrate from old layers |
| Comment | Comments on posts with optional images | Migrate from old layers |
| Vote | Upvote/downvote on posts and comments | Migrate from old layers |
| Follow | Public follow, private follow request/accept/decline | Greenfield |
| Group | Create, invite, request join, group chat, group posts | Greenfield |
| Event | Create event with title/description/day-time, RSVP (going/not going) | Greenfield |
| Chat | 1-on-1 direct messaging via WebSocket, follow-gated | Migrate from old layers |
| Notification | Event bus subscriber: follow-request, follow-accepted, group-invite, group-join, event-creation | Migrate from old layers |
| OAuth | GitHub and Google third-party authentication | Migrate from old layers |

---

## Design Decisions (from Architecture Discussions)

These are the settled conclusions. Every phase below follows them.

### D1: Vertical Slices

All features live in `internal/<feature>/` with the same internal structure:

```
internal/<feature>/
  <feature>.go       # Entity structs + Repository interface
  commands.go        # Write operations (functions that mutate state)
  queries.go         # Read operations (functions that return data)
  transport/
    http.go          # HTTP handlers
    ws.go            # WebSocket handlers (only for chat, group chat)
  store/
    sqlite.go        # SQLite implementation of Repository
```

No exceptions. Every feature, new or migrated, follows this layout.

### D2: Interface Strategy — One Rule

- **Within a feature**: the `Repository` interface lives in `<feature>.go`. Commands and queries in the same package accept the full `Repository`. This is simple and readable — no fragmentation.
- **Across features**: the consumer defines a narrow local interface. The producer implements it implicitly (Go duck typing). Wired at boot in `bootstrap.go`.

```go
// WITHIN feature — full interface, same package
// follow/follow.go
type Repository interface {
    CreateFollow(ctx context.Context, f *Follow) error
    DeleteFollow(ctx context.Context, followerID, followeeID string) error
    GetFollowers(ctx context.Context, userID string) ([]Follow, error)
    GetFollowing(ctx context.Context, userID string) ([]Follow, error)
    AreConnected(ctx context.Context, a, b string) (bool, error)
    // ...
}

// follow/commands.go
type Service struct { repo Repository }  // accepts the full interface — fine, same package
```

```go
// ACROSS features — narrow interface, consumer-defined
// chat/commands.go
type FollowChecker interface {
    AreConnected(ctx context.Context, a, b string) (bool, error)
}

type Service struct {
    repo    Repository      // own repo
    follows FollowChecker   // narrow — only what chat needs from follow
}
```

```go
// bootstrap/bootstrap.go — wires concrete into narrow interface
chatSvc := chat.NewService(chatRepo, followSvc)  // followSvc satisfies chat.FollowChecker
```

### D3: Cross-Slice Communication — Three Strategies, Consistently Applied

| When | Strategy | How |
|------|----------|-----|
| **Data references** | ID-only | `Comment` has `AuthorID string`, never `Author user.User` |
| **Sync behavior checks** | Consumer-defined interface | `chat` defines `FollowChecker`, `follow` implements it |
| **Mutation side-effects** | Event bus publish | `follow` publishes `follow.requested`, `notification` subscribes |

The event bus starts as an **in-process Go implementation** (channels). Later, we swap it for RabbitMQ by implementing the same interface. The feature code never changes.

### D4: Database Access — Factory Pattern from Day 1

One `DB` interface. One factory function. Start with SQLite. Add PostgreSQL later without changing any feature code.

```go
// platform/database/database.go
type DB interface {
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func NewDB(cfg Config) (DB, error) {
    switch cfg.Driver {
    case "sqlite":
        return newSQLite(cfg.DSN)  // _journal_mode=WAL, _busy_timeout=5000
    default:
        return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
    }
}
```

Feature stores accept `DB`, not `*sql.DB`. When PostgreSQL support arrives, we add `case "postgres"` — one line in the factory, zero changes in features.

### D5: Boundary Rules

```
feature root (entity.go, commands.go, queries.go)
  ├── MUST NOT import own transport/ or store/
  ├── MAY import platform/eventbus (interface only)
  └── MUST NOT import another feature's transport/ or store/

transport/http.go
  ├── Imports own feature root
  ├── Imports internal/core/session/ for auth context
  └── MUST NOT import store/

store/sqlite.go
  ├── Imports own feature root (entities + repository interface)
  ├── Imports platform/database (DB interface)
  └── MUST NOT import transport/

bootstrap/bootstrap.go
  └── Imports everything, wires concrete implementations
```

### D6: Dependency Graph

```
user           → (nothing)
session        → user
follow         → user, eventbus
topic          → user
comment        → user, topic
vote           → user, topic, comment, eventbus
group          → user, eventbus
event          → group, eventbus
chat           → user, FollowChecker (interface, not follow import)
notification   → user (subscribes to eventbus, no feature imports)
oauth          → user
```

`notification` is never imported by other features. It subscribes to events at boot time. This prevents circular dependencies and keeps notification as a pure side-effect consumer.

---

## Directory Tree — Final Target

```
cmd/
  server/
    main.go         # Application entry point. Imports config, bootstrap, and core/server.

internal/
  # ─── Feature Slices (all follow D1 layout) ───
  user/             # Absorbs activity. Entities: User
  topic/            # Absorbs category. Entities: Topic, Category, AllowedUser
  comment/          # Entities: Comment
  vote/             # Entities: Vote
  follow/           # NEW. Entities: Follow, FollowRequest
  group/            # NEW. Entities: Group, GroupMember, Invitation, JoinRequest, ChatMessage
  event/            # NEW. Entities: Event, EventRSVP
  chat/             # Entities: Chat, Message
  notification/     # Entities: Notification
  oauth/            # Entities: OAuthState

  # ─── Cross-cutting Core ───
  core/
    session/        # Session entity + manager + store
    realtime/       # WebSocket hub, client, router
    middleware/     # Auth, CORS, rate limiter, logging
    server/         # HTTP server, route registration, graceful shutdown

  # ─── Platform (behind interfaces) ───
  platform/
    database/       # DB interface + factory (SQLite now, PostgreSQL later)
    eventbus/       # EventBus interface + in-process impl (RabbitMQ later)
    cache/          # Cache interface + in-memory map (Redis later)

  # ─── Shared Utilities ───
  pkg/
    bcrypt/         # Password hashing
    uuid/           # ID generation
    validator/      # Request validation
    helpers/        # Generic utilities
    oauth/          # OAuth provider clients (github/, google/, client.go)
    imgutil/        # Magic-byte MIME validation

  config/           # App config loader
  bootstrap/        # Composition root — wires everything
```

---

## Phase 1: Critical Bug Fixes

*Fix bugs that cause crashes or security holes. No structural changes.*

| # | Bug | File (current path) | Fix |
|---|-----|---------------------|-----|
| 1.1 | Migration delimiter | `infra/storage/sqlite/init.go` | `":"` → `";"` |
| 1.2 | SQLite DSN missing WAL/timeout | `init.go`, `.env` | Add `_journal_mode=WAL&_busy_timeout=5000` |
| 1.3 | OAuth `Scan()` with `ctx` arg | `infra/storage/sqlite/oauth/oauthRepo.go` | Remove `ctx` from `Scan()` params |
| 1.4 | WebSocket `CheckOrigin` returns `true` | `infra/ws/handler.go` | Validate against configured origin |
| 1.5 | SQL injection in ORDER BY | `sqlite/topics/topicRepo.go`, `sqlite/categories/categoryRepo.go` | Whitelist `["ASC", "DESC"]` |
| 1.6 | Prepared stmt uses `db.Exec` | `sqlite/users/userRepo.go` | Use `stmt.ExecContext` |
| 1.7 | WS goroutine panic recovery | `infra/ws/client.go` | Add `defer recover()` to ReadPump/WritePump |
| 1.8 | RateLimiter ticker leak | `middleware/ratelimiter/rateLimiter.go` | Add `stop chan struct{}` |

**Verify**: `go vet ./...` + `go test -race ./...`

---

## Phase 2: Platform Foundation + Migration System

*Set up the platform abstractions that all features will use. Still SQLite only, but behind interfaces.*

### 2.1 Database Factory (`internal/platform/database/`)

- `database.go` — `DB` interface wrapping `*sql.DB` methods, `NewDB(cfg)` factory
- `sqlite.go` — SQLite init with WAL + busy timeout
- `migrations.go` — Sequential migration runner, `schema_migrations` tracking table

The factory returns `DB`. All feature stores will accept `DB`, not `*sql.DB`.

### 2.2 In-Process Event Bus (`internal/platform/eventbus/`)

- `eventbus.go` — `EventBus` interface: `Publish(ctx, eventType, payload)`, `Subscribe(eventType, handler)`
- `memory.go` — In-process implementation using Go channels and goroutines

~40 lines of code. No external dependencies. Features publish events; subscribers receive them asynchronously in-process.

### 2.3 Simple In-Memory Cache (`internal/platform/cache/`)

- `cache.go` — `Cache` interface: `Get(ctx, key, val)`, `Set(ctx, key, val, ttl)`, `Delete(ctx, key)`
- `memory.go` — In-process implementation using a thread-safe map (`sync.RWMutex`) and cleanup goroutine.

### 2.4 Migration Scripts (`db/migrations/`)

Create numbered migration scripts:

- `000001_initial_schema.up.sql` — Current tables (users, topics, comments, categories, votes, sessions, chats, notifications, oauth_states)
- `000001_initial_schema.down.sql` — Drop all
- `000002_user_profile_fields.up.sql` — Add `date_of_birth`, `about_me`, `is_private` to users; drop `age`
- `000002_user_profile_fields.down.sql` — Reverse
- `000003_topic_privacy.up.sql` — Add `visibility` to topics; create `topic_allowed_users`
- `000003_topic_privacy.down.sql` — Reverse
- `000004_follow_system.up.sql` — Create `follows`, `follow_requests`
- `000005_groups.up.sql` — Create `groups`, `group_members`, `group_invitations`, `group_join_requests`, `group_chat_messages`
- `000006_events.up.sql` — Create `events`, `event_rsvps`

**Verify**: Run migrations on fresh DB. `go vet ./...`.

---

## Phase 3: Cross-Cutting Core

*Move existing cross-cutting concerns into their target locations.*

### 3.1 Session (`internal/core/session/`)

- `session.go` — Session entity, `Manager` interface
- `store/sqlite.go` — SQLite session store (moved from `infra/storage/sessionstore/`)

### 3.2 Realtime (`internal/core/realtime/`)

- `hub.go` — WebSocket hub (moved from `infra/ws/`)
- `client.go` — Client lifecycle with `defer recover()` (bug 1.7 applied here)
- `router.go` — WS message routing by type

### 3.3 Middleware (`internal/core/middleware/`)

- `auth.go` — Session auth middleware
- `cors.go` — CORS with proper origin validation
- `ratelimiter.go` — Rate limiter (flattened from sub-package, bug 1.8 applied)
- `logging.go` — Request logging

### 3.4 Server (`internal/core/server/`)

- `server.go` — HTTP server, `ListenAndServe()`, graceful shutdown
- `routes.go` — Route registration

### 3.5 Shared Utilities

- Rename `pkg/oAuth/` → `pkg/oauth/`
- Rename `pkg/oAuth/githubclient/` → `pkg/oauth/github/`
- Rename `pkg/oAuth/googleclient/` → `pkg/oauth/google/`
- Flatten `pkg/oAuth/httpclient/` → `pkg/oauth/client.go`
- Create `pkg/imgutil/detect.go` — `http.DetectContentType` wrapper

**Verify**: `go vet ./...` + `go test -race ./...` — everything still compiles and passes.

---

## Phase 4: Greenfield Feature Slices

*Build new features that don't exist yet. Vertical slices from scratch. No migration needed.*

### 4.1 Follow (`internal/follow/`)

| File | Contents |
|------|----------|
| `follow.go` | `Follow`, `FollowRequest` entities, `Repository` interface |
| `commands.go` | `Follow()`, `Unfollow()`, `SendRequest()`, `AcceptRequest()`, `DeclineRequest()` |
| `queries.go` | `GetFollowers()`, `GetFollowing()`, `GetPendingRequests()`, `AreConnected()` |
| `transport/http.go` | HTTP handlers |
| `store/sqlite.go` | SQLite implementation |

**Key behavior**:
- Public profile → instant follow, publish `follow.accepted` to eventbus
- Private profile → creates follow request, publish `follow.requested` to eventbus
- `notification` subscribes to both events and creates appropriate notifications

### 4.2 Group (`internal/group/`)

| File | Contents |
|------|----------|
| `group.go` | `Group`, `GroupMember`, `Invitation`, `JoinRequest`, `ChatMessage` entities, `Repository` interface |
| `commands.go` | CRUD, `Invite()`, `RespondToInvite()`, `RequestJoin()`, `RespondToJoin()`, `SendChatMessage()` |
| `queries.go` | `GetGroup()`, `ListGroups()`, `GetMembers()`, `GetChatHistory()` |
| `transport/http.go` | REST handlers |
| `transport/ws.go` | Group chat WebSocket handlers |
| `store/sqlite.go` | SQLite implementation |

**Key behavior**:
- Invite/join request → publish `group.invited` / `group.join_requested` to eventbus
- Group chat uses WebSocket hub for real-time message delivery

### 4.3 Event (`internal/event/`)

| File | Contents |
|------|----------|
| `event.go` | `Event`, `EventRSVP` entities, `Repository` interface |
| `commands.go` | `CreateEvent()`, `RSVP()` |
| `queries.go` | `GetGroupEvents()`, `GetRSVPs()` |
| `transport/http.go` | HTTP handlers |
| `store/sqlite.go` | SQLite implementation |

**Key behavior**:
- Event creation → publish `event.created` to eventbus (fans out to all group members)
- Requires `GroupMemberChecker` interface (defined locally, satisfied by `group.Service`)

**Verify**: All new features compile, handlers respond, events publish and trigger notifications. `go test -race ./...`.

---

## Phase 5: Migrate Existing Features to Vertical Slices

*Move existing code from `domain/ → app/ → infra/` into vertical slices. One feature at a time.*

### Per-Feature Migration Steps

For each feature (user, topic, comment, vote, chat, notification, oauth):

1. Create `internal/<feature>/` with D1 layout
2. Copy entity from `domain/<feature>/` → `<feature>/<feature>.go`
3. Merge CQRS from `app/<feature>/commands/` + `queries/` → `commands.go` + `queries.go`
4. Merge handlers from `infra/http/<feature>/` → `transport/http.go`
5. Copy store from `infra/storage/sqlite/<feature>/` → `store/sqlite.go`
6. Update imports
7. `go vet ./...` + `go test -race ./...`
8. Delete old directories

### Special Merge Notes

**`user/` — absorbs `activity/`**
- `domain/user/user.go` + `domain/activity/` → `user/user.go`
- Add `DateOfBirth`, `AboutMe`, `IsPrivate` fields (drop `Age`)
- All user handlers + activity handler → `user/transport/http.go`

**`topic/` — absorbs `category/`**
- `domain/topic/` + `domain/category/` → `topic/topic.go`
- Add `Visibility` enum, `AllowedUser` entity
- Add `http.DetectContentType` for upload MIME validation (uses `pkg/imgutil/`)

**`chat/` — gets `transport/ws.go`**
- Move WS message handlers from `infra/ws/handlers/` → `chat/transport/ws.go`
- Add chat access constraint: require follow relationship (uses `FollowChecker` interface)

**`notification/` — becomes event consumer**
- Wire eventbus subscriptions in `bootstrap.go`
- Subscribes to: `follow.requested`, `follow.accepted`, `group.invited`, `group.join_requested`, `event.created`

### Cleanup

After all features migrated:
- Delete `internal/domain/`
- Delete `internal/app/`
- Delete `internal/infra/`

**Verify**: Full `go vet ./...` + `go test -race ./...` + manual smoke test of all API endpoints.

---

## Phase 6: Next.js Frontend

### 6.1 Scaffold

- Scaffold Next.js app in `frontend/` (App Router)
- Component Library: Integrate **shadcn/ui** for UI components.
- Styling: **Tailwind CSS** with custom HSL values (dark mode, glassmorphism, micro-animations).
- Structure: `src/app/` (routes), `src/components/ui/` (primitives), `src/components/features/` (composite elements), `src/styles/`.
- Code Quality: **Biome** for fast linting, formatting, and import sorting (configured via `biome.json`).
- Typography: Google Fonts (Inter or Outfit).

### 6.2 Core Pages

- Registration form: email, password, first name, last name, date of birth, avatar (optional), nickname (optional), about me (optional)
- Login (email/username + password, GitHub OAuth, Google OAuth)
- Home feed with posts (filtered by visibility)
- Profile page (`/profile/[id]`) with privacy lock screen for non-followers
- Post creation with visibility selector (public / almost_private / private + user picker)

### 6.3 Social Features

- Follow/unfollow buttons with confirmation popups
- Follow request notification with accept/decline
- Groups directory, group page with members, posts, events, chat
- Event RSVP (going / not going)
- Notifications panel with live SSE/WebSocket updates

### 6.4 Real-time

- WebSocket connection for chat (direct + group)
- SSE or WebSocket for notification streaming
- Typing indicators, online presence

---

## Phase 7: Docker Compose (2 Services)

```yaml
services:
  backend:
    build: .
    ports: ["8080:8080"]
    volumes: ["./data:/app/data"]  # SQLite persistence
    environment:
      DATABASE_DRIVER: sqlite
      DATABASE_DSN: /app/data/social.db?_journal_mode=WAL&_busy_timeout=5000

  frontend:
    build: ./frontend
    ports: ["3000:3000"]
    environment:
      NEXT_PUBLIC_API_URL: http://backend:8080
```

**This completes all spec requirements.** Everything below is optional learning.

---

## Phase 8: PostgreSQL Support (Optional — Learning)

*Goal: understand database portability and the factory pattern in practice.*

### 8.1 Add PostgreSQL Driver

- Add `pgx` to `go.mod`
- Create `platform/database/postgres.go` — connection pool init
- Add `case "postgres"` to `NewDB()` factory

### 8.2 Per-Feature `store/postgres.go`

For each feature, create `store/postgres.go` implementing the same `Repository` interface with PostgreSQL-specific SQL syntax.

Since every store accepts the `DB` interface (not `*sql.DB`), and every store implements a `Repository` interface defined in the feature root, this is a parallel implementation — no feature code changes.

### 8.3 Config Switch

```env
DATABASE_DRIVER=postgres
DATABASE_DSN=postgres://user:pass@localhost:5432/social?sslmode=disable
```

### 8.4 Docker Compose Update

Add `postgres` service. Backend switches to PostgreSQL by changing `DATABASE_DRIVER`.

---

## Phase 9: Redis (Optional — Learning)

*Goal: understand caching, distributed rate limiting, and pub/sub.*

### 9.1 Redis Cache (`internal/platform/cache/`)

- `redis.go` — Implements `Cache` interface using Redis connection pool
- `pubsub.go` — Pub/sub wrapper for cross-instance messaging

### 9.2 Session Cache

- Create `internal/core/session/store/redis.go` — Cache implementation utilizing `platform/cache` interface wrapper with TTL. DB remains source of truth.
- Wire in `bootstrap.go` to wrap the SQLite/Postgres session store.

### 9.3 Rate Limiter

- Update `internal/core/middleware/ratelimiter.go` to use the `platform/cache` Redis implementation (e.g. `INCR` + `EXPIRE`) instead of in-memory map.
- Enables consistent rate limiting across multiple backend instances.

### 9.4 Realtime Pub/Sub

- Update `internal/core/realtime/hub.go` to subscribe to Redis channels via the `pubsub` client.
- When a notification is created, publish to Redis. All backend instances receive it and push to their connected WebSocket clients.

### 9.5 Docker Compose Update

Add `redis` service.

---

## Phase 10: RabbitMQ (Optional — Learning)

*Goal: understand message brokers, exchanges, queues, dead-letter handling.*

### 10.1 RabbitMQ Client (`internal/platform/rabbitmq/`)

- `rabbitmq.go` — Connection, channel management, auto-reconnect
- `publisher.go` — Implements `eventbus.EventBus` interface using AMQP
- `consumer.go` — Background consumer dispatching to service methods
- `exchanges.go` — Exchange/queue/binding declarations

### 10.2 Swap EventBus Implementation

In `bootstrap.go`, change one line:

```go
// Before (in-process)
bus := eventbus.NewMemoryBus()

// After (RabbitMQ)
bus := rabbitmq.NewPublisher(amqpConn)
```

Feature code (follow, group, event, notification) doesn't change — they all use the `eventbus.EventBus` interface.

### 10.3 Exchange/Queue Topology

| Exchange | Routing Key | Queue | Consumer |
|----------|-------------|-------|----------|
| `social.events` | `follow.requested` | `notifications.follow` | notification service |
| `social.events` | `follow.accepted` | `notifications.follow` | notification service |
| `social.events` | `group.invited` | `notifications.group` | notification service |
| `social.events` | `group.join_requested` | `notifications.group` | notification service |
| `social.events` | `event.created` | `notifications.event` | notification service |

Dead-letter exchange for failed messages with configurable retry.

### 10.4 Docker Compose Update

Add `rabbitmq` service. Final compose has 4-5 services.

---

## Verification Checklist

### Automated Verification (run after every phase)

#### Backend (Go)
```bash
go vet ./...
go build ./...
go test -race -coverprofile=coverage.out ./...
golangci-lint run
govulncheck ./...
```

#### Frontend (Next.js)
```bash
# Lint and Format checks (Biome)
npx @biomejs/biome lint src/
npx @biomejs/biome format src/

# Type Checking
tsc --noEmit

# Unit & Component Testing
npm run test # runs Vitest

# E2E Testing
npx playwright test
```

### Boundary Verification

```bash
# No feature's transport/ or store/ imports another feature's transport/ or store/
grep -rn 'import' internal/*/transport/ internal/*/store/ | grep 'internal/' | grep -v 'platform/' | grep -v 'pkg/' | grep -v 'infra/'
```

### Manual Test Scenarios

**A: Registration & Login**
1. Register under-13 → rejected
2. Register without nickname/about → succeeds
3. Upload non-image as avatar → rejected (magic bytes)

**B: Follow & Privacy**
1. Set User B private
2. A follows B → follow request + notification to B
3. B declines → no relationship
4. A views B's profile → "Private" lock screen
5. B accepts → A sees full profile, can start chat
6. A unfollows → confirmation popup, relationship severed

**C: Post Privacy**
1. A creates "almost_private" post → visible to followers, hidden from non-followers
2. A creates "private" post selecting only B → visible to B, hidden from others

**D: Group & Event**
1. A creates group
2. A invites B → B gets notification, accepts, joins chat
3. C requests join → A gets notification, accepts
4. A creates event → all members notified
5. B RSVPs "going" → count updates
