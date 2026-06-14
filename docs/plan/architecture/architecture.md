# High-Level Architecture & Code Organization

This document provides a concise, high-level overview of the architectural patterns, code organization, and structural boundaries governing the Social Network application.

---

## 1. Guiding Principles

1. **One Pattern, Everywhere**: Maintain strict consistency across features. If multiple approaches exist, we select one standard pattern and enforce it uniformly.
2. **Feature-Based Vertical Slices**: Features are modular, self-contained packages. All domain logic, commands, queries, storage, and transport handlers for a given feature reside together.
3. **Decoupled Infrastructure**: External services (database, event bus, cache) are behind abstract interfaces in `internal/platform/`. Changing the concrete implementation (e.g., SQLite to PostgreSQL, or memory channel to RabbitMQ) requires zero modification to feature slices.

---

## 2. Directory Structure & Code Organization

The project is structured around self-contained vertical slices inside `internal/` and decoupled cross-cutting layers:

```
.
├── cmd/
│   └── server/
│       └── main.go         # Application entry point & service bootstrap configuration
├── db/
│   └── migrations/         # Numbered database up/down migration SQL scripts
├── internal/
│   # ─── Feature Slices (Vertical Slices) ───
│   ├── chat/               # Direct messages & chat presence
│   ├── comment/            # Comment creation and querying
│   ├── event/              # Event management & RSVPs
│   ├── follow/             # Follow relationships & request workflow
│   ├── group/              # Group definitions, memberships, invites, and group chat
│   ├── notification/       # Notification delivery and subscription
│   ├── oauth/              # OAuth state & third-party auth delegation
│   ├── topic/              # Posts and category management
│   ├── user/               # User profiles & activity tracking
│   ├── vote/               # Voting logic for posts and comments
│   #
│   # ─── Cross-Cutting Core ───
│   ├── core/
│   │   ├── middleware/     # Auth checks, CORS, logging, and rate limiting
│   │   ├── realtime/       # WebSocket hub, connection lifecycle, and routing
│   │   ├── server/         # HTTP server configuration & graceful shutdown
│   │   └── session/        # Session tokens & state manager
│   #
│   # ─── Decoupled Platform Abstractions ───
│   ├── platform/
│   │   ├── cache/          # In-memory and Redis cache interfaces
│   │   ├── database/       # DB factory (SQLite/PostgreSQL interface)
│   │   └── eventbus/       # Async event publishing & subscription
│   #
│   # ─── Bootstrap & Config ───
│   ├── bootstrap/          # Composition root (wiring slices and platform services)
│   └── config/             # Config loaders
└── internal/pkg/           # Reusable helper packages (bcrypt, uuid, validator, oauth)
```

---

## 3. Vertical Slice Layout

Each feature slice within `internal/<feature>/` adheres to a strict internal package structure:

```
internal/<feature>/
  ├── <feature>.go       # Domain entity structs and Repository interface definition
  ├── commands.go        # Mutative operations (writes)
  ├── queries.go         # Read-only operations (reads)
  ├── transport/
  │     ├── http.go      # REST HTTP handlers
  │     └── ws.go        # WebSocket connection/message handlers (chat & group chat only)
  └── store/
        └── sqlite.go    # Concrete SQLite repository implementation
```

### Boundary & Dependency Rules

To keep vertical slices clean and decouple business logic from infrastructure details, we enforce strict boundary checks:

1. **No Outward Store/Transport Imports**: The feature root (`commands.go`, `queries.go`, `<feature>.go`) **must not** import its own `transport/` or `store/` packages, nor may it import another feature's `transport/` or `store/` packages.
2. **Platform Interfaces Only**: Feature logic interacts with storage only through its own `Repository` interface and with database connections via the `platform/database.DB` interface.
3. **No Direct Feature-to-Feature Cross-Imports**: Slices communicate with each other through **narrow consumer-defined interfaces** or asynchronously via the **Event Bus**.

---

## 4. Cross-Slice Communication

To prevent circular dependencies and tight coupling, features interact via three defined patterns:

| Integration Type | Strategy | Implementation Example |
|------------------|----------|------------------------|
| **Data References** | ID-only mapping | A `Comment` struct contains an `AuthorID string` rather than embedding a `User` struct. |
| **Synchronous Queries** | Consumer-defined interfaces | `internal/chat/commands.go` defines a narrow local `FollowChecker` interface, which is satisfied by `internal/follow/Service` during bootstrapping. |
| **Asynchronous Effects** | Event Bus pub/sub | `internal/follow/` publishes a `follow.requested` event. `internal/notification/` subscribes to it to dispatch alerts. |

---

## 5. Dependency Graph

Features import only what they need, keeping the graph acyclic:

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

`notification` is never imported by other features. It subscribes to events at boot time, preventing circular dependencies.

---

## 6. Technology Stack & Runtime Infrastructure

### Backend (Go)
- **Database Engine**: Handled via `platform/database.DB`. Defaults to SQLite with Write-Ahead Logging (`WAL`) enabled and a busy timeout configured to prevent locking. Portability for PostgreSQL is built-in.
- **WebSocket Protocol**: Built-in HTTP upgrade routing to `internal/core/realtime/` with token verification on handshake.
- **Asynchronous Processing**: Non-blocking channel-based event bus for localized operations. Portability for RabbitMQ is built-in.

### Frontend (Next.js)
- **Architecture**: Next.js App Router providing server and client-side rendering.
- **Component Library**: **shadcn/ui** is used for core reusable elements (buttons, inputs, dialogs, cards, dropdowns, etc.), providing accessible and customizable components.
- **Styling**: **Tailwind CSS** coupled with Vanilla CSS overrides for the design system (glassmorphism, dark/light themes, customized HSL color palettes, and interactive transitions).
- **Communication**: REST APIs for basic CRUD operations, WebSocket channels for real-time chat, and SSE or WebSockets for live notifications.
