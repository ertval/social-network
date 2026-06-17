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
- [ ] **S0-BE-01:** Go Project Scaffold
- [ ] **S0-BE-04:** Bug Fixes (B1.1, B1.2, B1.5)

### BE-B (Backend B)
- [ ] **S0-BE-02:** Makefile + CI Pipeline
- [ ] **S0-BE-05:** Bug Fixes (B1.3, B1.4, B1.6, B1.7, B1.8)

### SD-QA (System Design/QA)
- [ ] **S0-BE-03:** golangci-lint Config
- [ ] **S0-DEV-01:** Docker Compose Development Environment
- [ ] **S0-DEV-02:** Pre-commit Hooks
- [ ] **S0-DEV-03:** Dev Environment Docs

### FE-A (Frontend A)
- [ ] **S0-FE-01:** Next.js Scaffold + Tooling

### FE-B (Frontend B)
- [ ] **S0-FE-02:** shadcn/ui Components + Layout

---

## Sprint 1: Platform & Core Infrastructure (Week 3–4)
### BE-A (Backend A)
- [ ] **S1-BE-01:** Platform: DB Factory
- [ ] **S1-BE-04:** Custom Migration System
- [ ] **S1-BE-05:** Core: Session Management
- [ ] **S1-BE-07:** Core: Middlewares
- [ ] **S1-BE-10:** Shared: Image Type Verification Utility

### BE-B (Backend B)
- [ ] **S1-BE-02:** Platform: Event Bus
- [ ] **S1-BE-03:** Platform: Cache
- [ ] **S1-BE-06:** Core: Realtime WebSocket Hub
- [ ] **S1-BE-08:** Core: HTTP Server Bootstrap
- [ ] **S1-BE-09:** Shared: Refactor OAuth Packages

### SD-QA (System Design/QA)
- [ ] **S1-BE-11:** Platform: Database Seeding (Gap Fix)
- [ ] **S1-FE-04:** API Mocking Service

### FE-A (Frontend A)
- [ ] **S1-FE-01:** Auth Pages (Login & Registration UI)
- [ ] **S1-FE-02:** API Client Wrapper

### FE-B (Frontend B)
- [ ] **S1-FE-03:** Nav Layout Shell

---

## Sprint 2: User & Topic Features (Week 5–6)
### Joint BE-A & BE-B
- [ ] **S2-BE-JOINT:** Wire User & Topic bootstrap routes

### BE-A (Backend A)
- [ ] **S2-BE-01:** User: Entity & Repository Interface
- [ ] **S2-BE-02:** User: SQLite Store
- [ ] **S2-BE-03:** User: Register Command
- [ ] **S2-BE-04:** User: Login Command
- [ ] **S2-BE-05:** User: Logout Command
- [ ] **S2-BE-06:** User: Update Profile Command
- [ ] **S2-BE-07:** User: Toggle Privacy Command
- [ ] **S2-BE-08:** User: Get Profile Query
- [ ] **S2-BE-09:** User: Get Activity Query
- [ ] **S2-BE-10:** User: List Users Query
- [ ] **S2-BE-11:** User: HTTP Transport Routing

### BE-B (Backend B)
- [ ] **S2-BE-13:** Topic: Entity & Repository Interface
- [ ] **S2-BE-14:** Topic: SQLite Store
- [ ] **S2-BE-15:** Topic: Create Topic Command
- [ ] **S2-BE-16:** Topic: Cast Vote Command
- [ ] **S2-BE-17:** Topic: Get Feed Query
- [ ] **S2-BE-18:** Topic: Get User Topics Query
- [ ] **S2-BE-19:** Topic: Get Topic Query
- [ ] **S2-BE-20:** Topic: Get Votes Query
- [ ] **S2-BE-21:** Topic: HTTP Transport Routing

### SD-QA (System Design/QA)
- [ ] **S2-BE-12:** User Slice: Migration Verification Contract Tests
- [ ] **S2-BE-22:** Topic Slice: Migration Verification Contract Tests
- [ ] **S2-FE-08:** E2E: User Signup to Feed Journey

### FE-A (Frontend A)
- [ ] **S2-FE-01:** Registration Form
- [ ] **S2-FE-02:** Login Page
- [ ] **S2-FE-03:** Profile Page
- [ ] **S2-FE-04:** Privacy Toggle with Confirmation Popup (Bonus)

