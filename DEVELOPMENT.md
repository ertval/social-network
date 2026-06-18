# 🛠️ Development Environment & Contributor Guide

Welcome to the **Social Network** project! This document outlines the instructions to set up your local development environment, build the project, run tests, and adhere to our team's contribution guidelines.

---

## 📋 Table of Contents
1. [Prerequisite Tools](#-prerequisite-tools)
2. [Step-by-Step Setup](#%EF%B8%8F-step-by-step-setup)
3. [Docker Compose Commands](#-docker-compose-commands)
4. [Local Development Without Docker](#-local-development-without-docker)
5. [Linting & Formatting Checks](#-linting--formatting-checks)
6. [Branch Naming & PR Conventions](#-branch-naming--pr-conventions)
7. [Refactoring & TDD Methodology](#-refactoring--tdd-methodology)

---

## 🛠️ Prerequisite Tools

Before starting, ensure you have the following tools installed on your local machine:

| Tool | Recommended Version | Purpose | Installation Guide |
|---|---|---|---|
| **Go** | `1.24` (or `1.24.4`+) | Backend service and tools execution | [Go Installation](https://go.dev/doc/install) |
| **Bun** | `1.1`+ | Frontend package manager & runtime | [Bun Installation](https://bun.sh) |
| **Docker** | Latest | Containerization for local services | [Docker Desktop](https://www.docker.com/products/docker-desktop/) |
| **openssl** | Latest | Local HTTPS certificate generation | Pre-installed on most Unix systems |

---

## 🚀 Step-by-Step Setup

Follow these steps to configure your local development environment:

### 1. Configure Local Environment
Copy the example environment file and adjust variables as needed:
```bash
cp .env.example .env
```

### 2. Generate Local SSL/TLS Certificates
Our local development environment uses HTTPS/WSS. Generate the development TLS certificate files using `openssl` via our helper script:
```bash
sh scripts/makecerts.sh
```
> [!IMPORTANT]
> The generated certificates will be stored under `certs/` as defined in your `.env` file (by default `certs/localhost+2.pem` and `certs/localhost+2-key.pem`).

### 3. Install Development Tools
Install backend linter, formatter, vulnerability checkers, and helper CLI utilities:
```bash
make setup
```
This command runs `go install` for `goimports`, `staticcheck`, `golangci-lint`, and `govulncheck`.

### 4. Build and Start the Environment
Launch the local Docker containers (running Go backend, Next.js frontend, and persistent SQLite database):
```bash
make dev
```
> [!NOTE]
> `make dev` is an alias to `make docker-dev` which launches the dev compose configuration.

### 5. Verify Setup (Run CI Checks)
Ensure everything compiles, formats, and tests successfully:
```bash
make ci
```

---

## 🐳 Docker Compose Commands

We support Docker Compose configurations for both hot-reloading development and production-like builds.

### Start Dev Services (With Hot-Reload)
Starts the backend on port `8080` and frontend on port `3000` with volume mounts to reload on code changes:
```bash
make docker-dev
```
If you need to rebuild the containers during startup:
```bash
make docker-dev-build
```

### Stop Dev Services
Tear down running containers:
```bash
make docker-down
```

### Reset Database & Clean Resources
Remove containers, volumes, local SQLite databases, and local images:
```bash
make docker-clean
```
Or to clean just the SQLite database file:
```bash
make db-clean
```

### Run SQL Queries Inside Container
If you need to inspect database state inside the running container, run:
```bash
make docker-db
```
*(This opens the database `db/data/forum.db` with SQLite header output).*

---

## 💻 Local Development Without Docker

If you prefer to run services natively on your host OS:

### Backend Development
1. Verify Go modules:
   ```bash
   go mod tidy
   ```
2. Start the Go server:
   ```bash
   go run cmd/server/main.go
   ```

### Frontend Development
1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```
2. Install frontend dependencies using Bun:
   ```bash
   bun install
   ```
3. Start the Next.js development server:
   ```bash
   bun run dev
   ```
   *(By default, the server runs on `http://localhost:3000`)*

---

## 🔍 Linting & Formatting Checks

Before pushing code or opening a pull request, ensure it passes all verification gates.

### Backend Validation
Run linter (`golangci-lint`, `staticcheck`, and `govulncheck`), formatting, and test checks:
```bash
make ci
```
To run tests separately with code coverage:
```bash
make test
```

### Frontend Validation
Navigate to `frontend/` and run standard format, lint, type checking, and unit tests:
```bash
cd frontend
bun run lint
bun run format:check
tsc --noEmit
bun run test
```

---

## 🌿 Branch Naming & PR Conventions

We enforce a strict branching and commit strategy to maintain a clean git history.

### Branch Naming Convention
Branches must be named in the following format:
```
<username>/<type>-<detail>
```
* **username**: Gitea username (check `origin` remote) — known devs: `ekaramet`, `dkotsi`, `epapamic`, `nwntaspap`, `smichail`
* **type**: Action category (`feat`, `fix`, `chore`, `refactor`, `docs`, `arch`)
* **detail**: kebab-case description of the work. Ticket ID (e.g. `S3-fix-`) may prefix type for traceability, optional.

**Examples:**
- `ekaramet/feat-user-auth`
- `dkotsi/fix-websocket-panic`
- `epapamic/docs-dev-env`

### Pull Request & Commit Strategy
1. **Trunk-Based Development**: Keep feature branches short-lived (≤ 3 days).
2. **Squash Merge**: All branches are squashed into `main` as a single commit.
3. **Conventional Commits**: The squashed commit must follow the [Conventional Commits](https://www.conventionalcommits.org/) format:
   ```
   <type>(<scope>): <description>
   ```
   *Examples:*
   - `feat(user): add register command with age validation`
   - `fix(core): recover from WebSocket goroutine panic`
   - `docs(dev): add onboarding setup instructions`
4. **PR Description**: Include the standard PR description template (found in [general-instructions.md](file://docs/sprints/general-instructions.md)).

---

## 🗺️ Refactoring & TDD Methodology

### Test-Driven Development (TDD)
Implement every command, query, and store method using the **Red-Green-Refactor** pattern:
1. **RED**: Write a failing test first.
2. **GREEN**: Write the minimal code to make the test pass.
3. **REFACTOR**: Clean up code and verify everything remains green.

### Strangler Fig Pattern
For migrating legacy endpoints, follow the Strangler Fig approach:
1. Write contract tests against the legacy API.
2. Build the new slice package in `internal/<slice>/` alongside the legacy code.
3. Verify that the new slice passes identical contract tests.
4. Swap the routing in the composition root (`bootstrap/bootstrap.go`).
5. After a confidence window, delete the old legacy files.
