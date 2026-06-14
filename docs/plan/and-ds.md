# PR Review: SQLite → PostgreSQL Migration

## Findings

### What Was Done
- Full postgres repository implementations mirroring every SQLite repo
- 13 `golang-migrate` migration pairs in `db/postgres/migrations/`
- Docker compose for postgres
- Config updated with `PG_URL` / `DB_DRIVER` env vars
- Hardcoded `postgres.NewRepositories(db)` in bootstrap

### Issues

**1. Dead Code Left Behind**
SQLite repos at `internal/infra/storage/sqlite/` are completely orphaned — `sqlite.InitializeDB`, `sqlite.NewRepositories` are never called anywhere. Dead code rots and confuses.

**2. `DB_DRIVER` Config is Ignored**
Config defaults to `"sqlite3"` and reads `DB_DRIVER` from env, but `bootstrap.go` and `cmd/server/main.go` hardcode postgres. The driver config is never dispatched on.

**3. `OpenDB` has Bogus Signature** (`postgres/init.go`)
```go
func OpenDB(cfg config.ServerConfig) (*sql.DB, *sql.DB, error)
```
Second return is always nil. `InitializeDB` passes it through blindly. Copy-paste artifact.

**4. Split Migration Strategy**
- **SQLite**: inline migrations on startup in `init.go`
- **Postgres**: external CLI (`golang-migrate` via Makefile), no auto-migration on startup

**5. Legacy Driver `lib/pq`**
In maintenance mode — Go ecosystem recommends `pgx`. Also listed as `// indirect` in `go.mod`.

**6. SQLite oauthRepo Bug** (`sqlite/oauth/oauthRepo.go:183`)
```go
rows.Scan(ctx, ...)  // ctx passed as scan destination — will panic
```
The kind of rot dead code produces.

**7. No Error Abstraction**
SQLite uses `sqlite3.Error` structs. Postgres uses `pq.Error` with numeric codes. No common `RepositoryError` type for callers.

**8. No Tests**
Zero tests for any postgres repo.

---

## Proposal: Dual-Driver Support (SQLite + PostgreSQL)

### Architecture

```
internal/
  domain/                    # Interfaces unchanged
    user/repository.go       # type Repository interface { ... }
    topic/repository.go
    ...

  infra/storage/
    factory.go               # NEW — dispatches on driver
    postgres/                # Postgres implementations
    sqlite/                  # SQLite implementations
    common/
      errors.go              # Shared error types
      migrations.go          # Shared migration runner
```

### 1. Storage Factory

```go
// internal/infra/storage/factory.go
package storage

import (
    "database/sql"
    "fmt"

    "github.com/arnald/forum/internal/config"
    "github.com/arnald/forum/internal/domain/activity"
    "github.com/arnald/forum/internal/domain/category"
    "github.com/arnald/forum/internal/domain/chat"
    "github.com/arnald/forum/internal/domain/comment"
    "github.com/arnald/forum/internal/domain/notification"
    "github.com/arnald/forum/internal/domain/oauth"
    "github.com/arnald/forum/internal/domain/topic"
    "github.com/arnald/forum/internal/domain/user"
    "github.com/arnald/forum/internal/domain/vote"
    "github.com/arnald/forum/internal/domain/session"
    "github.com/arnald/forum/internal/infra/storage/postgres"
    "github.com/arnald/forum/internal/infra/storage/sqlite"
    postgressession "github.com/arnald/forum/internal/infra/storage/postgres/sessionstore"
    sqlitesession "github.com/arnald/forum/internal/infra/storage/sqlite/sessionstore"
)

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

func NewRepositories(driver string, db *sql.DB) (*Repositories, error) {
    switch driver {
    case "postgres":
        return postgres.NewRepositories(db), nil
    case "sqlite3":
        return sqlite.NewRepositories(db), nil
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", driver)
    }
}

func NewSessionManager(driver string, db *sql.DB, cfg config.SessionManagerConfig) (session.Manager, error) {
    switch driver {
    case "postgres":
        return postgressession.NewSessionManager(db, cfg), nil
    case "sqlite3":
        return sqlitesession.NewSessionManager(db, cfg), nil
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", driver)
    }
}
```

### 2. Config-Aware Initialization

```go
// internal/infra/storage/init.go
package storage

import (
    "database/sql"
    "fmt"

    "github.com/arnald/forum/internal/config"
    "github.com/arnald/forum/internal/infra/storage/postgres"
    "github.com/arnald/forum/internal/infra/storage/sqlite"
)

func OpenDB(cfg config.ServerConfig) (*sql.DB, error) {
    switch cfg.Database.Driver {
    case "postgres":
        return postgres.OpenDB(cfg)
    case "sqlite3":
        return sqlite.OpenDB(cfg)
    default:
        return nil, fmt.Errorf("unsupported driver: %s", cfg.Database.Driver)
    }
}

func InitializeDB(cfg config.ServerConfig) (*sql.DB, error) {
    db, err := OpenDB(cfg)
    if err != nil {
        return nil, err
    }
    if err := RunMigrations(cfg, db); err != nil {
        return nil, fmt.Errorf("migrations failed: %w", err)
    }
    return db, nil
}
```