### FE-B (Frontend B)
- [ ] **S2-FE-05:** Home Feed Page
- [ ] **S2-FE-06:** Post Creation Form
- [ ] **S2-FE-07:** Post Card Component

---

## Sprint 3: Follow, Comment & Notification (Week 7–8)
### BE-A (Backend A)
- [ ] **S3-BE-01:** Follow: Entities & Repository Interface
- [ ] **S3-BE-02:** Follow: SQLite Store
- [ ] **S3-BE-03:** Follow: Follow User Command
- [ ] **S3-BE-04:** Follow: Unfollow User Command
- [ ] **S3-BE-05:** Follow: Accept Request Command
- [ ] **S3-BE-06:** Follow: Decline Request Command
- [ ] **S3-BE-07:** Follow: Get Followers Query
- [ ] **S3-BE-08:** Follow: Get Following Query
- [ ] **S3-BE-09:** Follow: Get Pending Requests Query
- [ ] **S3-BE-10:** Follow: Are Connected Query **P0**
- [ ] **S3-BE-11:** Follow: HTTP Transport Routing

### BE-B (Backend B)
- [ ] **S3-BE-13:** Comment: Entity & Repository Interface
- [ ] **S3-BE-14:** Comment: SQLite Store
- [ ] **S3-BE-15:** Comment: Create Comment Command
- [ ] **S3-BE-16:** Comment: Get Comments Query
- [ ] **S3-BE-17:** Comment: HTTP Transport Routing
- [ ] **S3-BE-19:** Notification: Entity & Repository Interface
- [ ] **S3-BE-20:** Notification: SQLite Store
- [ ] **S3-BE-21:** Notification: Event Bus Consumer
- [ ] **S3-BE-22:** Notification: Mark Read Command
- [ ] **S3-BE-23:** Notification: List Notifications Query
- [ ] **S3-BE-24:** Notification: HTTP Transport Routing
- [ ] **S3-BE-25:** Notification: Old Schema→New Schema Migration

### Joint BE-A & BE-B
- [ ] **S3-BE-JOINT:** Wire Follow, Comment & Notification bootstrap routes

### SD-QA (System Design/QA)
- [ ] **S3-BE-12:** Follow: Event Publishing Verification
- [ ] **S3-BE-18:** Comment Slice: Contract Tests
- [ ] **S3-FE-07:** E2E: Relationships Notifications Flow
- [ ] **S3-FE-08:** E2E: Posts Comments Notification Flow

### FE-A (Frontend A)
- [ ] **S3-FE-01:** Follow Button with Popup
- [ ] **S3-FE-02:** Followers List Pages
- [ ] **S3-FE-03:** Follow Request Notifications

### FE-B (Frontend B)
- [ ] **S3-FE-04:** Comment Section Components
- [ ] **S3-FE-05:** Notifications Panel
- [ ] **S3-FE-06:** Notifications Live Stream

---

## Sprint 4: Group & Event Features (Week 9–10)
### Joint BE-A & BE-B
- [ ] **S4-BE-JOINT:** Wire Group & Event bootstrap routes

### BE-A (Backend A)
- [ ] **S4-BE-01:** Group: Entities & Repository Interface
- [ ] **S4-BE-02:** Group: SQLite Store
- [ ] **S4-BE-03:** Group: Create Group Command
- [ ] **S4-BE-04:** Group: Invite Member Command
- [ ] **S4-BE-05:** Group: Respond Invite Command
- [ ] **S4-BE-06:** Group: Request Join Command
- [ ] **S4-BE-07:** Group: Respond Join Command
- [ ] **S4-BE-08:** Group: Create Post Command
- [ ] **S4-BE-09:** Group: Send Group Message Command
- [ ] **S4-BE-10:** Group: List Groups Query
- [ ] **S4-BE-11:** Group: Get Group Detail Query
- [ ] **S4-BE-12:** Group: Get Group Feed Query
- [ ] **S4-BE-13:** Group: Get Group Chat History Query
- [ ] **S4-BE-14:** Group: HTTP Transport Routing
- [ ] **S4-BE-15:** Group: WS Transport Routing

### BE-B (Backend B)
- [ ] **S4-BE-16:** Event: Entities & Repository Interface
- [ ] **S4-BE-17:** Event: SQLite Store
- [ ] **S4-BE-18:** Event: Create Event Command
- [ ] **S4-BE-19:** Event: RSVP Command
- [ ] **S4-BE-20:** Event: List Group Events Query
- [ ] **S4-BE-21:** Event: HTTP Transport Routing

