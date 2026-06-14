# Go Package Architecture Proposals

Three alternatives to the current `domain/ → app/ → infra/` horizontal layering, all respecting the AGENTS.md constraint: **domain never imports infrastructure**.

This document includes a deep review of each proposal verified against the live codebase and `docs/plan/sn-merged-plan.md`.

---

## Current Codebase: Quantified Pain Points

| Metric | Actual |
|--------|--------|
| Handler packages under `infra/http/` | **32 directories** (one per action) |
| Aliased imports in `internal/infra/http/server.go` | **38 aliased imports** (lines 15–53) |
| CQRS command/query packages under `app/` | **27 directories** |
| Total Go files in `internal/` | **182** |
| Domain packages | **10** (each as its own subdirectory) |
| Storage repo packages | **9** subdirectories under `infra/storage/sqlite/` |

> **Core problem**: 32 handler packages and 27 CQRS directories for ~10 entities. The ratio of packaging overhead to business logic is roughly 3:1.

---

## Proposal 1: Refined Horizontal Layers (still Clean CQRS)

Conservative evolution. Keeps layer separation but renames to Go-idiomatic terms and flattens the 32+ per-action handler packages.

### Directory Tree

```
internal/
  domain/                         # Business entities + repository interfaces
    user.go                         user.User, user.Repository
    topic.go                        topic.Topic, topic.Repository
    comment.go                      comment.Comment, comment.Repository
    category.go                     category.Category, category.Repository
    chat.go                         chat.Chat, chat.Repository
    notification.go                 notification.Notification, notification.Repository
    vote.go                         vote.Vote, vote.Repository
    activity.go                     activity.Activity, activity.Repository
    oauth.go                        oauth.OAuth, oauth.Repository
    session.go                      session.Session, session.Manager
    follow.go                       follow.Follow, follow.FollowRequest, follow.Repository
    group.go                        group.Group, group.GroupMember, group.Repository
    event.go                        event.Event, event.EventRSVP, event.Repository

  service/                        # Use cases (commands + queries per entity)
    user.go                         usersvc.Service (Register, Login, GetMe, Update)
    topic.go                        topicsvc.Service (Create, Get, List, Update, Delete)
    comment.go                      commentsvc.Service (Create, Get, ListByTopic, Update, Delete)
    category.go                     categorysvc.Service (Create, Get, List, Update, Delete)
    chat.go                         chatsvc.Service (Init, Send, GetHistory, GetUsers)
    notification.go                 notificationsvc.Service (Create, List, Stream, MarkRead)
    vote.go                         votesvc.Service (Cast, Delete, GetCounts)
    activity.go                     activitysvc.Service (GetUserActivity)
    oauth.go                        oauthsvc.Service (LoginGithub, LoginGoogle)
    follow.go                       followSvc.Service (Follow, Unfollow, Request, Respond, List)
    group.go                        groupsvc.Service (CRUD, Invite, Join, Members, Posts)
    event.go                        eventsvc.Service (Create, RSVP, List)
    services.go                     Service registry (DI wiring)

  handler/                        # HTTP handlers — ONE file per entity, NOT per action
    user.go                         userhandler — Register, Login, GetMe, Logout, Update
    topic.go                        topichandler — CreateTopic, GetTopic, GetAllTopics, etc.
    comment.go                      commenthandler — Create, Get, ListByTopic, Update, Delete
    category.go                     categoryhandler — Create, Get, List, Update, Delete
    chat.go                         chathandler — InitChat, GetUsers
    notification.go                 notificationhandler — List, UnreadCount, MarkRead, Stream
    vote.go                         votehandler — Cast, Delete, GetCounts
    activity.go                     activityhandler — GetUserActivity
    oauth.go                        oauthhandler — GithubLogin, GoogleLogin, Callbacks
    follow.go                       followhandler — Follow, Unfollow, Requests, Respond
    group.go                        grouphandler — CRUD, Invite, Join, Members, Posts
    event.go                        eventhandler — Events, RSVP
    health.go                       healthhandler — HealthCheck
    server.go                       Router setup, middleware wiring

  realtime/                       # WebSocket + SSE (cross-cutting, not entity-specific)
    hub.go                          Realtime hub (connection registry)
    client.go                       WebSocket client lifecycle
    router.go                       WS message routing
    notifications/
      notifier.go                   SSE notification dispatcher

  repository/                     # SQLite implementations
    user.go                         userrepo — UserRepository impl
    topic.go                        topicrepo — TopicRepository impl
    comment.go                      commentrepo — CommentRepository impl
    category.go                     categoryrepo — CategoryRepository impl
    chat.go                         chatrepo — ChatRepository impl
    notification.go                 notificationrepo — NotificationRepository impl
    vote.go                         voterepo — VoteRepository impl
    activity.go                     activityrepo — ActivityRepository impl
    oauth.go                        oauthrepo — OAuthRepository impl
    sessionstore.go                 sessionrepo — SessionStore impl
    follow.go                       followrepo — FollowRepository impl
    group.go                        grouprepo — GroupRepository impl
    event.go                        eventrepo — EventRepository impl
    sqlite.go                       DB init, connection config, migration runner

  middleware/                     # HTTP middleware (moved from infra/)
    auth.go                         Session auth middleware
    cors.go                         CORS middleware
    ratelimiter/                    Rate limiter
      rateLimiter.go
    logging.go                      Request logging

  config/
    config.go

  bootstrap/
    bootstrap.go                    App assembly

  pkg/
    bcrypt/
    uuid/
    validator/
    helpers/
    oAuth/
      githubclient/
      googleclient/
      httpclient/
    path/
    testing/
```

