# PR Review: SQLite → PostgreSQL Migration

## What the PR does

- Creates parallel `internal/infra/storage/postgres/` with repos for all 9 domains + session store
- Versioned SQL migrations under `db/postgres/migrations/000001_*` → `000013_*`
- Docker Compose for PG + Redis
- `golang-migrate` CLI in Makefile
- Hardcodes `postgres.InitializeDB` / `postgres.NewRepositories` in bootstrap
- SQLite code moved to `db/sqlite/` but NOT removed; `go-sqlite3` still in go.mod

## Good

| Practice | Why |
|---|---|
| Versioned up/down migrations | Proper rollback, golang-migrate compatible |
| `RETURNING` for IDs | Idiomatic PG, avoids extra round-trips |
| `ON CONFLICT` upsert | Correct PG pattern |
| Docker Compose for PG + Redis | Reproducible dev env |
| Domain interfaces unchanged | Clean separation, no domain leak |

## Critical issues

### 1. `OpenDB()` returns garbage second `*sql.DB`
`postgres/init.go:20`: `func OpenDB(cfg) (*sql.DB, *sql.DB, error)` — second return always nil. Carried over from sqlite. Should be `(*sql.DB, error)`.

### 2. Manual param-position counting — fragile
`topicRepo.go`, `commentRepo.go`, `categoryRepo.go` etc use:
```go
paramPos++
query += fmt.Sprintf(" AND title LIKE $%d", paramPos)
```
One missed `paramPos++` corrupts the query. ~6 methods do this.

### 3. Transaction boilerplate copy-pasted 5×
Identical deferred rollback + error wrapping in `CreateTopic`, `UpdateTopic`, `SendMessage`, `UpdateComment`, etc. Needs a `WithTransaction` helper.

### 4. `PrepareContext` for every single-use query
Adds round-trips for no benefit. Use `ExecContext`/`QueryContext` directly for one-off queries.

### 5. Date formatting in repo layer
`time.Parse(RFC3339, ...).Format("02/01/2006")` in every read method. Presentation concern → move to handlers/serializers.

### 6. Inconsistent error handling
- `users/` has `MapPQError()` — others don't use it
- `categories/` checks `pq` directly via `isUniqueViolation`
- Some return sentinel errors, some return generic wrapped errors
- `lib/pq` is listed as `// indirect` in go.mod

