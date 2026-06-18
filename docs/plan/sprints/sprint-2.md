# Sprint 2: User & Topic Features (Week 5–6)

**Outcome:** User account features (register, login, profile, privacy toggle) and Topic features (posts with public/almost_private/private visibility, post creation, voting) work end-to-end. Both frontend and backend implementations are completed.

> **FollowChecker stubs:** S2-BE-08 and S2-BE-17 define local `FollowChecker` interfaces. The Follow slice does not exist until Sprint 3. Until then, inject a stub that always returns `true` (public profiles bypass) or `false` (private profiles blocked, no follow-gating until Sprint 3). Mark with `// TODO: replace with real FollowChecker in Sprint 3`.
>
> **Migration dependencies:** S1-BE-04 must have created `000002_user_profile_fields` before S2-BE-01 User repo works, and `000003_topic_privacy` before S2-BE-13 Topic repo works. Verify migration order — these are implicit dependencies on S1-BE-04.

---

### S2-BE-JOINT: Wire User & Topic bootstrap routes
* **Priority:** P0
* **Assignee:** BE-A + BE-B
* **Story Points:** 3
* **Dependencies:** S2-BE-11, S2-BE-21
* **Description:** Register new slice HTTP routes in `bootstrap.go` so endpoints are live immediately after this sprint. Without this ticket, new slices compile but are unreachable.
* **Detailed Steps:**
  1. In `internal/bootstrap/bootstrap.go`, import `internal/user/transport` and `internal/topic/transport`.
  2. Call their route registration functions on the HTTP mux (e.g. `userTransport.RegisterRoutes(mux)`).
  3. Verify with `curl http://localhost:8080/api/register` and `curl http://localhost:8080/api/feed`.
* **Verification:** `go build ./...` passes, endpoints respond 200/401 (not 404).

---

## BE-A (Backend A) Tickets

### S2-BE-01: User: Entity & Repository Interface
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Description:** Define the domain entity model for User and the repository interface mapping SQLite operations. Absorbs old `domain/activity/` — user's activity (post counts, follower counts) becomes a query on user data.
* **Detailed Steps:**
   1. Create `internal/user/user.go`. Define the `User` struct (ID, Email, PasswordHash, FirstName, LastName, DateOfBirth, Nickname, AboutMe, AvatarPath, IsPrivate, CreatedAt). **Explicitly drop `Age` field** — replaced by `DateOfBirth` for age calculation at runtime.
   2. Define the `Repository` interface specifying required CRUD queries (e.g. `Create`, `GetByID`, `GetByEmail`, `Update`, `TogglePrivacy`, `ListAll`).
* **Verification:** Compile check `go build ./internal/user/...`.

---

### S2-BE-02: User: SQLite Store
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S2-BE-01
* **Description:** Implement the User `Repository` interface using SQLite database operations.
* **Detailed Steps:**
  1. Create `internal/user/store/sqlite.go`. Implement the repository using the `platform/database.DB` interface.
  2. Implement scan functions translating SQLite rows into `User` domain structures.
* **Verification:** Write store integration tests in `sqlite_test.go` utilizing an in-memory SQLite database connection. Run `go test -v ./internal/user/store/...`.

---

### S2-BE-03: User: Register Command
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S2-BE-01
* **Description:** Create the write-use-case handler for user registration with input validation.
* **Detailed Steps:**
  1. Create `internal/user/commands/register.go`.
  2. Implement age validation (must be at least 13 years old), password strength rules, and duplicate email checking.
  3. Hash password using bcrypt. Store optional avatar image using `pkg/imgutil` validation.
* **Verification:** Write table-driven unit tests checking valid registration, underage rejection, duplicate email blocking, and invalid fields.

---

### S2-BE-04: User: Login Command
* **Priority:** P0
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S2-BE-01
* **Description:** Implement credential validation and user session mapping.
* **Detailed Steps:**
  1. Create `internal/user/commands/login.go`.
  2. Query user by email or nickname. Check password matching with bcrypt.
  3. Call session manager to generate a session cookie token.
* **Verification:** Unit tests validating correct credentials login, wrong email rejection, and wrong password lockouts.

---

