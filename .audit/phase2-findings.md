# Phase 2: Layered Codebase Analysis — Findings Summary

## Layer A — Architecture

| ID | Finding | Severity | Location | Category |
|----|---------|----------|----------|----------|
| A-001 | Handler imports storage-layer error for comparison (layer-skip) | MEDIUM | `internal/infra/http/topic/getTopic/getTopicHandler.go:14,102` | Layer violation |
| A-002 | Unused `*sql.DB` field in HTTP Server struct | LOW | `internal/infra/http/server.go:73` | Code smell |
| A-003 | Domain purity: All 23 domain files verified clean | PASS | All `internal/domain/` | Domain |
| A-004 | SQL migration delimiter uses `:` instead of `;` | HIGH | `internal/infra/storage/sqlite/init.go:118` | Migration bug |
| A-005 | Google OAuth route uses GitHub service handler | HIGH | `internal/infra/http/server.go:197` | Logic bug |
| A-006 | No versioned migration numbering | LOW | `db/migrations/` | Migration |
| A-007 | Duplicate indexes in schema.sql and indexes.sql | LOW | `db/migrations/schema.sql:151-152` vs `indexes.sql:2-3` | Migration |

## Layer B — Allowed Packages

| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| B-001 | No authorized migration library used (custom hand-rolled) | MEDIUM | `internal/infra/storage/sqlite/init.go:65-80` |
| B-002 | All external imports in allowed list | PASS | go.mod |
| B-003 | `gorilla/websocket` mis-tagged as `// indirect` in go.mod | LOW | go.mod:11 |

## Layer C — Go Idiom

| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| C-001 | `ctx` passed as Scan destination in oauthRepo | HIGH | `internal/infra/storage/sqlite/oauth/oauthRepo.go:182-183` |
| C-002 | Prepared statement created but never used in UserRegister | MEDIUM | `internal/infra/storage/sqlite/users/userRepo.go:70-76` |
| C-003 | OrderBy string concatenation without whitelist validation on `ORDER` | MEDIUM | `topics/topicRepo.go:414-420`, `categories/categoryRepo.go:68` |
| C-004 | StateManager cleanup goroutine has no stop mechanism (leak) | HIGH | `internal/pkg/oAuth/stateManager.go:43,85-98` |

## Layer D — Security & Functional Spec

### D1 — Registration & Auth
| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| D1-001 | Registration uses `Age` (int) instead of Date of Birth | MEDIUM | `registerhandler.go:23`, `schema.sql:9` |
| D1-002 | Avatar/image upload not supported during registration | MEDIUM | `registerhandler.go:17-25` |
| D1-003 | "About Me" field absent from user model/schema/handlers | MEDIUM | All user files |
| D1-004 | Login error leaks user email/username in response | HIGH | `userRepo.go:138`, `LoginEmailHandler.go:64-73` |
| D1-005 | bcrypt cost=12 | PASS | `encryption.go:9` |
| D1-006 | Cookie flags (HttpOnly, Secure, SameSite, Expires) | PASS | `manager.go:102-123` |
| D1-007 | Session persistence via refresh tokens | PASS | `requireAuthorization.go:59-70` |
| D1-008 | Duplicate registration detection via SQL constraints | PASS | `users/errors.go:23-43` |

### D2 — SQL Injection & SQLite Config
| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| D2-001 | Parameter binding used in ALL queries | PASS | All repositories |
| D2-002 | WAL mode in DSN | PASS | `config.go:142` |
| D2-003 | `_busy_timeout` MISSING from DSN | HIGH | `config.go:142` |
| D2-004 | Only SetMaxOpenConns configured (no SetMaxIdleConns/SetConnMaxLifetime) | MEDIUM | `init.go:59-61` |