### What Changes vs Current

| Current | Proposal 1 | Why |
|---------|-----------|-----|
| `internal/app/` | `internal/service/` | "app" is vague; "service" is the standard Go name for business logic orchestration |
| `internal/infra/` | Split into `handler/`, `repository/`, `realtime/`, `middleware/` | Name packages by *what they are*, not their architectural role |
| `infra/http/<entity>/<action>/` (32 packages) | `handler/<entity>.go` (~12 files) | One file per entity is the Go standard; 32 packages with aliased imports is noise |
| `infra/storage/sqlite/<entity>/` | `repository/<entity>.go` | Flatter, fewer directories, same structure |
| `infra/ws/` | `realtime/` | "WS" is too specific; "realtime" covers WS + SSE |
| `infra/middleware/` | `middleware/` | Middleware is not "infrastructure", it's a first-class HTTP concern |
| `infra/logger/` | (keep separate or merge) | Logger is fine but could live in `pkg/` |
| `infra/services.go` | (n/a — handlers wire directly to service) | No more infra DI file — handlers are injected directly |

### Pros

- **Minimal diff** — same conceptual layers, just renames and flattens. Easy to migrate incrementally.
- **Reduces 32 handler packages to ~12 files** — eliminates 100+ LOC of import aliases in server.go.
- **Eliminates "infra"** — the least Go-idiomatic name is gone.
- **Clear separation persists** — teams used to Clean Architecture still see familiar boundaries.
- **`service/` is standard Go** — used by stdlib-adjacent patterns (net/http middleware, etc.).

### Cons

- **Still horizontal** — adding a `follow` feature still requires touching 4 directories (`domain/`, `service/`, `handler/`, `repository/`).
- **Package name collisions remain** — `domain/user` vs `handler/user` need different package names. We use the `domain/` bare name convention and suffix others (`userhandler`, `usersvc`, `userrepo`).
- **No reduction in file count** — the same number of `.go` files, just rearranged.
- **Stuttering** — `usersvc.Service`, `userhandler.Handler` still carry redundant context.

### Issues Found (Deep Review)

1. **`service/` creates a flat "god package"** — 12+ entity files in one `package service` means every type is `service.UserService`, `service.TopicService`. A package with 12+ files importing different domain types is effectively a god package with the same tight coupling as the current `app/services.go`, just spread across more files.
2. **`repository/` has the same flat-package problem** — all types share `package repository`. `repository.UserRepo`, `repository.TopicRepo` collide in one namespace.
3. **Domain as flat files is a regression** — `domain/user.go`, `domain/topic.go` in a single `domain` package loses the explicit import path (`user.User` becomes `domain.User`). The current sub-package model (`domain/user/`, `domain/topic/`) is superior.
4. **Missing: where do `errors.go` files live?** — the flat `repository/topic.go` model loses the natural place for per-entity error definitions.
5. **`handler/server.go` still has the import problem** — import count drops from 38 to ~12, but the coupling pattern is identical.
6. **`realtime/notifications/` is oddly nested** — SSE notifier under the WebSocket package is confusing.