### S2-BE-05: User: Logout Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 1
* **Dependencies:** S2-BE-01
* **Description:** Handle session termination.
* **Detailed Steps:**
  1. Create `internal/user/commands/logout.go`.
  2. Revoke active session token from the session store.
* **Verification:** Unit test asserting lookups of revoked sessions fail.

---

### S2-BE-06: User: Update Profile Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S2-BE-01
* **Description:** Handle user profile edits (First Name, Last Name, Nickname, About Me, Avatar).
* **Detailed Steps:**
  1. Create `internal/user/commands/update_profile.go`.
  2. Implement input sanitation, optional fields validation, and update db records.
* **Verification:** Test updates to verify information modifies correctly in database.

---

### S2-BE-07: User: Toggle Privacy Command
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S2-BE-01
* **Description:** Implement profile visibility toggle (public/private profiles).
* **Detailed Steps:**
  1. Create `internal/user/commands/toggle_privacy.go`.
  2. Flip `is_private` boolean field in database.
* **Verification:** Unit tests asserting that toggle updates the flag successfully.

---

### S2-BE-08: User: Get Profile Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S2-BE-01
* **Description:** Implement profile retrieval. Block contents for non-followers when profile is private.
* **Detailed Steps:**
  1. Create `internal/user/queries/get_profile.go`.
  2. Accept target UserID and requester UserID.
   3. Define a local `FollowChecker` interface to check if requester follows the profile.
   4. **Stub for Sprint 2:** Inject a `FollowChecker` stub — use `true` (no follow-gating, show all) or `false` (block all private profiles). Annotate with `// TODO: replace with real FollowChecker in Sprint 3`.
   5. If private and not following, return limited details (nickname/avatar only) and a privacy error.
* **Verification:** Unit tests checking: public profile gets full info, private profile followed gets full info, private profile non-followed gets blocked.

---

### S2-BE-09: User: Get Activity Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S2-BE-01
* **Description:** Retrieve list of posts created by the user.
* **Detailed Steps:**
  1. Create `internal/user/queries/get_activity.go`.
  2. Query user's own posts, comments, or votes counts.
* **Verification:** Unit tests asserting correctness of count retrievals.

---

### S2-BE-10: User: List Users Query
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 2
* **Dependencies:** S2-BE-01
* **Description:** List/browse all registered users for exploration.
* **Detailed Steps:**
  1. Create `internal/user/queries/list_users.go`.
  2. Retrieve list of users (excluding sensitive details like passwords).
* **Verification:** Unit tests checking pagination and correct output mapping.

---

### S2-BE-11: User: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-A
* **Story Points:** 3
* **Dependencies:** S2-BE-03..10
* **Description:** Bind all user commands and queries to HTTP handler endpoints.
* **Detailed Steps:**
   1. Create `internal/user/transport/http.go`.
   2. Wire up `POST /api/register`, `POST /api/login`, `POST /api/logout`, `GET /api/users/:id/profile`, `GET /api/users/:id/activity` (maps to S2-BE-09 get_activity), `GET /api/users` (maps to S2-BE-10 list_users), `PUT /api/profile`, `POST /api/profile/privacy`.
* **Verification:** Integration tests verifying status codes and JSON response outputs over mock HTTP requests. Every command and query must have at least one route.

---

## BE-B (Backend B) Tickets

### S2-BE-13: Topic: Entity & Repository Interface
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 2
* **Description:** Define domain entity model for posts/topics, categories, privacy scopes, and votes.
* **Detailed Steps:**
  1. Create `internal/topic/topic.go`.
  2. Define `Topic` entity containing: ID, AuthorID, Content, ImagePath, Visibility (public, almost_private, private), and CreatedAt.
  3. Define `AllowedUser` entity to map which specific users can view private posts.
   4. Define `Vote` entity (UserID, TargetID, TargetType: post, Direction: +1/-1). **Defer `TargetType: comment` to Sprint 3** (comment slice) to avoid hidden `topic → comment` dependency (arch D6).
  5. Define the `Repository` interface.
* **Verification:** Compile check `go build ./internal/topic/...`.

---

### S2-BE-14: Topic: SQLite Store
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S2-BE-13
* **Description:** Implement Topic `Repository` queries in SQLite.
* **Detailed Steps:**
  1. Create `internal/topic/store/sqlite.go`.
  2. Implement storage queries using `platform/database.DB`. Write complex visibility queries checking permissions, followers, and allowed lists.