### D3 — Profile Privacy & Followers
| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| D3-001 | Follow/follower system COMPLETELY ABSENT | CRITICAL | No code exists |
| D3-002 | Profile privacy toggle (public/private) NOT IMPLEMENTED | HIGH | No code exists |
| D3-003 | Profile display (all registration fields) PARTIAL | MEDIUM | `getMe/handler.go:43-50` |
| D3-004 | Auto-follow on public profiles NOT IMPLEMENTED | HIGH | No code exists |
| D3-005 | Follow request flow (accept/decline) NOT IMPLEMENTED | HIGH | No code exists |
| D3-006 | Server-side privacy enforcement NOT IMPLEMENTED | HIGH | No code exists |

### D4 — Posts & Comments
| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| D4-001 | Create post/comment with auth | PASS | Various |
| D4-002 | Three privacy scopes (public/almost private/private) NOT IMPLEMENTED | HIGH | No privacy column in schema |
| D4-003 | Media MIME validation (JPG/PNG/GIF) | PASS | `validator.go:113-163` |
| D4-004 | Backend privacy enforcement NOT IMPLEMENTED | HIGH | No code exists |

### D5 — Groups & Events
| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| D5-001 | All group features COMPLETELY ABSENT | CRITICAL | No code exists |
| D5-002 | Group chat rooms NOT IMPLEMENTED | HIGH | CHAT_FEATURE.md:38 |

### D6 — WebSocket Chat Security
| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| D6-001 | Handshake token verification before upgrade | PASS | `ws/handler.go:37-43`, middleware |
| D6-002 | Follow-based chat authorization NOT IMPLEMENTED | HIGH | `initChat.go:28-33` |
| D6-003 | Read limits (4096 bytes) | PASS | `client.go:14,46` |
| D6-004 | Read/Write deadlines configured | PASS | `client.go:47-49,77,88` |
| D6-005 | Emoji support (Go UTF-8 strings) | PASS | `message.go:5-12` |
| D6-006 | Private message targeting (recipient-only) | PASS | `chatSend.go:44-55`, `hub.go:176-199` |

### D7 — Notifications
| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| D7-001 | Follow request notification NOT IMPLEMENTED | HIGH | No follow system |
| D7-002 | Group invitation notification NOT IMPLEMENTED | HIGH | No groups |
| D7-003 | Group join request notification NOT IMPLEMENTED | HIGH | No groups |
| D7-004 | New event notification NOT IMPLEMENTED | HIGH | No events |
| D7-005 | Global notification access (SSE streaming) | PASS | `streamNotificationHandler.go:23-94` |
| D7-006 | Notifications vs messages separation | PASS | Separate domain/tables/transport |

## Layer E — Performance

| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| E-001 | No N+1 queries detected | PASS | All repos |
| E-002 | MaxOpenConns=1 (appropriate for SQLite) | PASS | `init.go:60` |
| E-003 | StateManager cleanup goroutine leak (no stop) | HIGH | `stateManager.go:43` |
| E-004 | Image resizing/compression NOT IMPLEMENTED | MEDIUM | No image lib imported |
| E-005 | Missing `_busy_timeout` for write contention | MEDIUM | `config.go:142` |

## Layer F — Bonus

| ID | Finding | Severity | Location |
|----|---------|----------|----------|
| F-001 | OAuth GitHub provider | PRESENT | `githubclient/githubClient.go` |
| F-002 | OAuth Google provider | PRESENT | `googleclient/googleclient.go` |
| F-003 | DB seeding (dev_data.sql) | PRESENT | `db/seeds/dev_data.sql` |
| F-004 | Confirmation popups on unfollow | MISSING | No feature exists |
| F-005 | Confirmation popups on privacy toggle | MISSING | No feature exists |
| F-006 | Docker: 1 container (not 2) | PARTIAL | `docker-compose.yml` |
| F-007 | Extra notification types | MISSING | Only 4 types, 3 active |
| F-008 | StateManager goroutine leak (also E-003) | HIGH | `stateManager.go:43` |
