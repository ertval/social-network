# Requirements Compliance Verification — Proposal 3

Cross-reference of every requirement from `docs/requirements/audit.md` and `docs/requirements/readme.md` against the optimized proposal in `docs/plan/arch-proposals.md` (Proposal 3) and `docs/plan/arch-optimized-plan.md`.

---

## audit.md — Line-by-Line Verification

### Functional / Project Structure

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L3 | Allowed packages respected? | ⚠️ **GAP** | See Gap #1 below |
| L7 | Backend: well-organized structure, clear separation of packages and migrations folders? | ✅ | Vertical slices under `internal/`. Migrations in `db/migrations/` with numbered up/down format. |
| L9 | Frontend well organized? | ✅ | Out of arch scope — handled by `sn-merged-plan.md` Phase 4 (Next.js in `frontend/`). |

### Backend 3-Tier Separation

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L13 | Clear separation: **Server, App, Database**? | ⚠️ **GAP** | See Gap #2 below |
| L15 | Server receives incoming requests, entry point? | ✅ | `internal/server/server.go` — HTTP server, mux setup, route registration. |
| L17 | App listens for requests, retrieves from DB, sends responses? | ✅ | Each feature's `service.go` contains core logic. `transport/http.go` handles HTTP. `store/sqlite.go` handles DB. |
| L19 | Core logic in App, handling HTTP/other protocols? | ✅ | `service.go` per feature. WebSocket in `transport/ws.go`. |

### Database

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L23 | SQLite being used? | ✅ | `platform/database/sqlite.go`, every feature has `store/sqlite.go`. |
| L25 | Clients can request/submit data without errors? | ✅ | Standard repo pattern. Parameterized queries. |
| L27 | Migration system implemented? | ✅ | `platform/database/migrations.go`. Reads `db/migrations/` sequentially. |
| L29 | Migration file system well organized? | ✅ | `db/migrations/000001_*.up.sql` / `000001_*.down.sql` — matches the subject's example structure. |
| L33 | Migrations applied by migration system? | ✅ | Tracked via `schema_migrations` table. |

### Authentication

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L37 | Sessions for authentication? | ✅ | `session/session.go` (entity + Manager interface), `session/store/sqlite.go`, `middleware/auth.go`. |
| L39 | Registration form: Email, Password, First Name, Last Name, DOB, Avatar (opt), Nickname (opt), About Me (opt)? | ✅ | `user/user.go` entity has all fields. `user/transport/http.go` handles registration. Frontend form in `sn-merged-plan.md` Phase 4.2. |
| L43 | Registration saves user to DB? | ✅ | `user/service.go` → `user/store/sqlite.go` → `users` table. |
| L47 | Login works? | ✅ | `user/service.go` Login methods. |
| L51 | Wrong password/email detected? | ✅ | `user/service.go` validates credentials via bcrypt. |
| L55 | Duplicate email/user detected? | ✅ | `user/store/sqlite.go` enforces UNIQUE on email. |
| L59 | Non-logged browser stays unregistered? | ✅ | `session/` + `middleware/auth.go` — session cookie only set on login. |
| L63 | Both browsers keep right users? | ✅ | Separate session cookies per browser. |

### Followers

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L69 | Send follow request to private user? | ✅ | `follow/service.go` — `SendRequest()`. `follow_requests` table. |
| L73 | Follow public user without request? | ✅ | `follow/service.go` — `Follow()` auto-follows if target is public. |
| L77 | Accept/decline follow request? | ✅ | `follow/service.go` — `RespondToRequest()`. |
| L81 | Unfollow? | ✅ | `follow/service.go` — `Unfollow()`. |

### Profile

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L87 | Profile shows all registration info except password? | ✅ | `user/transport/http.go` — profile endpoint. `user/user.go` entity. |
| L91 | Profile shows all posts by user? | ✅ | `user/service.go` — `GetActivity()` (absorbed from `activity/`). |
| L95 | Profile shows followers and following? | ✅ | `follow/service.go` — `GetFollowers()`, `GetFollowing()`. |
| L99 | Toggle private/public? | ✅ | `user/user.go` — `IsPrivate` field. `user/service.go` — `Update()`. |
| L103 | See followed private profile? | ✅ | `user/transport/http.go` checks follow relationship. |
| L107 | Prevented from seeing non-followed private profile? | ✅ | Same check — returns restricted view if not following. |
| L111 | See non-followed public profile? | ✅ | Public profiles visible to all. |
| L115 | See followed public profile? | ✅ | Public profiles visible to all. |

