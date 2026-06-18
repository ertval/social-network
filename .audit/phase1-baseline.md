# Phase 1: Deterministic Baseline

## 1.1 Linting (golangci-lint)
Run with `~/go/bin/golangci-lint run --config scratch/golangci.yml`. Output:
```
internal/app/topics/commands/createTopic.go:1: : # social-network/internal/app/topics/commands [social-network/internal/app/topics/commands.test]
internal/app/topics/commands/createTopic_test.go:87:36: not enough arguments in call to NewCreateTopicHandler
        have (*testhelpers.MockRepository)
        want (topic.Repository, topics.FileStorageManager)
internal/app/topics/commands/createTopic_test.go:100:35: not enough arguments in call to NewCreateTopicHandler
        have (*testhelpers.MockRepository)
        want (topic.Repository, topics.FileStorageManager)
internal/app/topics/commands/updateTopic_test.go:89:36: not enough arguments in call to NewUpdateTopicHandler
        have (*testhelpers.MockRepository)
        want (topic.Repository, topics.FileStorageManager)
internal/app/topics/commands/updateTopic_test.go:102:35: not enough arguments in call to NewUpdateTopicHandler
        have (*testhelpers.MockRepository)
        want (topic.Repository, topics.FileStorageManager) (typecheck)
```

## 1.2 Vulnerability scan (govulncheck)
Run with `~/go/bin/govulncheck ./...`. Output:
- 28 vulnerabilities from the Go standard library (affecting crypto/tls, crypto/x509, net/http, net/mail, encoding/pem, encoding/asn1, net/url).
- Found in standard library crypto/tls@go1.25.1, fixed in go1.25.7 / go1.25.2 / etc.
- No direct vulnerabilities in imported packages.

## 1.3 Go vet
Run with `go vet ./...`. Output:
```
# social-network/internal/app/topics/commands
vet: internal/app/topics/commands/createTopic_test.go:87:40: not enough arguments in call to NewCreateTopicHandler
        have (*testhelpers.MockRepository)
        want (topic.Repository, topics.FileStorageManager)
vet: internal/app/topics/commands/createTopic_test.go:100:39: not enough arguments in call to NewCreateTopicHandler
        have (*testhelpers.MockRepository)
        want (topic.Repository, topics.FileStorageManager)
vet: internal/app/topics/commands/updateTopic_test.go:89:40: not enough arguments in call to NewUpdateTopicHandler
        have (*testhelpers.MockRepository)
        want (topic.Repository, topics.FileStorageManager)
vet: internal/app/topics/commands/updateTopic_test.go:102:39: not enough arguments in call to NewUpdateTopicHandler
        have (*testhelpers.MockRepository)
        want (topic.Repository, topics.FileStorageManager)
```

## 1.4 Module graph
Run with `go mod graph`. Output:
```
social-network github.com/google/uuid@v1.6.0
social-network github.com/gorilla/websocket@v1.5.3
social-network github.com/mattn/go-sqlite3@v1.14.28
social-network go@1.24.4
social-network golang.org/x/crypto@v0.40.0
go@1.24.4 toolchain@go1.24.4
golang.org/x/crypto@v0.40.0 golang.org/x/net@v0.41.0
golang.org/x/crypto@v0.40.0 golang.org/x/sys@v0.34.0
golang.org/x/crypto@v0.40.0 golang.org/x/term@v0.33.0
golang.org/x/crypto@v0.40.0 golang.org/x/text@v0.27.0
golang.org/x/crypto@v0.40.0 go@1.23.0
```

All direct dependencies are in the allowed list:
- `google/uuid` ✓ (allowed: google/uuid or gofrs/uuid)
- `gorilla/websocket` ✓ (allowed)
- `mattn/go-sqlite3` ✓ (allowed)
- `golang.org/x/crypto` ✓ (allowed: bcrypt)

No migration library in use — project uses custom SQL file execution rather than golang-migrate/rubenv/Boostport.

## Project Structure
- Module: `social-network`
- Backend entry: `cmd/server/main.go` → bootstrap → sqlite init → HTTP server
- Frontend entry: `cmd/client/main.go` → proxy/serve static SPA
- Clean Architecture: `internal/domain/` → `internal/app/` → `internal/infra/`
- Domain imports: domain imports only other domain packages (category→topic, topic→comment, oauth→user, session→user). No infra imports in domain. Purity maintained.