* **Verification:** Store integration tests using in-memory SQLite checking correct write/read of visibility permissions.

---

### S2-BE-15: Topic: Create Topic Command
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S2-BE-13
* **Description:** Build write-use-case for creating posts with visibility restrictions and file attachments.
* **Detailed Steps:**
  1. Create `internal/topic/commands/create_topic.go`.
  2. Implement input checking. Extract and validate attached images (JPG, PNG, GIF) using magic bytes verification.
  3. Save attached image to path if present. Save visibility permission records.
* **Verification:** Unit tests verifying public post creation, and private post creation with specific allowed users.

---

### S2-BE-16: Topic: Cast Vote Command
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S2-BE-13
* **Description:** Cast upvotes and downvotes on posts.
* **Detailed Steps:**
  1. Create `internal/topic/commands/cast_vote.go`.
  2. Check if vote already exists. If direction is identical, remove vote. If different, update direction.
* **Verification:** Unit tests validating: voting up, changing to down, and canceling votes.

---

### S2-BE-17: Topic: Get Feed Query
* **Priority:** P0
* **Assignee:** BE-B
* **Story Points:** 5
* **Dependencies:** S2-BE-13
* **Description:** Get home feed posts filtered by privacy scopes.
* **Detailed Steps:**
  1. Create `internal/topic/queries/get_feed.go`.
  2. Retrieve list of topics. Accept requester UserID.
   3. Define a local `FollowChecker` interface.
   4. **Stub for Sprint 2:** Inject a `FollowChecker` stub — use `true` (show all visible posts) or annotate with `// TODO: replace with real FollowChecker in Sprint 3`.
   5. Build visibility filter:
     - Public posts are visible to everyone.
     - `almost_private` posts are visible to author and followers.
     - `private` posts are visible to author and explicitly allowed users.
* **Verification:** Unit tests checking that posts with various privacy scopes are correctly filtered out depending on follow status.

---

### S2-BE-18: Topic: Get User Topics Query
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S2-BE-13
* **Description:** Retrieve posts created by a specific user, ensuring visibility checks are enforced.
* **Detailed Steps:**
  1. Create `internal/topic/queries/get_user_topics.go`.
* **Verification:** Unit tests validating privacy check boundaries.

---

### S2-BE-19: Topic: Get Topic Query
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S2-BE-13
* **Description:** Retrieve details for a single post with visibility checks.
* **Detailed Steps:**
  1. Create `internal/topic/queries/get_topic.go`.
* **Verification:** Tests checking access controls.

---

### S2-BE-20: Topic: Get Votes Query
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 2
* **Dependencies:** S2-BE-13
* **Description:** Retrieve total count of upvotes and downvotes for a target post.
* **Detailed Steps:**
  1. Create `internal/topic/queries/get_votes.go`.
* **Verification:** Tests verifying correct sum outputs.

---

### S2-BE-21: Topic: HTTP Transport Routing
* **Priority:** P1
* **Assignee:** BE-B
* **Story Points:** 3
* **Dependencies:** S2-BE-15..20
* **Description:** Setup HTTP routing handlers for posts and votes.
* **Detailed Steps:**
   1. Create `internal/topic/transport/http.go`.
   2. Wire up `POST /api/posts`, `GET /api/feed`, `GET /api/posts/:id`, `GET /api/users/:id/posts` (maps to S2-BE-18 get_user_topics), `GET /api/posts/:id/votes` (maps to S2-BE-20 get_votes), `POST /api/posts/:id/vote`.
* **Verification:** HTTP mock integration tests verifying correct endpoint codes. Every command and query must have at least one route.

---

## FE-A (Frontend A) Tickets

### S2-FE-01: Registration Form
* **Priority:** P0
* **Assignee:** FE-A
* **Story Points:** 5
* **Description:** Implement registration form details, image file selection, and payload submissions.
* **Detailed Steps:**
  1. Bind form inputs to state hooks.
  2. Convert upload image file to multipart form data structure.
  3. Validate forms for local client bounds before posting.
* **Verification:** Playwright interactive flow testing correct submissions.

---

