# Optimized Architecture Plan — Vertical Slices with Infrastructure Services

Standalone execution plan for Proposal 3 from `docs/plan/arch-proposals.md`.

---

## Target Architecture Overview

```
internal/
  ┌── Feature Slices ──────────────────────────────────────────┐
  │  user/          topic/         comment/       vote/         │
  │  follow/        group/         event/         chat/         │
  │  notification/  oauth/                                      │
  │  Each: entity.go + commands.go + queries.go + transport/ + store/ │
  └─────────────────────────────────────────────────────────────┘
  ┌── Cross-cutting ───────────────────────────────────────────┐
  │  session/       realtime/      middleware/     server/       │
  └─────────────────────────────────────────────────────────────┘
  ┌── Platform ────────────────────────────────────────────────┐
  │  platform/database/   platform/redis/   platform/eventbus/  │
  └─────────────────────────────────────────────────────────────┘
  ┌── Shared Utilities ────────────────────────────────────────┐
  │  pkg/bcrypt  pkg/uuid  pkg/validator  pkg/helpers           │
  │  pkg/oauth/  pkg/imgutil/                                   │
  └─────────────────────────────────────────────────────────────┘
  config/   bootstrap/
```

---

## Phase 0: Platform Infrastructure

### 0.1 Database Factory (`internal/platform/database/`)

Create the database abstraction layer supporting both SQLite and PostgreSQL.

#### Files

- `internal/platform/database/database.go` — `DB` interface (wrapping `*sql.DB` methods), `NewDB(cfg)` factory
- `internal/platform/database/sqlite.go` — SQLite init with `_journal_mode=WAL`, `_busy_timeout=5000`
- `internal/platform/database/postgres.go` — PostgreSQL connection pool init
- `internal/platform/database/migrations.go` — Migration runner reading from `db/migrations/`, tracking in `schema_migrations` table

#### Key Rules

- SQLite DSN **must** include `_journal_mode=WAL` and `_busy_timeout=5000`
- PostgreSQL DSN uses `pgx` driver with connection pool config
- Migration files: `000001_initial_schema.up.sql` / `000001_initial_schema.down.sql`
- Migration delimiter: split on `";"` (never `":"`)
- Config selects driver via `DATABASE_DRIVER=sqlite|postgres`

### 0.2 Redis Client (`internal/platform/redis/`)

#### Files

- `internal/platform/redis/redis.go` — Client init, connection pool, health check, config
- `internal/platform/redis/pubsub.go` — Pub/Sub wrapper for realtime event fan-out

#### Use Cases

| Use Case | Implementation |
|----------|----------------|
| Session cache | `session/store/redis.go` uses `GET`/`SETEX` with TTL |
| Rate limiting | `middleware/ratelimiter.go` uses `INCR` + `EXPIRE` |
| Realtime pub/sub | SSE/WS hub subscribes to Redis channels for cross-instance fan-out |
| Online presence | Sorted sets tracking user online status |

### 0.3 Event Bus & RabbitMQ Client (`internal/platform/eventbus/` & `internal/platform/rabbitmq/`)

To allow swapping RabbitMQ for Kafka, NATS, or other message brokers in the future, all event publishing must be abstracted. Use-case layers MUST NOT import RabbitMQ directly.

#### Files

- `internal/platform/eventbus/eventbus.go` — `EventBus` interface definition (`Publish(ctx, event, payload) error`)
- `internal/platform/rabbitmq/rabbitmq.go` — Connection, channel management, auto-reconnect
- `internal/platform/rabbitmq/publisher.go` — Concrete implementation of `EventBus` using RabbitMQ
- `internal/platform/rabbitmq/consumer.go` — Consume events, dispatch to services
- `internal/platform/rabbitmq/exchanges.go` — Exchange/queue/binding declarations

#### Exchanges and Queues

| Exchange | Routing Key | Queue | Consumer |
|----------|-------------|-------|----------|
| `social.events` | `follow.requested` | `notifications.follow` | notification service |
| `social.events` | `follow.accepted` | `notifications.follow` | notification service |
| `social.events` | `group.invited` | `notifications.group` | notification service |
| `social.events` | `group.join_requested` | `notifications.group` | notification service |
| `social.events` | `event.created` | `notifications.event` | notification service |

Dead-letter exchange for failed messages with configurable retry policy.

---

## Phase 1: Cross-cutting Core

### 1.1 Session (`internal/session/`)

- `session/session.go` — Session entity, Manager interface
- `session/store/sqlite.go` — SQLite session store
- `session/store/postgres.go` — PostgreSQL session store
- `session/store/redis.go` — Redis session cache (read-through to DB)

### 1.2 Realtime (`internal/realtime/`)

Move from `internal/infra/ws/`:

- `realtime/hub.go` — WebSocket connection hub
- `realtime/client.go` — Client lifecycle (ReadPump/WritePump with `defer recover()`)
- `realtime/router.go` — WS message routing by type

### 1.3 Middleware (`internal/middleware/`)

Move from `internal/infra/middleware/`:

- `middleware/auth.go` — Session auth middleware
- `middleware/cors.go` — CORS middleware (validate origin properly)
- `middleware/ratelimiter.go` — Rate limiter using Redis (flatten from sub-package)
- `middleware/logging.go` — Request logging

### 1.4 Server (`internal/server/`)

Move from `internal/infra/http/server.go`:

- `server/server.go` — HTTP server struct, `NewServer()`, `ListenAndServe()` with graceful shutdown (handles `SIGTERM`/`SIGINT`, drains RabbitMQ, and closes DB pools).
- `server/routes.go` — Route registration table including `/healthz` (liveness) and `/readyz` (readiness check for DB, Redis, RabbitMQ).

---

## Phase 2: Greenfield Feature Slices

Build new features as vertical slices. No migration needed — these don't exist yet.

### 2.1 Follow (`internal/follow/`)

| File | Contents |
|------|----------|
| `follow/follow.go` | `Follow`, `FollowRequest` entities, `Repository` interface |
| `follow/commands.go` | `Follow()`, `Unfollow()`, `SendRequest()`, `RespondToRequest()` |
| `follow/queries.go` | `GetFollowers()`, `GetFollowing()`, `GetPendingRequests()` |
| `follow/transport/http.go` | HTTP handlers for all follow API endpoints |
| `follow/store/sqlite.go` | SQLite impl for `follows` + `follow_requests` tables |
| `follow/store/postgres.go` | PostgreSQL impl |

**Tables**: `follows`, `follow_requests`

**Notification dispatch**: On follow-request creation, publish `follow.requested` event to RabbitMQ. On accept, publish `follow.accepted`.

### 2.2 Group (`internal/group/`)

| File | Contents |
|------|----------|
| `group/group.go` | `Group`, `GroupMember`, `GroupInvitation`, `GroupJoinRequest`, `GroupChatMessage` entities, `Repository` interface |
| `group/commands.go` | C/U/D, `InviteToGroup()`, `RespondToInvitation()`, `RequestJoin()`, `RespondToJoinRequest()`, `LeaveGroup()`, send chat |
| `group/queries.go` | Read, `GetMembers()`, `GetPosts()`, get chat history |
| `group/transport/http.go` | REST endpoints for group management + content |
| `group/transport/ws.go` | Group chat WebSocket message handling |
| `group/store/sqlite.go` | SQLite impl for `groups`, `group_members`, `group_invitations`, `group_join_requests`, `group_chat_messages` |
| `group/store/postgres.go` | PostgreSQL impl |

**Tables**: `groups`, `group_members`, `group_invitations`, `group_join_requests`, `group_chat_messages`

**Notification dispatch**: Publish `group.invited` and `group.join_requested` events to RabbitMQ.

### 2.3 Event (`internal/event/`)

| File | Contents |
|------|----------|
| `event/event.go` | `Event`, `EventRSVP` entities, `Repository` interface |
| `event/commands.go` | `CreateEvent()`, `RSVP()` |
| `event/queries.go` | `GetGroupEvents()`, `GetRSVPs()` |
| `event/transport/http.go` | HTTP handlers |
| `event/store/sqlite.go` | SQLite impl for `events` + `event_rsvps` |
| `event/store/postgres.go` | PostgreSQL impl |

**Tables**: `events`, `event_rsvps`

**Notification dispatch**: Publish `event.created` to RabbitMQ (fans out to all group members).

---

## Phase 3: Migrate Existing Features

Move existing code from the current `domain/ → app/ → infra/` structure into vertical slices.

### Per-Feature Migration Steps

For each feature (user, topic, comment, vote, chat, notification, oauth):

1. Create `internal/<feature>/` directory
2. **Entity**: Copy from `domain/<feature>/*.go` → `internal/<feature>/<feature>.go`
3. **CQRS**: Copy from `app/<feature>/commands/*.go` → `internal/<feature>/commands.go` and `app/<feature>/queries/*.go` → `internal/<feature>/queries.go`
4. **Handler**: Merge from `infra/http/<feature>/<action>/*.go` → `internal/<feature>/transport/http.go`
5. **Store (SQLite)**: Copy from `infra/storage/sqlite/<feature>/*.go` → `internal/<feature>/store/sqlite.go`
6. **Store (PostgreSQL)**: Create `internal/<feature>/store/postgres.go` implementing same interface
7. Update all imports project-wide
8. Run `go vet ./...` + `go test -race ./...`
9. Delete old directory