Fix postgres `OpenDB`:
```go
func OpenDB(cfg config.ServerConfig) (*sql.DB, error) {
    db, err := sql.Open("postgres", cfg.Database.PostgresURL)
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(cfg.Database.OpenConn)
    return db, nil
}
```

### 3. Unified Migration Runner

Embed `golang-migrate/migrate` as a library in both drivers:

```go
// internal/infra/storage/common/migrations.go
package common

import (
    "database/sql"
    "fmt"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    "github.com/golang-migrate/migrate/v4/source/file"

    "github.com/arnald/forum/internal/config"
)

func RunMigrations(cfg config.ServerConfig, db *sql.DB) error {
    if !cfg.Database.MigrateOnStart {
        return nil
    }

    var sourceURL, dbInstance migrate.Database
    var err error

    switch cfg.Database.Driver {
    case "postgres":
        sourceURL = "file://db/postgres/migrations"
        dbInstance, err = postgres.WithInstance(db, &postgres.Config{})
    case "sqlite3":
        sourceURL = "file://db/sqlite/migrations"
        dbInstance, err = sqlite3.WithInstance(db, &sqlite3.Config{})
    default:
        return fmt.Errorf("unsupported driver: %s", cfg.Database.Driver)
    }
    if err != nil {
        return err
    }

    m, err := migrate.NewWithDatabaseInstance(sourceURL, cfg.Database.Driver, dbInstance)
    if err != nil {
        return err
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    return nil
}
```

### 4. Shared Error Types

```go
// internal/infra/storage/common/errors.go
package common

import (
    "errors"
    "fmt"
)

type ErrCode string

const (
    ErrUniqueViolation  ErrCode = "UNIQUE_VIOLATION"
    ErrNotFound         ErrCode = "NOT_FOUND"
    ErrForeignKeyViolation ErrCode = "FOREIGN_KEY_VIOLATION"
)

type RepositoryError struct {
    Code    ErrCode
    Message string
    Err     error
}

func (e *RepositoryError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
}

func (e *RepositoryError) Unwrap() error {
    return e.Err
}

func NewUniqueViolation(msg string, err error) *RepositoryError {
    return &RepositoryError{Code: ErrUniqueViolation, Message: msg, Err: err}
}

func NewNotFoundError(msg string, err error) *RepositoryError {
    return &RepositoryError{Code: ErrNotFound, Message: msg, Err: err}
}
```

Error mappers per driver:

```go
// internal/infra/storage/postgres/users/errors.go
package users

import (
    "github.com/lib/pq"
    "github.com/arnald/forum/internal/infra/storage/common"
)

func mapPQError(err error) error {
    if pqErr, ok := err.(*pq.Error); ok {
        switch pqErr.Code {
        case "23505":
            return common.NewUniqueViolation("user already exists", err)
        }
    }
    return err
}
```

```go
// internal/infra/storage/sqlite/users/errors.go
package users

import (
    "github.com/mattn/go-sqlite3"
    "github.com/arnald/forum/internal/infra/storage/common"
)

func mapSQLiteError(err error) error {
    if sqliteErr, ok := err.(sqlite3.Error); ok {
        if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
            return common.NewUniqueViolation("user already exists", err)
        }
    }
    return err
}
```

### 5. Updated Bootstrap

```go
// internal/bootstrap/bootstrap.go
import storage "github.com/arnald/forum/internal/infra/storage"

func Bootstrap(db *sql.DB, cfg *config.ServerConfig) *App {
    repos, err := storage.NewRepositories(cfg.Database.Driver, db)
    // handle err
    sessionManager, err := storage.NewSessionManager(cfg.Database.Driver, db, cfg.SessionManager)
    // handle err
    // ...
}
```

### 6. Config Defaults

```go
// internal/config/config.go
Driver: helpers.Env("DB_DRIVER", "postgres"), // was "sqlite3"
```

### Migration Steps

1. Add `golang-migrate/migrate` to `go.mod` as a library dependency
2. Create `internal/infra/storage/factory.go` and `common/` package
3. Convert postgres migrations from flat SQL files to `golang-migrate` format (already in that format)
4. Convert sqlite migrations from inline execution to `golang-migrate` format
5. Update `postgres/init.go` — fix `OpenDB` signature, remove inline migration logic
6. Update `sqlite/init.go` — remove inline migration logic, delegate to `common.RunMigrations`
7. Wire factory into bootstrap
8. Add `storage.InitializeDB` call in `cmd/server/main.go` dispatching on driver
9. Delete unused code paths
10. Write integration tests for both drivers using shared test suite
