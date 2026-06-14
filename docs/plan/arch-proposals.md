# Go Package Architecture Proposals

Two alternatives to the current `domain/ → app/ → infra/` horizontal layering, both respecting the AGENTS.md constraint: **domain never imports infrastructure**.

---

## Proposal 1: Refined Horizontal Layers

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

---

## Comparison Table

| Criterion | Current (Baseline) | Proposal 1: Horizontal | Proposal 2: Vertical |
|-----------|-------------------|----------------------|---------------------|
| Package count | ~55 | ~40 | ~35 |
| Handler imports in server.go | 32 aliased | ~12 clean | ~12 clean |
| Dirs to touch per feature | 4 (domain, app, infra/http, infra/sqlite) | 4 (domain, service, handler, repo) | **1** (feature dir) |
| Import stuttering | `usercommands.CreateUserHandler` | `usersvc.CreateUser` | `user.CreateUser` |
| Go-idiomatic naming | Low (infra, app, per-action handlers) | Medium (service, handler, repo) | **High** (package-by-feature) |
| Migration effort | — | Low (rename + flatten only) | High (restructure all files) |
| CQRS clarity | High (explicit commands/queries) | Medium (merged into service) | Medium (merged into service) |
| Feature isolation | Low (scattered) | Low (scattered) | **High** (co-located) |

---

## Recommendation

**Use Proposal 2 for new features** (follow, group, event) immediately — they're greenfield, no migration cost.

**Migrate existing features incrementally** when there's budget. The current structure works; the worst pain points are the 32 handler packages (fixable in Proposal 1) and the `infra/` naming. Even a partial migration to vertical slices for high-churn features (user, topic, chat) provides immediate locality benefits.