### SD-QA (System Design/QA)
- [ ] **S4-FE-08:** E2E: Complete Groups Workspace Journey

### FE-A (Frontend A)
- [ ] **S4-FE-01:** Groups Directory Page
- [ ] **S4-FE-02:** Group Profile Page
- [ ] **S4-FE-03:** Group Posts Feed
- [ ] **S4-FE-04:** Group Chat Workspace

### FE-B (Frontend B)
- [ ] **S4-FE-05:** Event Creation Dialog
- [ ] **S4-FE-06:** Events List Component
- [ ] **S4-FE-07:** RSVP Switch Actions

---

## Sprint 5: Chat & OAuth (Week 11–12)
### Joint BE-A & BE-B
- [ ] **S5-BE-JOINT:** Wire Chat & OAuth bootstrap routes

### BE-A (Backend A)
- [ ] **S5-BE-01:** Chat: Entity & Repository Interface
- [ ] **S5-BE-02:** Chat: SQLite Store
- [ ] **S5-BE-03:** Chat: Send Private Message Command
- [ ] **S5-BE-04:** Chat: Get Chat History Query
- [ ] **S5-BE-05:** Chat: List Conversations Query
- [ ] **S5-BE-06:** Chat: HTTP Transport Routing
- [ ] **S5-BE-07:** Chat: WS Transport Routing

### BE-B (Backend B)
- [ ] **S5-BE-09:** OAuth: Entity & Repository Interface
- [ ] **S5-BE-10:** OAuth: SQLite Store
- [ ] **S5-BE-11:** OAuth: Initiate Login Command
- [ ] **S5-BE-12:** OAuth: Callback Processor Command
- [ ] **S5-BE-13:** OAuth: HTTP Transport Routing
- [ ] **S5-BE-14:** OAuth Client: GitHub Implementation
- [ ] **S5-BE-15:** OAuth Client: Google Implementation

### SD-QA (System Design/QA)
- [ ] **S5-BE-08:** Chat Slice: Contract Tests
- [ ] **S5-BE-16:** OAuth Slice: Contract Tests
- [ ] **S5-FE-06:** E2E: Messaging Real-Time Delivery Journey
- [ ] **S5-FE-07:** E2E: GitHub OAuth Sign In

### FE-A (Frontend A)
- [ ] **S5-FE-01:** Chat Feed View
- [ ] **S5-FE-02:** Realtime Live Sockets Hook
- [ ] **S5-FE-03:** Chat Message Bubble Component

### FE-B (Frontend B)
- [ ] **S5-FE-04:** GitHub OAuth Button Integration
- [ ] **S5-FE-05:** Google OAuth Button Integration

---

## Sprint 6: Integration, Cleanup & Polish (Week 13–14)
### BE-A (Backend A)
- [ ] **S6-BE-01:** Clean Legacy Slices: Domain
- [ ] **S6-BE-02:** Clean Legacy Slices: App

### BE-B (Backend B)
- [ ] **S6-BE-03:** Clean Legacy Slices: Infra

### Joint BE-A & BE-B
- [ ] **S6-BE-04:** Bootstrap Wiring

### SD-QA (System Design/QA)
- [ ] **S6-BE-05:** Full Integration Test Suite
- [ ] **S6-BE-06:** Performance Benchmarks
- [ ] **S6-BE-07:** Vertical Slice Boundary Checks
- [ ] **S6-BE-08:** Audit.md Automation Test Suite (Gap Fix)
- [ ] **S6-DEV-01:** Production Docker Setup
- [ ] **S6-DEV-02:** Health Check Endpoints
- [ ] **S6-DEV-03:** Graceful Server Shutdowns
- [ ] **S6-DEV-04:** Twelve-Factor Configurations Mappings
- [ ] **S6-DEV-05:** Docker Smoke Verification Script
- [ ] **S6-FE-01:** Full E2E Test Suite
- [ ] **S6-FE-04:** Accessibility (a11y) Audit
- [ ] **S6-FE-05:** Frontend Performance Audits
- [ ] **S6-FE-07:** E2E Audit.md Playwright Suite (Gap Fix)

### FE-A (Frontend A)
- [ ] **S6-FE-02:** Responsive Design Check

### FE-B (Frontend B)
- [ ] **S6-FE-03:** Components Error Boundaries & Loading States

### Joint FE-A & FE-B
- [ ] **S6-FE-06:** Production Build Validation