---

## Proposal 2: Vertical Slices (Package-by-Feature)

Groups all code for a feature in one place. The feature root holds domain types + service logic; `store/` and `transport/` sub-packages implement adapters. This is the "Screaming Architecture" pattern described in Go best-practice literature.

### Directory Tree

```
internal/
  user/
    user.go                     User entity, UserRepository interface
    service.go                  Register, Login, GetMe, UpdateProfile, GetActivity
    transport/
      http.go                   All user HTTP handlers (register, login, getMe, logout, update)
    store/
      sqlite.go                 SQLite UserRepository implementation

  topic/
    topic.go                    Topic entity, TopicRepository interface
    service.go                  Create, Get, GetAll, Update, Delete, FileStorage logic
    transport/
      http.go                   Topic HTTP handlers (CRUD, list, privacy)
    store/
      sqlite.go                 SQLite TopicRepository + topic_allowed_users queries

  comment/
    comment.go                  Comment entity, CommentRepository interface
    service.go                  Create, Get, ListByTopic, Update, Delete
    transport/
      http.go
    store/
      sqlite.go

  category/
    category.go                 Category entity, CategoryRepository interface
    service.go                  Create, Get, GetAll, Update, Delete
    transport/
      http.go
    store/
      sqlite.go

  chat/
    chat.go                     Chat, Message entities, ChatRepository interface
    service.go                  InitChat, SendMessage, GetHistory, GetUsers, MarkAsRead
    transport/
      http.go                   REST endpoints (init, history, users)
      ws.go                     WebSocket message handling (send, receive, typing)
    store/
      sqlite.go

  notification/
    notification.go             Notification entity, NotificationRepository interface
    service.go                  Create, GetList, GetUnreadCount, MarkRead, MarkAllRead, Stream
    transport/
      http.go                   REST endpoints + SSE stream endpoint
    store/
      sqlite.go

  vote/
    vote.go                     Vote entity, VoteRepository interface
    service.go                  CastVote, DeleteVote, GetVoteCounts
    transport/
      http.go
    store/
      sqlite.go

  follow/                       # NEW feature
    follow.go                   Follow, FollowRequest entities, FollowRepository interface
    service.go                  Follow, Unfollow, SendRequest, RespondToRequest,
                                GetFollowers, GetFollowing, GetPendingRequests
    transport/
      http.go
    store/
      sqlite.go

  group/                        # NEW feature
    group.go                    Group, GroupMember, GroupInvitation, GroupJoinRequest entities
                                GroupRepository interface
    service.go                  CRUD, InviteToGroup, RespondToInvitation, RequestJoin,
                                RespondToJoinRequest, LeaveGroup, GetMembers, GetPosts
    transport/
      http.go
    store/
      sqlite.go

  event/                        # NEW feature
    event.go                    Event, EventRSVP entities, EventRepository interface
    service.go                  CreateEvent, RSVP, GetGroupEvents, GetRSVPs
    transport/
      http.go
    store/
      sqlite.go

  oauth/
    oauth.go                    OAuth entity, OAuthRepository interface
    service.go                  LoginGithub, LoginGoogle (generic Provider interface)
    transport/
      http.go                   OAuth redirect + callback handlers
    store/
      sqlite.go

  activity/                     # Read-only, thin
    activity.go                 Activity entity, ActivityRepository interface
    service.go                  GetUserActivity
    transport/
      http.go
    store/
      sqlite.go

  realtime/                     # Cross-cutting (not per-feature)
    hub.go                      WebSocket connection hub
    client.go                   Client connection lifecycle
    router.go                   WS message routing by type

  middleware/                   # Cross-cutting
    auth.go
    cors.go
    ratelimiter/
      rateLimiter.go
    logging.go

  config/
    config.go

  bootstrap/
    bootstrap.go                Wire all features together

  pkg/
    bcrypt/
    uuid/
    validator/
    helpers/
    oAuth/
      githubclient/
      googleclient/
      httpclient/
    path/
    testing/
```

### Boundary Rule (enforced by AGENTS.md)

- `user/user.go`, `user/service.go` — **must not** import `user/transport/` or `user/store/`
- `user/store/sqlite.go` — imports `user/` (domain entities + repo interface) + `database/sql`
- `user/transport/http.go` — imports `user/` (service) + `domain/session` for auth context
- `bootstrap/bootstrap.go` — imports everything, wires concrete impls into service interfaces