### Posts

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L119 | Create post and comment? | ✅ | `topic/service.go`, `comment/service.go`. |
| L123 | Include image (JPG/PNG) or GIF in post? | ✅ | `topic/service.go` + `pkg/imgutil/` for MIME validation. |
| L127 | Include image (JPG/PNG) or GIF in comment? | ✅ | `comment/service.go` + image handling. |
| L131 | Post privacy: public, almost private, private? | ✅ | `topic/topic.go` — `Visibility` enum (`public`, `almost_private`, `private`). |
| L133 | Private: specify allowed users? | ✅ | `topic/topic.go` — `AllowedUser` type. `topic/store/sqlite.go` — `topic_allowed_users` table. |

### Groups

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L139 | Invite follower to group? | ✅ | `group/service.go` — `InviteToGroup()`. |
| L143 | Group invitation accept/decline? | ✅ | `group/service.go` — `RespondToInvitation()`. `group_invitations` table. |
| L147 | Group join request accept/decline? | ✅ | `group/service.go` — `RespondToJoinRequest()`. `group_join_requests` table. |
| L149 | Non-creator member can invite? | ✅ | `group/service.go` — `InviteToGroup()` checks membership, not just creator. |
| L151 | User can request to join? | ✅ | `group/service.go` — `RequestJoin()`. |
| L153 | Group member can create posts/comments? | ✅ | `group/service.go` — `GetPosts()`, plus `topic/` scoped to group. |
| L157 | Event: title, description, day/time, going/not going? | ✅ | `event/event.go` — `Event` entity (Title, Description, EventTime). `EventRSVP` with Going/NotGoing. |
| L161 | Other user can see event and vote? | ✅ | `event/service.go` — `GetGroupEvents()`, `RSVP()`. |

### Chat

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L167 | Private message received in realtime? | ✅ | `chat/transport/ws.go` + `realtime/hub.go`. WebSocket push. |
| L171 | Cannot chat between non-following users? | ✅ | `chat/service.go` — `InitChat()` checks follow relationship (imports `follow/`). |
| L175 | Chat doesn't crash server? | ✅ | `realtime/client.go` — ReadPump/WritePump with `defer recover()`. |
| L179 | Only targeted user receives PM? | ✅ | `realtime/hub.go` routes to specific client connection. |
| L183 | Group chat: all group members receive in realtime? | ✅ | `group/transport/ws.go` + `realtime/hub.go`. `group_chat_messages` table in `group/store/`. |
| L187 | Group chat doesn't crash? | ✅ | Same panic recovery as above. |
| L189 | Emojis via chat? | ✅ | UTF-8 text field — emojis are just text. No special handling needed. |

### Notifications

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L193 | Notifications visible on every page? | ✅ | `notification/transport/http.go` — REST endpoints + SSE stream. Frontend navbar component. |
| L197 | Notification: follow request? | ✅ | RabbitMQ routing key `follow.requested` → notification service. |
| L201 | Notification: group invitation? | ✅ | RabbitMQ routing key `group.invited` → notification service. |
| L205 | Notification: group join request? | ✅ | RabbitMQ routing key `group.join_requested` → notification service. |
| L209 | Notification: event created? | ✅ | RabbitMQ routing key `event.created` → notification service. |

### Docker

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L215 | Two containers (backend + frontend), non-zero sizes? | ⚠️ **GAP** | See Gap #3 below |
| L219 | Accessible via browser after docker? | ✅ | `docker-compose.yml` — backend:8080, frontend:3000. |

### Bonus

| # | Audit Question | Verdict | Where in Proposal 3 |
|---|---------------|---------|---------------------|
| L223 | OAuth (GitHub/external)? | ✅ | `oauth/` feature slice. `pkg/oauth/github/`, `pkg/oauth/google/`. |
| L225 | Migration to fill database (seed)? | ✅ | `platform/database/migrations.go` supports seed migrations. |
| L227 | Unfollow confirmation popup? | ✅ | Frontend concern — `sn-merged-plan.md` Phase 4.4. |
| L229 | Profile privacy change confirmation popup? | ✅ | Frontend concern — `sn-merged-plan.md` Phase 4.4. |
| L231 | Additional notifications? | ✅ | `follow.accepted` notification. RabbitMQ extensible for more. |
| L233 | Build script for images/containers? | ✅ | `Makefile` + `docker-compose.yml`. |

---

## readme.md — Requirement Verification

### Frontend (L16-36)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Use a JS framework (Next.js/Vue/Svelte/Mithril) | ✅ | Next.js — `sn-merged-plan.md` Phase 4. |