### Specific Merge Notes

#### `user/` — absorbs `activity/`

- `domain/user/user.go` + `domain/activity/activity.go` → `user/user.go`
- `app/user/commands/*.go` → `user/commands.go`
- `app/user/queries/*.go` + `app/activities/queries/*.go` → `user/queries.go`
- All user handlers + activity handler → `user/transport/http.go`
- `infra/storage/sqlite/users/*.go` + `infra/storage/sqlite/activity/*.go` → `user/store/sqlite.go`

#### `topic/` — absorbs `category/`

- `domain/topic/topic.go` + `domain/category/category.go` → `topic/topic.go`
- `app/topics/commands/**` + `app/categories/commands/**` → `topic/commands.go`
- `app/topics/queries/**` + `app/categories/queries/**` → `topic/queries.go`
- All topic + category handlers → `topic/transport/http.go`
- `infra/storage/sqlite/topics/*.go` + `infra/storage/sqlite/categories/*.go` → `topic/store/sqlite.go`
- Add `topic_allowed_users` queries to `topic/store/sqlite.go`

#### `chat/` — gets `transport/ws.go`

- Move WS message handlers from `infra/ws/handlers/` → `chat/transport/ws.go`
- The shared hub stays in `realtime/`

---

## Phase 4: Cleanup

### 4.1 Delete Empty Directories

After all features are migrated:
- Delete `internal/domain/`
- Delete `internal/app/`
- Delete `internal/infra/`

### 4.2 Rename `pkg/` Contents

| Current | New | Why |
|---------|-----|-----|
| `pkg/oAuth/` | `pkg/oauth/` | Go naming: no camelCase in package names |
| `pkg/oAuth/githubclient/` | `pkg/oauth/github/` | Remove stutter |
| `pkg/oAuth/googleclient/` | `pkg/oauth/google/` | Remove stutter |
| `pkg/oAuth/httpclient/` | `pkg/oauth/client.go` | Single file, no sub-package needed |

### 4.3 Add `pkg/imgutil/`

Create `pkg/imgutil/detect.go` — wrapper around `http.DetectContentType` for image MIME validation (JPEG, PNG, GIF magic bytes).

---

## Phase 5: Notification Consumer Wiring

Wire RabbitMQ consumers to the notification service:

1. `platform/rabbitmq/consumer.go` subscribes to `social.events` exchange
2. Routes events by routing key to notification service methods
3. `notification/commands.go` creates the appropriate notification and pushes through the SSE/WS notifier
4. Failed dispatches go to dead-letter queue for retry

---

## Boundary Rules (Enforced)

```
Feature root (user.go, commands.go, queries.go)
  ├── MUST NOT import transport/ or store/ from same feature
  ├── MAY import other features' root packages (entity + interface)
  ├── MAY import platform/ packages (eventbus, redis)  # EventBus interface only (no rabbitmq/kafka)
  └── MUST NOT import other features' transport/ or store/

transport/http.go
  ├── Imports own feature root (service, entities)
  ├── Imports session/ for auth context
  └── MUST NOT import store/

store/sqlite.go, store/postgres.go
  ├── Imports own feature root (entities + repository interface)
  ├── Imports database/sql
  └── MUST NOT import transport/

bootstrap/bootstrap.go
  └── Imports EVERYTHING, wires concrete implementations (e.g., rabbitmq.Publisher) into platform interfaces (e.g., eventbus.EventBus).
```

---

## Cross-Feature Import Map

These cross-imports are acceptable and documented:

```
topic        → (none)
comment      → topic
vote         → topic, comment, notification
chat         → user, follow
follow       → notification (via RabbitMQ publish)
group        → notification (via RabbitMQ publish)
event        → group, notification (via RabbitMQ publish)
notification → (shared kernel — imported by others)
```

---

## Docker Compose Services

```yaml
services:
  backend:    # Go API server (port 8080)
  frontend:   # Next.js (port 3000)
  redis:      # Redis (port 6379)
  rabbitmq:   # RabbitMQ (port 5672 + management 15672)
  # SQLite: file-based, no service needed
  # PostgreSQL: optional, add postgres service when using PG driver
```

---

## Verification Checklist

### Automated
- `go vet ./...`
- `go build ./...`
- `go test -race -coverprofile=coverage.out ./...`
- `golangci-lint run`
- `govulncheck ./...`
- Grep: no `transport/` or `store/` package imports a sibling feature's `transport/` or `store/`

### Manual
- `server/routes.go` import list ≤ 10 clean imports (no aliases needed)
- Each feature's `_test.go` files compile and pass independently
- Redis connection: session cache hit rate, rate limiter counters
- RabbitMQ: messages published and consumed for each notification type
- Docker compose: all 4 services start and communicate