### Pros

- **Full locality** — a developer adding "follow" touches only `internal/follow/`. No context-switching across 4 directories.
- **No stuttering imports** — `import "internal/user"` → `user.User`, `user.Service`, `user.Repository`. Clean.
- **No alias needed** — `user` and `topic` are unique import names; they never collide because each feature is self-contained.
- **Go-idiomatic** — matches the standard library's own structure (`net/http/`, `net/mail/`, `crypto/tls/`, etc.) and major Go projects (Kubernetes, Hugo, Caddy).
- **Incremental adoption** — new features (follow, group, event) can be built vertically from day 1. Old code can migrate per-feature.
- **Preserves boundary** — `user.go` + `service.go` don't import transport/store, exactly like current `domain/` doesn't import `infra/`.
- **Simpler `services.go`** — fewer, more cohesive files. Bootstrap wires `user.Service`, `topic.Service`, etc. instead of 32 individual command/query handlers.
- **Removes CQRS ceremony for simple CRUD** — `service.go` exports clean methods instead of requiring `commands.CreateXHandler` + `queries.GetXHandler` for every operation. (Strict CQRS fans can still split `command.go` / `query.go` inside a feature.)

### Cons

- **Larger refactor** — requires moving every existing file from the current layer-based structure.
- **Shared domain cross-imports** — `topic.Topic` references `comment.Comment`. This cross-feature import is acceptable per AGENTS.md ("domain cross-imports acceptable") but creates awareness that features are not fully isolated.
- **Cross-cutting splitting** — WebSocket hub (`realtime/`) is a shared infra. The chat feature owns WS message handling via `chat/transport/ws.go`, but the hub itself lives outside. This is slightly less pure but follows Go's "shared nothing, import what you need" philosophy.
- **CQRS enthusiasts lose explicit separation** — commands and queries merge into `service.go`. If you want strict CQRS, you'd add `command.go` and `query.go` per feature instead of `service.go`.

### Issues Found (Deep Review)

1. **Session: missing from the feature list** — `session` is a core cross-cutting concern used by middleware, handlers, and bootstrap. The current codebase has `domain/session/session.go`, `domain/session/sessionManager.go`, plus `infra/storage/sessionstore/`. Needs explicit placement.
2. **`activity/` is overkill** — exactly 2 domain files and 1 query (`GetUserActivity`). A full 4-file feature slice for a single read-only query is overhead. Should fold into `user/`.
3. **`vote/` is tightly coupled** — `CastVote` needs `topicRepo`, `commentRepo`, `notificationsRepo`, and `notifier` — three cross-feature imports for one small entity, undermining isolation.
4. **`category/` is similarly thin** — CRUD-only, no business logic, pure passthrough. Should fold into `topic/` since categories exist solely to tag topics.
5. **`transport/` naming** — the package name becomes `transport`. Functions read as `transport.RegisterRoutes()`. Acceptable when the project may later add gRPC or other transports.
6. **Cross-feature domain imports** — `notification` is the most-imported domain (every feature creating notifications needs it). This is acceptable as a shared kernel, but should be documented.
7. **`group_chat_messages` placement gap** — neither shown in `group/store/` nor `chat/store/`. Should live in `group/store/` with WS handling in `group/transport/ws.go`.
8. **`topic_allowed_users` placement gap** — the post-privacy join table queries aren't shown. Should live in `topic/store/sqlite.go`.

---

## Proposal 3: Optimized Vertical Slices (Recommended)

Based on the deep review of Proposals 1 and 2. Key improvements:

1. **Fold thin features** (`activity`, `category`) into parent features
2. **Keep `transport/` naming** — leaves room for future gRPC/NATS transports
3. **Dual database support** — `sqlite/` and `postgres/` sub-packages under each feature's `store/`
4. **Add Redis** — session caching, rate limiting, pub/sub for realtime
5. **Add RabbitMQ** — async event/notification dispatch, decoupled service communication
6. **Add missing slices** (`session/`)
7. **Explicit placement** for all `sn-merged-plan.md` tables

### Directory Tree