### Backend 3-Part Structure (L39-58)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Server: receives requests, entry point | ✅ | `server/server.go` |
| App: listens, retrieves from DB, sends responses | ✅ | Feature `service.go` files |
| Database: organize and store data | ✅ | `platform/database/` + `store/sqlite.go` per feature |
| Sessions and cookies for auth | ✅ | `session/` + `middleware/auth.go` |
| Image handling (JPEG, PNG, GIF) | ✅ | `pkg/imgutil/` + feature handlers |
| WebSocket for realtime | ✅ | `realtime/` + `chat/transport/ws.go` + `group/transport/ws.go` |

### SQLite (L60-66)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Use SQLite as database | ✅ | Primary DB. PostgreSQL as optional secondary. |

### Migrations (L67-99)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Migrations create tables on app run | ✅ | `platform/database/migrations.go` |
| Folder structure: numbered up/down SQL files | ✅ | `db/migrations/000001_*.up.sql` / `*.down.sql` |
| `sqlite.go` for connection + migration apply | ✅ | `platform/database/sqlite.go` + `migrations.go` |

### Docker (L101-121)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Two Docker images (backend + frontend) | ✅ | `Dockerfile` (backend) + `frontend/Dockerfile` |
| Backend container: server logic, handle requests, DB | ✅ | Go binary in container |
| Frontend container: serve HTML/CSS/JS, HTTP to backend | ✅ | Next.js in container |
| Expose necessary ports | ✅ | 8080 (backend), 3000 (frontend) |

### Authentication (L123-141)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Registration: Email, Password, First Name, Last Name, DOB, Avatar (opt), Nickname (opt), About Me (opt) | ✅ | `user/user.go` entity |
| Sessions and cookies | ✅ | `session/` |
| Stay logged in until logout | ✅ | Session persistence |

### Followers (L143-149)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Follow and unfollow | ✅ | `follow/service.go` |
| Follow request → accept/decline | ✅ | `follow_requests` table |
| Public profile: auto-follow (bypass request) | ✅ | `follow/service.go` — checks `IsPrivate` |

### Profile (L151-164)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| User info (all registration fields except password) | ✅ | `user/user.go` |
| User activity (posts) | ✅ | `user/service.go` — `GetActivity()` |
| Followers and following lists | ✅ | `follow/service.go` |
| Public vs private profile | ✅ | `user/user.go` — `IsPrivate` |
| Toggle public/private | ✅ | `user/service.go` — `Update()` |

### Posts (L166-177)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Create posts and comments | ✅ | `topic/`, `comment/` |
| Include image or GIF | ✅ | File upload + `pkg/imgutil/` |
| Privacy: public, almost private, private | ✅ | `topic/topic.go` — `Visibility` enum |
| Private: choose followers who can see | ✅ | `topic_allowed_users` table |

### Groups (L179-199)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Create group (title, description) | ✅ | `group/group.go` |
| Invite users, accept/decline | ✅ | `group_invitations` table |
| Members can invite others | ✅ | `group/service.go` |
| Request to join, creator accepts/declines | ✅ | `group_join_requests` table |
| Browse all groups | ✅ | `group/transport/http.go` — list endpoint |
| Group posts and comments (members only) | ✅ | `group/service.go` — `GetPosts()` |
| Group events: title, description, day/time, going/not going | ✅ | `event/event.go` |
| All users can choose event option | ✅ | `event/service.go` — `RSVP()` |

### Chat (L201-211)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Private messages to followed/following users | ✅ | `chat/service.go` — follow check |
| Realtime via WebSocket | ✅ | `chat/transport/ws.go` + `realtime/hub.go` |
| Emojis | ✅ | UTF-8 text |
| Group chat room | ✅ | `group/transport/ws.go` + `group_chat_messages` table |

### Notifications (L213-226)

| Requirement | Verdict | Notes |
|-------------|---------|-------|
| Visible on every page | ✅ | SSE stream + frontend navbar |
| Follow request notification (private profile) | ✅ | `follow.requested` event |
| Group invitation notification | ✅ | `group.invited` event |
| Group join request notification (creator) | ✅ | `group.join_requested` event |
| Event created notification (group members) | ✅ | `event.created` event |

### Allowed Packages (L229-238)

| Package | Verdict | Notes |
|---------|---------|-------|
| Standard Go packages | ✅ | Used throughout |
| gorilla/websocket | ✅ | `realtime/` |
| golang-migrate / sql-migration / migration | ✅ | `platform/database/migrations.go` |
| sqlite3 (mattn/go-sqlite3) | ✅ | `platform/database/sqlite.go` |
| bcrypt (golang.org/x/crypto/bcrypt) | ✅ | `pkg/bcrypt/` |
| gofrs/uuid or google/uuid | ✅ | `pkg/uuid/` |

