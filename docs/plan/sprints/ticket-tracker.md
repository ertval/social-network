# Social Network Refactoring — Ticket Tracker Checklist

For project details, refactoring strategies, TDD methodologies, QA plans, and appendices, see [general-instructions.md](general-instructions.md).

For step-by-step ticket instructions, see individual sprint files:
* [Sprint 0: Foundation](sprint-0.md)
* [Sprint 1: Platform & Core Infrastructure](sprint-1.md)
* [Sprint 2: User & Topic Features](sprint-2.md)
* [Sprint 3: Follow, Comment & Notification](sprint-3.md)
* [Sprint 4: Group & Event Features](sprint-4.md)
* [Sprint 5: Chat & OAuth](sprint-5.md)
* [Sprint 6: Integration, Cleanup & Polish](sprint-6.md)

---


## Sprint 0: Foundation (Week 1–2)
### BE-A (Backend A)
- [ ] **S0-BE001:** Go Project Scaffold
- [ ] **S0-BE002:** Bug Fixes (B1.1, B1.2, B1.5)

### BE-B (Backend B)
- [ ] **S0-BE003:** Makefile + CI Pipeline
- [ ] **S0-BE004:** Bug Fixes (B1.3, B1.4, B1.6, B1.7, B1.8)

### FE-A (Frontend A)
- [ ] **S0-FE001:** Next.js Scaffold + Tooling

### FE-B (Frontend B)
- [ ] **S0-FE002:** shadcn/ui Components + Layout

### SD-QA (System Design/QA)
- [ ] **S0-SD001:** golangci-lint Config
- [ ] **S0-SD002:** Docker Compose Development Environment
- [ ] **S0-SD003:** Pre-commit Hooks
- [ ] **S0-SD004:** Dev Environment Docs

---

## Sprint 1: Platform & Core Infrastructure (Week 3–4)
### BE-A (Backend A)
- [ ] **S1-BE005:** Platform: DB Factory
- [ ] **S1-BE006:** Custom Migration System
- [ ] **S1-BE007:** Core: Session Management
- [ ] **S1-BE008:** Core: Middlewares
- [ ] **S1-BE009:** Shared: Image Type Verification Utility

### BE-B (Backend B)
- [ ] **S1-BE010:** Platform: Event Bus
- [ ] **S1-BE011:** Platform: Cache
- [ ] **S1-BE012:** Core: Realtime WebSocket Hub
- [ ] **S1-BE013:** Core: HTTP Server Bootstrap

### FE-A (Frontend A)
- [ ] **S1-FE003:** Auth Pages (Login & Registration UI)
- [ ] **S1-FE004:** API Client Wrapper

### FE-B (Frontend B)
- [ ] **S1-FE005:** Nav Layout Shell

### SD-QA (System Design/QA)
- [ ] **S1-SD005:** Platform: Database Seeding (Gap Fix)
- [ ] **S1-SD006:** API Mocking Service

---

## Sprint 2: User & Topic Features (Week 5–6)
### Joint BE-A & BE-B
- [ ] **S2-BE014:** Wire User & Topic bootstrap routes

### BE-A (Backend A)
- [ ] **S2-BE015:** User: Entity & Repository Interface
- [ ] **S2-BE016:** User: SQLite Store
- [ ] **S2-BE017:** User: Register Command
- [ ] **S2-BE018:** User: Login Command
- [ ] **S2-BE019:** User: Logout Command
- [ ] **S2-BE020:** User: Update Profile Command
- [ ] **S2-BE021:** User: Toggle Privacy Command
- [ ] **S2-BE022:** User: Get Profile Query
- [ ] **S2-BE023:** User: Get Activity Query
- [ ] **S2-BE024:** User: List Users Query
- [ ] **S2-BE025:** User: HTTP Transport Routing

### BE-B (Backend B)
- [ ] **S2-BE026:** Topic: Entity & Repository Interface
- [ ] **S2-BE027:** Topic: SQLite Store
- [ ] **S2-BE028:** Topic: Create Topic Command
- [ ] **S2-BE029:** Topic: Cast Vote Command
- [ ] **S2-BE030:** Topic: Get Feed Query
- [ ] **S2-BE031:** Topic: Get User Topics Query
- [ ] **S2-BE032:** Topic: Get Topic Query
- [ ] **S2-BE033:** Topic: Get Votes Query
- [ ] **S2-BE034:** Topic: HTTP Transport Routing

### FE-A (Frontend A)
- [ ] **S2-FE006:** Registration Form
- [ ] **S2-FE007:** Login Page
- [ ] **S2-FE008:** Profile Page
- [ ] **S2-FE009:** Privacy Toggle with Confirmation Popup (Bonus)

