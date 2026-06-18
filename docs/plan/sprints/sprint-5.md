# Sprint 5: Chat & OAuth (Week 11–12)

**Outcome:** 1-on-1 private messaging featuring follow check constraints, emojis, and third-party login delegations (GitHub/Google OAuth) work end-to-end.

> **Old chat commands dropped:** Old code had `markAsRead`, `initChat`, `getChatUsers` WS/RPC methods. These are not mapped to the new slice. `initChat` is replaced by implicit conversation creation on first message (S5-BE-03). `markAsRead` dropped (no per-message read tracking in arch). `getChatUsers` dropped (conversation partner derived from `GET /api/chat/conversations`). Document this in the FE migration notes.

---

### S5-BE-JOINT: Wire Chat & OAuth bootstrap routes
* **Priority:** P0
* **Assignee:** BE-A + BE-B
* **Story Points:** 3
* **Dependencies:** S5-BE-06, S5-BE-07, S5-BE-13
* **Description:** Register new slice routes in `bootstrap.go` so endpoints are live immediately after this sprint.
* **Detailed Steps:**
  1. In `internal/bootstrap/bootstrap.go`, import chat and oauth transport packages.
  2. Call their route registration functions on the HTTP mux and WS router.
  3. Wire OAuth provider clients (github, google) per S5-BE-14/15.
* **Verification:** `go build ./...` passes, new endpoints respond 200/401/403 (not 404).

---

## BE-A (Backend A) Tickets

### S5-BE-01: Chat: Entity & Repository Interface
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Establish domain structures mapping messaging records.
* **Detailed Steps:**
  1. Create `internal/chat/chat.go`.
  2. Define `PrivateMessage` (ID, SenderID, ReceiverID, Content, CreatedAt) and `Repository` interface.
  3. Define a local `FollowChecker` interface containing `AreConnected(ctx context.Context, a, b string) (bool, error)`.
* **Verification:** Compile check `go build ./internal/chat/...`.

---

### S5-BE-02: Chat: SQLite Store
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S5-BE-01
* **Description:** Create SQLite storage queries mapping messages history.
* **Detailed Steps:**
  1. Create `internal/chat/store/sqlite.go`. Implement queries saving messages and loading chats history.
* **Verification:** Integration tests checking message writes and retrieval.

---

### S5-BE-03: Chat: Send Private Message Command
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S5-BE-01
* **Description:** Process messaging delivery. Enforce follow relationship checks and route through realtime sockets.
* **Detailed Steps:**
  1. Create `internal/chat/commands/send_private_msg.go`.
  2. Call local `FollowChecker` to verify that at least one of the users follows the other. If not, reject with error.
  3. Save message to database and dispatch to WebSocket coordinator.
* **Verification:** Unit tests verifying: reject messages between unconnected profiles, allow messages between followers, and verify socket dispatch calls.

---

### S5-BE-04: Chat: Get Chat History Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S5-BE-01
* **Description:** Retrieve historic messages log between active user and partner.
* **Detailed Steps:**
  1. Create `internal/chat/queries/get_chat_history.go`. Check credentials.
* **Verification:** Test log query mapping.

---

### S5-BE-05: Chat: List Conversations Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S5-BE-01
* **Description:** Retrieve distinct list of active chat partners for sidebar listings.
* **Detailed Steps:**
  1. Create `internal/chat/queries/list_conversations.go`.
* **Verification:** Test conversations listing maps correctly.

---

### S5-BE-06: Chat: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S5-BE-03..05
* **Description:** Bind HTTP REST handlers.
* **Detailed Steps:**
  1. Create `internal/chat/transport/http.go`.
  2. Route `GET /api/chat/conversations`, `GET /api/chat/:userId/history`.
* **Verification:** Mock integration HTTP tests.

---

### S5-BE-07: Chat: WS Transport Routing
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 5
* **Dependencies:** S5-BE-06
* **Description:** Migrate old WebSocket chat handlers from `internal/infra/ws/handlers/` into the new vertical slice and bind to core WebSocket hub.
* **Detailed Steps:**
   1. Create `internal/chat/transport/ws.go`.
   2. Move WS message handlers from `internal/infra/ws/handlers/` (chat_send.go, chat_history.go, etc.) into the new slice.
   3. Connect routing events to `internal/core/realtime/` hub.
* **Verification:** Test messaging over connections.

---

## BE-B (Backend B) Tickets

### S5-BE-09: OAuth: Entity & Repository Interface
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 1
* **Description:** Establish domain structures to validate third-party tokens state hashes.
* **Detailed Steps:**
  1. Create `internal/oauth/oauth.go`. Define `OAuthState` (StateString, Provider: Github/Google, CreatedAt) and `Repository` interface.
* **Verification:** Compile checks.

---

### S5-BE-10: OAuth: SQLite Store
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 1
* **Dependencies:** S5-BE-09
* **Description:** Create SQLite query maps for state validations.
* **Detailed Steps:**
  1. Create `internal/oauth/store/sqlite.go`.
* **Verification:** Store integration tests checks.

---

### S5-BE-11: OAuth: Initiate Login Command
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S5-BE-09
* **Description:** Generate state hashes and return third-party login redirection URLs.
* **Detailed Steps:**
  1. Create `internal/oauth/commands/initiate.go`.
  2. Create state string token, save to database, format URL parameters, and redirect client browser.
* **Verification:** Unit tests asserting that unique states write successfully to DB.

---