```
internal/
  user/
    user.go                     # Entity, Repository interface, Activity types
    service.go                  # Register, Login, GetMe, Update, GetActivity
    transport/
      http.go                   # All user HTTP handlers (register, login, me, update, activity)
    store/
      sqlite.go                 # User + Activity SQLite implementation
      postgres.go               # User + Activity PostgreSQL implementation

  topic/
    topic.go                    # Entity (with Visibility enum), Repository interface,
                                # AllowedUser, Category types, CategoryRepository interface
    service.go                  # CRUD, FileStorage, privacy filtering
    transport/
      http.go                   # Topic + Category + AllowedUsers HTTP handlers
    store/
      sqlite.go                 # topic, category, topic_allowed_users, topic_categories queries
      postgres.go               # PostgreSQL equivalent

  comment/
    comment.go                  # Entity, Repository interface
    service.go                  # CRUD, notification dispatch on create
    transport/
      http.go
    store/
      sqlite.go
      postgres.go

  vote/
    vote.go                     # Entity, Repository interface
    service.go                  # Cast, Delete, GetCounts (imports topic, comment, notification)
    transport/
      http.go
    store/
      sqlite.go
      postgres.go

  follow/                       # NEW
    follow.go                   # Follow, FollowRequest entities, Repository interface
    service.go                  # Follow, Unfollow, SendRequest, Respond, List
    transport/
      http.go
    store/
      sqlite.go                 # follows + follow_requests tables
      postgres.go

  group/                        # NEW
    group.go                    # Group, GroupMember, Invitation, JoinRequest,
                                # GroupChatMessage entities, Repository interface
    service.go                  # CRUD, Invite, Join, Members, Posts, GroupChat
    transport/
      http.go                   # REST endpoints for group management + content
      ws.go                     # Group chat WebSocket message handling
    store/
      sqlite.go                 # groups, group_members, group_invitations,
                                # group_join_requests, group_chat_messages tables
      postgres.go

  event/                        # NEW
    event.go                    # Event, EventRSVP entities, Repository interface
    service.go                  # Create, RSVP, List
    transport/
      http.go
    store/
      sqlite.go                 # events + event_rsvps tables
      postgres.go

  chat/
    chat.go                     # Chat, Message, ChatRead entities, Repository interface
    service.go                  # InitChat, Send, History, Users, MarkRead
    transport/
      http.go                   # REST (init, history, users)
      ws.go                     # WebSocket message handling (send, receive, typing)
    store/
      sqlite.go
      postgres.go

  notification/
    notification.go             # Entity, Repository interface, Notifier interface
    service.go                  # Create, List, Stream, MarkRead, MarkAllRead
    transport/
      http.go                   # REST + SSE stream
    store/
      sqlite.go
      postgres.go

  oauth/
    oauth.go                    # Entity, Repository interface
    service.go                  # LoginGithub, LoginGoogle (generic Provider)
    transport/
      http.go                   # OAuth redirect + callback handlers
    store/
      sqlite.go
      postgres.go

  # ─── Cross-cutting infrastructure ───

  session/
    session.go                  # Session entity, Manager interface
    store/
      sqlite.go                 # SQLite session store
      postgres.go               # PostgreSQL session store
      redis.go                  # Redis session cache (fast lookups, TTL-based expiry)

  realtime/
    hub.go                      # WebSocket connection hub
    client.go                   # Client lifecycle (ReadPump, WritePump)
    router.go                   # WS message routing by type

  middleware/
    auth.go                     # Session auth middleware
    cors.go                     # CORS middleware
    ratelimiter.go              # Rate limiter (single file; uses Redis for distributed state)
    logging.go                  # Request logging

  server/
    server.go                   # HTTP server, mux setup, route registration
    routes.go                   # Route table (separated for readability)

  config/
    config.go

  bootstrap/
    bootstrap.go                # DI wiring — imports all features, assembles the app

  # ─── Shared infrastructure services ───

  platform/
    database/
      database.go               # DB interface, Factory (SQLite vs PostgreSQL selector)
      sqlite.go                 # SQLite connection init, WAL config, busy_timeout
      postgres.go               # PostgreSQL connection pool init
      migrations.go             # Migration runner (reads db/migrations/ sequentially)
    redis/
      redis.go                  # Redis client init, connection pool, health check
      pubsub.go                 # Redis Pub/Sub wrapper for realtime event fan-out
    rabbitmq/
      rabbitmq.go               # RabbitMQ connection, channel management, reconnect
      publisher.go              # Publish events (follow-request, group-invite, etc.)
      consumer.go               # Consume events, dispatch to notification service
      exchanges.go              # Exchange/queue/binding declarations

  pkg/                          # Shared utilities (no business logic)
    bcrypt/
    uuid/
    validator/
    helpers/
    oauth/                      # Renamed from oAuth → oauth (Go naming convention)
      github/                   # Renamed from githubclient → github
      google/                   # Renamed from googleclient → google
      client.go                 # Shared HTTP client for OAuth
    imgutil/                    # NEW: http.DetectContentType wrapper for MIME validation
```