### FE-B (Frontend B)
- [ ] **S2-FE010:** Home Feed Page
- [ ] **S2-FE011:** Post Creation Form
- [ ] **S2-FE012:** Post Card Component

### SD-QA (System Design/QA)
- [ ] **S2-SD007:** User Slice: Migration Verification Contract Tests
- [ ] **S2-SD008:** Topic Slice: Migration Verification Contract Tests
- [ ] **S2-SD009:** Platform: User & Topic Migrations (000002 & 000003)
- [ ] **S2-SD010:** E2E: User Signup to Feed Journey

---

## Sprint 3: Follow, Comment & Notification (Week 7–8)
### BE-A (Backend A)
- [ ] **S3-BE036:** Follow: Entities & Repository Interface
- [ ] **S3-BE037:** Follow: SQLite Store
- [ ] **S3-BE038:** Follow: Follow User Command
- [ ] **S3-BE039:** Follow: Unfollow User Command
- [ ] **S3-BE040:** Follow: Accept Request Command
- [ ] **S3-BE041:** Follow: Decline Request Command
- [ ] **S3-BE042:** Follow: Get Followers Query
- [ ] **S3-BE043:** Follow: Get Following Query
- [ ] **S3-BE044:** Follow: Get Pending Requests Query
- [ ] **S3-BE045:** Follow: Are Connected Query **P0**
- [ ] **S3-BE046:** Follow: HTTP Transport Routing

### BE-B (Backend B)
- [ ] **S3-BE047:** Comment: Entity & Repository Interface
- [ ] **S3-BE048:** Comment: SQLite Store
- [ ] **S3-BE049:** Comment: Create Comment Command
- [ ] **S3-BE050:** Comment: Get Comments Query
- [ ] **S3-BE051:** Comment: HTTP Transport Routing
- [ ] **S3-BE052:** Notification: Entity & Repository Interface
- [ ] **S3-BE053:** Notification: SQLite Store
- [ ] **S3-BE054:** Notification: Event Bus Consumer
- [ ] **S3-BE055:** Notification: Mark Read Command
- [ ] **S3-BE056:** Notification: List Notifications Query
- [ ] **S3-BE057:** Notification: HTTP Transport Routing
- [ ] **S3-BE058:** Notification: Old Schema→New Schema Migration

### Joint BE-A & BE-B
- [ ] **S3-BE035:** Wire Follow, Comment & Notification bootstrap routes

### FE-A (Frontend A)
- [ ] **S3-FE013:** Follow Button with Popup
- [ ] **S3-FE014:** Followers List Pages
- [ ] **S3-FE015:** Follow Request Notifications

### FE-B (Frontend B)
- [ ] **S3-FE016:** Comment Section Components
- [ ] **S3-FE017:** Notifications Panel
- [ ] **S3-FE018:** Notifications Live Stream

### SD-QA (System Design/QA)
- [ ] **S3-SD011:** Follow: Event Publishing Verification
- [ ] **S3-SD012:** Comment Slice: Contract Tests
- [ ] **S3-SD013:** Platform: Follow System Migrations (000004)
- [ ] **S3-SD014:** E2E: Relationships Notifications Flow
- [ ] **S3-SD015:** E2E: Posts Comments Notification Flow

---

## Sprint 4: Group & Event Features (Week 9–10)
### Joint BE-A & BE-B
- [ ] **S4-BE059:** Wire Group & Event bootstrap routes

### BE-A (Backend A)
- [ ] **S4-BE060:** Group: Entities & Repository Interface
- [ ] **S4-BE061:** Group: SQLite Store
- [ ] **S4-BE062:** Group: Create Group Command
- [ ] **S4-BE063:** Group: Invite Member Command
- [ ] **S4-BE064:** Group: Respond Invite Command
- [ ] **S4-BE065:** Group: Request Join Command
- [ ] **S4-BE066:** Group: Respond Join Command
- [ ] **S4-BE067:** Group: Create Post Command
- [ ] **S4-BE068:** Group: Send Group Message Command
- [ ] **S4-BE069:** Group: List Groups Query
- [ ] **S4-BE070:** Group: Get Group Detail Query
- [ ] **S4-BE071:** Group: Get Group Feed Query
- [ ] **S4-BE072:** Group: Get Group Chat History Query
- [ ] **S4-BE073:** Group: HTTP Transport Routing
- [ ] **S4-BE074:** Group: WS Transport Routing