---

## Gaps Found

### Gap #1: Allowed Packages — Redis, RabbitMQ, PostgreSQL Drivers

> [!WARNING]
> The spec lists only: standard Go, gorilla/websocket, migration packages, sqlite3, bcrypt, uuid.
> 
> Proposal 3 adds **Redis** (`github.com/redis/go-redis`), **RabbitMQ** (`github.com/rabbitmq/amqp091-go`), and **PostgreSQL** (`github.com/jackc/pgx`). These are **not on the allowed list**.

**Impact**: The audit question "Has the requirement for the allowed packages been respected?" would fail if Redis/RabbitMQ/PostgreSQL drivers are imported.

**Resolution options**:
1. **Make Redis + RabbitMQ + PostgreSQL optional/bonus** — behind build tags or config flags. The core app uses only allowed packages. Redis/RabbitMQ are additive infrastructure for horizontal scaling.
2. **Interpret "allowed packages" loosely** — the spec says these packages are allowed, not that they're the *only* allowed packages. The spec also says "or other package" for migrations, implying flexibility.
3. **Drop Redis/RabbitMQ/PostgreSQL from the core submission** — implement direct function calls for notifications (current pattern) and in-memory rate limiting. Add Redis/RabbitMQ/PostgreSQL as a post-submission enhancement.

**Recommendation**: Option 3 for the graded submission. The architecture supports both modes — the `platform/` packages are optional. Build the core with allowed packages only, then add infrastructure services post-grading.

### Gap #2: "Server, App, Database" Naming vs Vertical Slices

> [!IMPORTANT]
> The audit explicitly asks: "Does the backend include a clear separation of responsibilities among its three major parts - **Server, App, and Database**?"
> 
> Vertical slices merge App + Database into feature packages. The auditor may not see the familiar 3-layer names.

**Why this is still compliant**: The separation exists, just organized differently:

| Audit's 3 parts | Proposal 3 equivalent |
|------------------|-----------------------|
| **Server** | `server/server.go` — receives requests, entry point |
| **App** | Each feature's `service.go` — core business logic |
| **Database** | Each feature's `store/sqlite.go` — data access |

The separation of **responsibilities** is maintained. The files are just co-located by feature instead of by layer. An auditor looking at any feature directory sees exactly 3 files mapping to Server/App/Database.

**Mitigation**: Add a `README.md` or `ARCHITECTURE.md` in the project root explaining the mapping.

### Gap #3: Docker — "Two Containers" vs Four Services

> [!IMPORTANT]
> The audit asks: "Can you confirm that there are **two containers** (backend and frontend)?"
> 
> Proposal 3's docker-compose has **four services**: backend, frontend, redis, rabbitmq.

**Why this is still compliant**: The spec says "two Docker images, one for the backend and another for the frontend." Redis and RabbitMQ are standard infrastructure — they use official images, not custom ones. The project still has exactly **two custom Docker images** (backend + frontend). The auditor runs `docker ps -a` and sees all containers; the two custom ones are clearly labeled.

**Mitigation**: Ensure the docker-compose labels clearly distinguish `backend` and `frontend` as the project's containers. Redis/RabbitMQ containers are additive infrastructure.

If using Resolution Option 3 from Gap #1 (drop Redis/RabbitMQ from graded submission), this gap disappears entirely — docker-compose would have exactly 2 services.

---

## Summary

| Category | Total Questions | ✅ Pass | ⚠️ Gap |
|----------|:--------------:|:------:|:------:|
| Functional / Structure | 3 | 2 | 1 |
| Backend | 4 | 3 | 1 |
| Database | 5 | 5 | 0 |
| Authentication | 9 | 9 | 0 |
| Followers | 4 | 4 | 0 |
| Profile | 8 | 8 | 0 |
| Posts | 5 | 5 | 0 |
| Groups | 7 | 7 | 0 |
| Chat | 7 | 7 | 0 |
| Notifications | 5 | 5 | 0 |
| Docker | 2 | 1 | 1 |
| Bonus | 6 | 6 | 0 |
| **readme.md** | all | all | 0 |
| **Total** | **65+** | **62** | **3** |

All 3 gaps are **resolvable without architectural changes** — they're about package selection and naming, not missing functionality. The recommended resolution: build the graded submission using only allowed packages (no Redis/RabbitMQ/PostgreSQL), add an `ARCHITECTURE.md` explaining the Server/App/Database mapping, and keep the `platform/` infrastructure as a post-grading enhancement.