### S5-BE-12: OAuth: Callback Processor Command
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S5-BE-09
* **Description:** Consume callback queries. Validate states, request token swaps, login/register user, and return cookies.
* **Detailed Steps:**
  1. Create `internal/oauth/commands/callback.go`.
  2. Validate request state against database.
  3. Call provider clients to swap code for access tokens and load user profiles (email, first/last names).
  4. If email is not in db -> register user record. Else -> load user.
  5. Setup session cookie mappings.
* **Verification:** Unit tests mocking provider services verifying register/login logic.

---

### S5-BE-13: OAuth: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S5-BE-11, S5-BE-12
* **Description:** Bind OAuth REST routes.
* **Detailed Steps:**
  1. Create `internal/oauth/transport/http.go`. Route `GET /api/auth/oauth/:provider/init`, `GET /api/auth/oauth/:provider/callback`.
* **Verification:** Integration path calls test.

---

### S5-BE-14: OAuth Client: GitHub Implementation
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Implement GitHub profile load client.
* **Detailed Steps:**
  1. Create `pkg/oauth/github/client.go`. Map credentials exchanges.
* **Verification:** Unit tests using mock web server responses.

---

### S5-BE-15: OAuth Client: Google Implementation
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Implement Google profile load client.
* **Detailed Steps:**
  1. Create `pkg/oauth/google/client.go`. Map credentials exchanges.
* **Verification:** Unit tests using mock web server responses.

---

## Sprint 1 Dependency Note

> **Prerequisite:** S1-BE-09 (OAuth package move to `pkg/oauth/`) must be completed before S5-BE-14/15 (OAuth client implementations). The client files at `pkg/oauth/github/client.go` and `pkg/oauth/google/client.go` assume the Sprint 1 move from `internal/pkg/oAuth/` to `pkg/oauth/`. If Sprint 1 was skipped, extend S5-BE-14/15 to include the move.

## FE-A (Frontend A) Tickets

### S5-FE-01: Chat Feed View
* **Priority:** P0
* **Assignee:** FE-A
* **Story Points:** 5
* **Description:** Implement direct messaging workspace `/chat` displaying conversational partner threads list and chats panel view.
* **Detailed Steps:**
  1. Retrieve conversations list. Render chat cards. Selecting card loads chat pane.
* **Verification:** Visual validation and states checking tests.

---

### S5-FE-02: Realtime Live Sockets Hook
* **Priority:** P0
* **Assignee:** FE-A
* **Story Points:** 5
* **Dependencies:** S5-FE-01
* **Description:** Connect real-time WebSocket messaging handling typing indicators, online presence indicators, and incoming message dispatches.
* **Detailed Steps:**
   1. Connect to websocket. Handle incoming payload types (`chat.message`, `chat.typing`, `chat.presence`).
   2. **Typing indicators:** On keystroke (debounced 500ms), send `chat.typing` WS message with recipientID. On receiving `chat.typing`, show "typing..." bubble for 2s after last event.
   3. **Online presence:** On WS connect/disconnect, broadcast `chat.presence` with status (online/offline). Track via hub client registry. Display green dot on online conversation partners.
   4. Dispatch incoming `chat.message` payloads to chat store and update badge count.
* **Verification:** Playwright message delivery checks. Verify typing bubble appears and disappears. Verify online indicator shows for active WS connections.

---

### S5-FE-03: Chat Message Bubble Component
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 2
* **Dependencies:** S5-FE-01
* **Description:** Render chat text bubble matching timestamps and emoji characters.
* **Detailed Steps:**
  1. Render styled message cells. Support emojis (Unicode formatting).
* **Verification:** Verify HTML characters render correctly.

---

## FE-B (Frontend B) Tickets

### S5-FE-04: GitHub OAuth Button Integration
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Implement GitHub button mapping clicks to initiation pathways.
* **Detailed Steps:**
  1. Add login option. Click routes to `/api/auth/oauth/github/init`.
* **Verification:** Test clicking routes to correct URL.

---

### S5-FE-05: Google OAuth Button Integration
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Dependencies:** S5-FE-04
* **Description:** Implement Google button mapping clicks to initiation pathways.
* **Detailed Steps:**
  1. Add login option. Click routes to `/api/auth/oauth/google/init`.
* **Verification:** Test clicking routes to correct URL.

---

## SD-QA (System Design/QA) Tickets

### S5-BE-08: Chat Slice: Contract Tests
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** S5-BE-07
* **Description:** Verify chat vertical slice compatibility with old domain.
* **Detailed Steps:**
  1. Create `internal/chat/store/sqlite_migration_test.go`.
* **Verification:** Assert equality of returned structures.

---

### S5-BE-16: OAuth Slice: Contract Tests
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 2
* **Dependencies:** S5-BE-13
* **Description:** Ensure OAuth vertical slice compatibility with old domain.
* **Detailed Steps:**
  1. Create `internal/oauth/store/sqlite_migration_test.go`.
* **Verification:** Assert equality of returned structures.

---

### S5-FE-06: E2E: Messaging Real-Time Delivery Journey
* **Priority:** P0
* **Assignee:** SD-QA
* **Story Points:** 3
* **Dependencies:** S5-FE-02
* **Description:** Full E2E Playwright test validating messaging loops.
* **Detailed Steps:**
  1. User A follows User B -> A messages B -> B receives in real-time.
* **Verification:** Runs successfully in CI.

---

### S5-FE-07: E2E: GitHub OAuth Sign In
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 3
* **Dependencies:** S5-FE-04
* **Description:** E2E Playwright testing OAuth mock flows.
* **Detailed Steps:**
  1. Launch Playwright browser -> Click GitHub login -> redirect callback success -> logged in.
* **Verification:** Test validates successfully.