### S2-FE-02: Login Page
* **Priority:** P0
* **Assignee:** FE-A
* **Story Points:** 3
* **Description:** Complete login page layout and credentials request mapping.
* **Detailed Steps:**
  1. Implement login view. Post username/email and password parameters to `/api/login`.
* **Verification:** Check visual styling and test input submits.

---

### S2-FE-03: Profile Page
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 5
* **Description:** Implement `/profile/[id]` layout with follower tallies, user activity posts feed, and private lock screens.
* **Detailed Steps:**
  1. Render user details, follow counts, and followers lists.
  2. Fetch profile data: if locked, show overlay screen with locked lock icon.
  3. Render feed list of posts created by the user.
* **Verification:** Mock api states (locked vs open) and test page output.

---

### S2-FE-04: Privacy Toggle with Confirmation Popup (Bonus)
* **Priority:** P1
* **Assignee:** FE-A
* **Story Points:** 2
* **Description:** Build privacy toggle switch with confirmation dialog box before updating state.
* **Detailed Steps:**
  1. Render switch component. When toggled, pop up confirmation Dialog.
  2. Send request to `/api/profile/privacy` only if confirmed.
* **Verification:** Interactive test confirming cancel ignores, accept updates database.

---

## FE-B (Frontend B) Tickets

### S2-FE-05: Home Feed Page
* **Priority:** P0
* **Assignee:** FE-B
* **Story Points:** 5
* **Description:** Build the main feed page with infinite scroll loading or pagination.
* **Detailed Steps:**
  1. Fetch items from `/api/feed`. Handle loading skeleton states.
* **Verification:** Render page with mock data arrays.

---

### S2-FE-06: Post Creation Form
* **Priority:** P0
* **Assignee:** FE-B
* **Story Points:** 5
* **Description:** Build post creation UI container with image file attachment selector and visibility settings.
* **Detailed Steps:**
  1. Build creation box. Add Select dropdown (public, almost private, private).
  2. If private is chosen, render user selector to pick permitted users.
* **Verification:** Interactive submission check testing parameters sent.

---

### S2-FE-07: Post Card Component
* **Priority:** P1
* **Assignee:** FE-B
* **Story Points:** 3
* **Description:** Render feed post item.
* **Detailed Steps:**
  1. Render header (author name, timestamp), Content, Image if present, and vote counts buttons.
* **Verification:** Test component renders correctly with varying input datasets.

---

## SD-QA (System Design/QA) Tickets

### S2-BE-12: User Slice: Migration Verification Contract Tests
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 3
* **Dependencies:** S0-BE-01 (old repo exists), S2-BE-11 (new slice for verification)
* **Description:** Per Strangler Fig (Step 1 then Step 3), first write contract tests against the OLD repo, then verify the NEW slice passes the same tests.
* **Detailed Steps:**
   1. **(Step 1)** Create `internal/user/store/sqlite_migration_test.go`. Write tests against the old repository (`internal/infra/storage/sqlite/...` queries) to capture current behavior.
   2. **(Step 2)** New slice is built (S2-BE-02).
   3. **(Step 3)** Run same contract tests against the new `internal/user/store/sqlite.go` — assert identical data mapping.
* **Verification:** Contract tests pass with 100% data compatibility against old repo first, then new slice.

---

### S2-BE-22: Topic Slice: Migration Verification Contract Tests
* **Priority:** P1
* **Assignee:** SD-QA
* **Story Points:** 3
* **Dependencies:** S0-BE-01 (old repo exists), S2-BE-21 (new slice for verification)
* **Description:** Per Strangler Fig (Step 1 then Step 3), write contract tests against OLD repo first, then verify NEW slice passes the same tests.
* **Detailed Steps:**
   1. **(Step 1)** Create `internal/topic/store/sqlite_migration_test.go`. Write tests against old topic repository.
   2. **(Step 3)** Run same tests against new vertical slice store.
* **Verification:** Assert equality of returned structures. Old-then-new ordering verified.

---

### S2-FE-08: E2E: User Signup to Feed Journey
* **Priority:** P0
* **Assignee:** SD-QA
* **Story Points:** 3
* **Description:** Core integration Playwright test validating registration, login, and posting to feed.
* **Detailed Steps:**
  1. Playwright scripts: Signup user -> Log in -> Write Post -> Inspect Feed for new post.
* **Verification:** Script executes successfully in headless CI runner.