### How Redis Fits

| Use Case | How |
|----------|-----|
| **Session cache** | `session/store/redis.go` — fast session lookups. DB is the source of truth, Redis is the hot cache with TTL-based expiry. |
| **Rate limiting** | `middleware/ratelimiter.go` — uses Redis `INCR` + `EXPIRE` for distributed rate limiting across multiple backend instances. |
| **Realtime pub/sub** | `platform/redis/pubsub.go` — when a notification is created, publish to a Redis channel. The SSE/WS hub subscribes and pushes to connected clients. Enables horizontal scaling of the backend. |
| **Online presence** | `realtime/hub.go` can use Redis sorted sets to track user online/offline status across multiple server instances. |

### How RabbitMQ Fits

| Use Case | How |
|----------|-----|
| **Async notification dispatch** | When `follow/service.go` creates a follow-request, it publishes a `follow.requested` event to RabbitMQ instead of calling `notification/service.go` directly. The consumer picks it up and creates the notification asynchronously. This decouples features. |
| **Group event broadcast** | When `event/service.go` creates an event, it publishes `event.created` to RabbitMQ. The consumer fans out notifications to all group members without blocking the HTTP response. |
| **Email/push notifications** | Future consumers can subscribe to the same exchanges for email delivery, push notifications, etc. without changing the publisher. |
| **Retry and dead-letter** | Failed notification dispatches go to a dead-letter queue for retry, instead of silently failing in a goroutine. |

### Database Factory Pattern

```go
// platform/database/database.go
type DB interface {
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
    // ... standard database/sql interface subset
}

// NewDB selects SQLite or PostgreSQL based on config
func NewDB(cfg config.DatabaseConfig) (DB, error) {
    switch cfg.Driver {
    case "sqlite":
        return newSQLite(cfg.DSN)   // includes _journal_mode=WAL, _busy_timeout=5000
    case "postgres":
        return newPostgres(cfg.DSN) // includes connection pool config
    default:
        return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
    }
}
```

Each feature's `store/sqlite.go` and `store/postgres.go` implement the same `Repository` interface. The `bootstrap.go` selects which implementation to wire based on config.

### Boundary Rules (enforced by AGENTS.md)

- `user/user.go`, `user/service.go` — **must not** import `user/transport/` or `user/store/`
- `user/store/sqlite.go` — imports `user/` (domain entities + repo interface) + `database/sql`
- `user/transport/http.go` — imports `user/` (service) + `session/` for auth context
- `bootstrap/bootstrap.go` — imports everything, wires concrete impls into service interfaces
- Feature services may import `platform/rabbitmq/` to publish events (one-way dependency)
- Feature services **never** import another feature's `transport/` or `store/`

### Key Decisions Explained

| Decision | Rationale |
|----------|-----------|
| **`activity` folded into `user`** | Activity is just "user's post + comment history" — a single query. Doesn't warrant its own slice. |
| **`category` folded into `topic`** | Categories exist only to tag topics. No independent business logic. Reduces package count. |
| **`transport/` kept (not `httpapi/`)** | Leaves room for future gRPC, NATS, or other transports. Industry-standard naming in Go service architecture. |
| **`store/` kept (not bare `sqlite/`)** | With dual DB support, `store/sqlite.go` and `store/postgres.go` sit naturally under `store/`. |
| **`session/` as standalone** | Session management is used by middleware and bootstrap. It's cross-cutting but has its own entity + store. |
| **`server/` for HTTP server** | Replaces `infra/http/server.go`. Contains the mux, middleware chain, and route registration. |
| **`ratelimiter.go` flattened** | The current `ratelimiter/` sub-package has exactly one file. A sub-package for one file violates Go conventions. |
| **`pkg/oauth/` renamed** | `oAuth` → `oauth` follows Go naming (no camelCase in package names). `githubclient` → `github` removes stutter. |
| **`imgutil/` added** | Merged plan requires magic-byte MIME validation. Shared utility, not feature-specific. |
| **Group chat in `group/`** | Group chat messages are access-controlled by group membership. Semantically belongs with the group feature. |
| **`platform/` for shared infra** | Database, Redis, and RabbitMQ are infrastructure services used by multiple features. Separating them from `pkg/` (which is pure utilities) makes the dependency direction clear. |
| **RabbitMQ over direct calls** | Decouples notification creation from the feature that triggers it. Enables async processing, retries, and future consumers (email, push). |
| **Redis for sessions + rate limiting** | Enables horizontal scaling. Multiple backend instances can share session state and rate limit counters. |