### BE-B (Backend B)
- [ ] **S4-BE075:** Event: Entities & Repository Interface
- [ ] **S4-BE076:** Event: SQLite Store
- [ ] **S4-BE077:** Event: Create Event Command
- [ ] **S4-BE078:** Event: RSVP Command
- [ ] **S4-BE079:** Event: List Group Events Query
- [ ] **S4-BE080:** Event: HTTP Transport Routing

### FE-A (Frontend A)
- [ ] **S4-FE019:** Groups Directory Page
- [ ] **S4-FE020:** Group Profile Page
- [ ] **S4-FE021:** Group Posts Feed
- [ ] **S4-FE022:** Group Chat Workspace

### FE-B (Frontend B)
- [ ] **S4-FE023:** Event Creation Dialog
- [ ] **S4-FE024:** Events List Component
- [ ] **S4-FE025:** RSVP Switch Actions

### SD-QA (System Design/QA)
- [ ] **S4-SD016:** Platform: Group & Event Migrations (000005 & 000006)
- [ ] **S4-SD017:** E2E: Complete Groups Workspace Journey

---

## Sprint 5: Chat & OAuth (Week 11–12)
### Joint BE-A & BE-B
- [ ] **S5-BE081:** Wire Chat & OAuth bootstrap routes

### BE-A (Backend A)
- [ ] **S5-BE082:** Chat: Entity & Repository Interface
- [ ] **S5-BE083:** Chat: SQLite Store
- [ ] **S5-BE084:** Chat: Send Private Message Command
- [ ] **S5-BE085:** Chat: Get Chat History Query
- [ ] **S5-BE086:** Chat: List Conversations Query
- [ ] **S5-BE087:** Chat: HTTP Transport Routing
- [ ] **S5-BE088:** Chat: WS Transport Routing

### BE-B (Backend B)
- [ ] **S5-BE089:** OAuth: Entity & Repository Interface
- [ ] **S5-BE090:** OAuth: SQLite Store
- [ ] **S5-BE091:** OAuth: Initiate Login Command
- [ ] **S5-BE092:** OAuth: Callback Processor Command
- [ ] **S5-BE093:** OAuth: HTTP Transport Routing
- [ ] **S5-BE094:** OAuth Client: GitHub Implementation
- [ ] **S5-BE095:** OAuth Client: Google Implementation
- [ ] **S5-BE096:** Shared: Refactor OAuth Packages

### FE-A (Frontend A)
- [ ] **S5-FE026:** Chat Feed View
- [ ] **S5-FE027:** Realtime Live Sockets Hook
- [ ] **S5-FE028:** Chat Message Bubble Component

### FE-B (Frontend B)
- [ ] **S5-FE029:** GitHub OAuth Button Integration
- [ ] **S5-FE030:** Google OAuth Button Integration

### SD-QA (System Design/QA)
- [ ] **S5-SD018:** Chat Slice: Contract Tests
- [ ] **S5-SD019:** OAuth Slice: Contract Tests
- [ ] **S5-SD020:** E2E: Messaging Real-Time Delivery Journey
- [ ] **S5-SD021:** E2E: GitHub OAuth Sign In

---

## Sprint 6: Integration, Cleanup & Polish (Week 13–14)
### BE-A (Backend A)
- [ ] **S6-BE097:** Clean Legacy Slices: Domain
- [ ] **S6-BE098:** Clean Legacy Slices: App

### BE-B (Backend B)
- [ ] **S6-BE099:** Clean Legacy Slices: Infra

### Joint BE-A & BE-B
- [ ] **S6-BE100:** Bootstrap Wiring

### FE-A (Frontend A)
- [ ] **S6-FE031:** Responsive Design Check

### FE-B (Frontend B)
- [ ] **S6-FE032:** Components Error Boundaries & Loading States

### Joint FE-A & FE-B
- [ ] **S6-FE033:** Production Build Validation

### SD-QA (System Design/QA)
- [ ] **S6-SD022:** Full Integration Test Suite
- [ ] **S6-SD023:** Performance Benchmarks
- [ ] **S6-SD024:** Vertical Slice Boundary Checks
- [ ] **S6-SD025:** Audit.md Automation Test Suite (Gap Fix)
- [ ] **S6-SD026:** Production Docker Setup
- [ ] **S6-SD027:** Health Check Endpoints
- [ ] **S6-SD028:** Graceful Server Shutdowns
- [ ] **S6-SD029:** Twelve-Factor Configurations Mappings
- [ ] **S6-SD030:** Docker Smoke Verification Script
- [ ] **S6-SD031:** Full E2E Test Suite
- [ ] **S6-SD032:** Accessibility (a11y) Audit
- [ ] **S6-SD033:** Frontend Performance Audits
- [ ] **S6-SD034:** E2E Audit.md Playwright Suite (Gap Fix)