### 7. Zombie config fields (postgres init ignores them)
- `MigrateOnStart`, `SeedOnStart` — do nothing
- `Pragma_Foreign_Keys`, `Pragma_Journal_Mode`, `Path` — SQLite-only
- `OpenConn` default = 1 (was SQLite's max; PG should be ~25)

### 8. Dead code remains
- `internal/infra/storage/sqlite/` — fully present, compiles
- `db/sqlite/` migrations + seeds
- `go-sqlite3` still a direct dependency

### 9. No migration on app startup
`postgres.InitializeDB` just opens a connection. Migrations run separately via `make migrate-up`.

### 10. DB driver hardcoded in main.go
Can't switch without code change. `cfg.Database.Driver` is unused at entry point.

### 11. `syncTopicCategories` does N DELETE + N INSERT
Instead of single `DELETE WHERE NOT IN (...)` + batch insert.

---

## Best Practice Proposal: Multi-DB Support

The domain layer is already cleanly abstracted via interfaces (9 `Repository` + 1 `session.Manager`).  
The `Repositories` struct is identical in both packages. Only the infra wiring needs formalizing.

### Strategy: Runtime provider selection via factory

### Step 1 — Shared `Repositories` struct + `Provider` interface

```go
// internal/infra/storage/storage.go
package storage

type Repositories struct {
    UserRepo         user.Repository
    CategoryRepo     category.Repository
    TopicRepo        topic.Repository
    CommentRepo      comment.Repository
    VoteRepo         vote.Repository
    NotificationRepo notification.Repository
    OauthRepo        oauth.Repository
    ActivityRepo     activity.Repository
    ChatRepo         chat.Repository
}

type Config struct {
    Driver         string // "sqlite3" | "postgres"
    DSN            string // postgres connection string
    Path           string // sqlite file path
    MigrateOnStart bool
    SeedOnStart    bool
    MaxOpenConns   int
}

type Provider interface {
    Open(ctx context.Context, cfg Config) (*sql.DB, error)
    Migrate(ctx context.Context, db *sql.DB, cfg Config) error
    Seed(ctx context.Context, db *sql.DB, env string) error
    NewRepositories(db *sql.DB) *Repositories
    NewSessionManager(db *sql.DB, sessionCfg config.SessionManagerConfig) session.Manager
}
```

### Step 2 — Each driver implements Provider

```go
// internal/infra/storage/sqlite/sqlite.go
type driver struct{}

func (d driver) Open(ctx context.Context, cfg storage.Config) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", cfg.Path+"?_foreign_keys=on&_journal_mode=WAL")
    db.SetMaxOpenConns(1)
    return db, err
}

func (d driver) Migrate(ctx context.Context, db *sql.DB, cfg storage.Config) error {
    return execSQLFile(db, "db/sqlite/migrations/schema.sql")
}

var Provider storage.Provider = driver{}
```

```go
// internal/infra/storage/postgres/postgres.go
import "github.com/golang-migrate/migrate/v4"

type driver struct{}

func (d driver) Open(ctx context.Context, cfg storage.Config) (*sql.DB, error) {
    db, err := sql.Open("postgres", cfg.DSN)
    db.SetMaxOpenConns(cfg.MaxOpenConns) // default 25
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    return db, err
}

func (d driver) Migrate(ctx context.Context, db *sql.DB, _ storage.Config) error {
    m, err := migrate.NewWithDatabaseInstance(
        "file://db/postgres/migrations", "postgres", db,
    )
    if err != nil { return err }
    return m.Up()
}

var Provider storage.Provider = driver{}
```

### Step 3 — Registry, selected by config

```go
// internal/infra/storage/registry.go
var providers = map[string]Provider{
    "sqlite3":  sqlite.Provider,
    "postgres": postgres.Provider,
}

func Open(ctx context.Context, cfg Config) (*sql.DB, Provider, error) {
    p, ok := providers[cfg.Driver]
    if !ok { return nil, nil, fmt.Errorf("unsupported driver: %s", cfg.Driver) }
    db, err := p.Open(ctx, cfg)
    return db, p, err
}
```

### Step 4 — Bootstrap becomes driver-agnostic

```go
func Bootstrap(cfg *config.ServerConfig) *App {
    db, p, _ := storage.Open(ctx, cfgToStorageCfg(cfg.Database))
    if cfg.Database.MigrateOnStart { p.Migrate(ctx, db, ...) }
    if cfg.Database.SeedOnStart    { p.Seed(ctx, db, cfg.Environment) }

    repos := p.NewRepositories(db)
    sessionManager := p.NewSessionManager(db, cfg.SessionManager)
    // ... rest identical ...
}
```

### Step 5 — Clean config

```go
type DatabaseConfig struct {
    Driver         string        // "sqlite3" | "postgres"
    PostgresURL    string        // PG connection string
    Path           string        // SQLite file path
    MigrateOnStart bool
    SeedOnStart    bool
    MaxOpenConns   int           // default 25
    // Removed: Pragma_Foreign_Keys, Pragma_Journal_Mode, OpenConn
}
```

## What this fixes

| Before | After |
|---|---|
| DB hardcoded in `main.go` | Set `DB_DRIVER=postgres` or `DB_DRIVER=sqlite3` |
| Migrations via external CLI | Embedded, auto-run on startup |
| `Repositories` struct duplicated | Single shared type |
| New DB = fork everything | Implement 4 methods → done |
| Zombie config fields | Clean, driver-agnostic |
| App knows about drivers | Only sees `storage.Repositories` |

## Optional: build-tag isolation

```go
// registry_postgres.go  //go:build !nopg
func init() { providers["postgres"] = postgres.Provider }
// registry_sqlite.go    //go:build !nosqlite
func init() { providers["sqlite3"] = sqlite.Provider }
```

Then `go build -tags nopg` drops PG entirely from the binary.

---

## Quick wins (fix in under 10 min each)

1. Fix `OpenDB()` to return `(*sql.DB, error)`
2. Extract `WithTransaction(ctx, db, fn)` helper — eliminates ~50 lines of boilerplate
3. Replace `PrepareContext` with `QueryContext`/`ExecContext` for single-use queries
4. Move date formatting to HTTP layer (repo returns `time.Time`, serialize in handler)
5. Remove `go-sqlite3` from go.mod if PG-only; otherwise keep but add build tags
6. Change `MaxOpenConns` default to 25 and make SQLite force it to 1 internally