---

## Comparison Table

| Criterion | Current (Baseline) | Proposal 1: Horizontal | Proposal 2: Vertical | **Proposal 3: Optimized** |
|-----------|-------------------|----------------------|---------------------|--------------------------|
| Package count | ~55 | ~40 | ~35 | **~30** |
| Handler imports in server.go | 32 aliased | ~12 clean | ~12 clean | **~10 clean** |
| Dirs to touch per feature | 4 (domain, app, infra/http, infra/sqlite) | 4 (domain, service, handler, repo) | **1** (feature dir) | **1** (feature dir) |
| Import stuttering | `usercommands.CreateUserHandler` | `usersvc.CreateUser` | `user.CreateUser` | `user.CreateUser` |
| Go-idiomatic naming | Low (infra, app, per-action handlers) | Medium (service, handler, repo) | **High** (package-by-feature) | **High** |
| Migration effort | — | Low (rename + flatten only) | High (restructure all files) | High (same as P2) |
| CQRS clarity | High (explicit commands/queries) | Medium (merged into service) | Medium (merged into service) | Medium |
| Feature isolation | Low (scattered) | Low (scattered) | **High** (co-located) | **High** |
| Flat "god package" risk | Medium (`app/services.go`) | **High** (`service/`, `repository/`) | None | **None** |
| Dual DB support | No | No | No | **Yes** (factory pattern) |
| Horizontal scaling | No | No | No | **Yes** (Redis + RabbitMQ) |
| Async event processing | No | No | No | **Yes** (RabbitMQ) |
| Missing merged-plan items | N/A | 2 | 3 | **0** |

---

## Recommendation

**Use Proposal 3 (Optimized Vertical Slices)** as the target architecture.

### Migration Strategy

**Phase A — Greenfield features first (zero migration)**
Build `follow/`, `group/`, `event/` as vertical slices immediately. They don't exist yet, so there's no old code to move. This validates the pattern before touching working code.

**Phase B — Infrastructure services**
Set up `platform/database/`, `platform/redis/`, `platform/rabbitmq/`. Wire through `bootstrap.go`. Existing features continue using the current structure but can start using the new database factory.

**Phase C — Extract high-churn features**
Move `user/`, `topic/`, `chat/`, `notification/` to vertical slices. Steps per feature:
1. Create the new feature directory (e.g., `internal/user/`)
2. Copy domain types from `domain/user/` → `internal/user/user.go`
3. Copy service logic from `app/user/` → `internal/user/service.go`
4. Copy handler logic from `infra/http/user/` → `internal/user/transport/http.go`
5. Copy repo logic from `infra/storage/sqlite/users/` → `internal/user/store/sqlite.go`
6. Add `store/postgres.go` stub implementing the same interface
7. Update imports across the codebase
8. Delete old directories
9. Run `go vet ./...` + tests

**Phase D — Clean up the rest**
Move `comment/`, `vote/`, `oauth/`, `session/`, `middleware/`, `realtime/`, `server/`. Delete the now-empty `domain/`, `app/`, `infra/` directories.

**Phase E — Rename `pkg/`**
Fix `oAuth` → `oauth`, `githubclient` → `github`, etc.

### Cross-Feature Import Map (Documented)

These cross-imports are acceptable and expected:

```
topic    → (none)
comment  → topic (comment belongs to topic)
vote     → topic, comment, notification (votes target either, trigger notifications)
chat     → user (chat participants), follow (chat access check)
follow   → notification (follow-request triggers notification)
group    → notification (group-invite/join triggers notification)
event    → group, notification (events belong to groups, trigger notifications)
```

`notification` serves as the shared kernel — every feature that dispatches notifications imports it. This is by design.
