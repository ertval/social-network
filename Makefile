# Ensure GOPATH/bin is in PATH so installed tools are available
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

MODULE := $(shell go list -m)

# Tool versions
GOLANGCI_LINT_VERSION := v2.2.1
STATICCHECK_VERSION := latest
GOIMPORTS_VERSION := latest
BENCHSTAT_VERSION := latest
GOVULNCHECK_VERSION := latest
GOFUMPT_VERSION := latest
GOSEC_VERSION := latest
GOARCHLINT_VERSION := latest

# ── Environment ───────────────────────────────────────────────────────

env: ## Show environment information
	@echo "=== System ===" && uname -a
	@echo "=== Go ===" && go version && go env
	@echo "=== Module ===" && echo "$(MODULE)"
	@echo "=== Packages ===" && go list ./... | tr '\n' ' ' && echo ""

dev: docker-dev ## Start development environment (alias)

# ── Tool Installation ─────────────────────────────────────────────────

install: ## Install all dependencies (deterministic, like npm ci)
	@echo "==> Installing Go module dependencies (from go.sum)..."
	go mod download
	@echo "==> Installing root JS tooling (from package-lock.json)..."
	npm ci
	@echo "==> Copying .env.example -> .env (if not exists)..."
	cp -n .env.example .env 2>/dev/null || true
	@echo "==> Generating SSL certificates..."
	sh scripts/makecerts.sh 2>/dev/null || echo "     [skip] certs already exist"
	@echo "==> Installing Go development tools..."
	$(MAKE) tools
	@echo "==> Installing git hooks..."
	$(MAKE) setup-hooks
	@if [ -f frontend/package.json ]; then \
		echo "==> Installing frontend dependencies..."; \
		cd frontend && bun install; \
	else \
		echo "==> [skip] frontend not scaffolded yet"; \
	fi
	@echo ""
	@echo "✅ All dependencies installed. Run 'make dev' to start."

setup: tools setup-hooks ## Install all development tools and hooks

tools: ## Install Go development tools (gofumpt, goimports, staticcheck, golangci-lint, govulncheck, gosec, go-arch-lint)
	@echo "==> Installing Go tools..."
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	go install github.com/fe3dback/go-arch-lint@$(GOARCHLINT_VERSION)

setup-hooks: ## Install Lefthook git hooks
	@echo "==> Installing Lefthook hooks..."
	go install github.com/evilmartians/lefthook/v2@latest
	lefthook install

bench-tools: ## Install benchmarking tools (benchstat)
	@echo "==> Installing benchmark tools..."
	go install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_VERSION)
	@echo "For flame graphs: macOS: brew install graphviz | Ubuntu: sudo apt-get install graphviz"

# ── Code Formatting ───────────────────────────────────────────────────

format: ## Format modified Go files using gofumpt and goimports
	@echo "==> Formatting code..."
	@MODIFIED_GO_FILES=$$(git status --porcelain | awk '{print $$2}' | grep '\.go$$' || true); \
	if [ -n "$$MODIFIED_GO_FILES" ]; then \
		gofumpt -w $$MODIFIED_GO_FILES && goimports -w -local $(MODULE) $$MODIFIED_GO_FILES; \
	else \
		echo "No modified Go files to format."; \
	fi

check-format: ## Verify formatting of modified Go files
	@echo "==> Checking code formatting..."
	@MODIFIED_GO_FILES=$$(git status --porcelain | awk '{print $$2}' | grep '\.go$$' || true); \
	if [ -n "$$MODIFIED_GO_FILES" ]; then \
		UNFORMATTED=$$(gofumpt -l $$MODIFIED_GO_FILES || true); \
		UNFORMATTED_IMPORTS=$$(goimports -l -local $(MODULE) $$MODIFIED_GO_FILES || true); \
		if [ -n "$$UNFORMATTED" ] || [ -n "$$UNFORMATTED_IMPORTS" ]; then \
			[ -n "$$UNFORMATTED" ] && echo "gofumpt errors:" && echo "$$UNFORMATTED"; \
			[ -n "$$UNFORMATTED_IMPORTS" ] && echo "goimports errors:" && echo "$$UNFORMATTED_IMPORTS"; \
			exit 1; \
		fi; \
	else \
		echo "No modified Go files to check."; \
	fi

# ── Linting & Static Analysis ─────────────────────────────────────────

staticcheck: ## Run staticcheck static analysis
	@echo "==> Running staticcheck..." && staticcheck ./...

golangci-lint: ## Run golangci-lint static analysis
	@echo "==> Running golangci-lint..." && golangci-lint run --timeout=5m

vulncheck: ## Run govulncheck security analysis
	@echo "==> Running govulncheck..." && govulncheck ./... || true

lint: staticcheck golangci-lint vulncheck ## Run all static analysis checks

# ── Testing ───────────────────────────────────────────────────────────

test: ## Run backend tests with race detector and coverage
	@echo "==> Running tests..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

test-short: ## Run quick/short backend tests
	go test -short ./...

# ── CI Pipeline ───────────────────────────────────────────────────────

ci-mod: ## Verify that Go modules are tidy
	@echo "==> Verifying Go modules..."
	go mod tidy
	git diff --exit-code go.mod go.sum || \
		(echo "Error: go.mod/go.sum out of date. Run 'go mod tidy'."; exit 1)

be-ci: ci-mod check-format lint test ## Run full backend CI pipeline

fe-ci: ## Run frontend CI pipeline (lint, format:check, type-check, test)
	@if [ -f frontend/package.json ]; then \
		echo "==> Running frontend CI..."; \
		cd frontend && bun run lint && bun run format:check && tsc --noEmit && bun run test; \
	else \
		echo "==> Skipping frontend CI: frontend not scaffolded yet."; \
	fi

ci: be-ci fe-ci ## Run complete CI pipeline (backend + frontend)

review-gates: ## Run all deterministic review gates (JSON output)
	@echo "==> Running review gates..." && go run cmd/gates/main.go --all

review-gates-fast: ci-mod check-format staticcheck ## Run fast quality gates
	@echo "✅ Fast gates passed"

review-gates-all: review-gates vulncheck ## Run all gates including slow ones
	@echo "✅ All review gates passed"

check-arch: ## Run go-arch-lint architectural boundary check
	@echo "==> Running go-arch-lint..." && go-arch-lint check

# ── Performance & Benchmarking ────────────────────────────────────────

ci-bench: ## Run performance benchmarks
	go test -run=NONE -bench=. -benchmem ./...

bench-compare: ## Compare benchmarks against main branch
	@echo "==> Comparing benchmarks (current vs main)..."
	git stash -u
	git checkout main && go test -run=NONE -bench=. -benchmem -count=5 ./... > bench-base.txt
	git checkout - && go test -run=NONE -bench=. -benchmem -count=5 ./... > bench-head.txt
	benchstat bench-base.txt bench-head.txt
	git stash pop

bench-profile: ## Run benchmarks with CPU/memory profiling
	go test -run=NONE -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo "Profiles: cpu.prof mem.prof"

bench-flame: ## Generate interactive CPU flame graph
	go tool pprof -http=:8080 cpu.prof

bench-clean: ## Clean profiling files
	rm -f *.prof bench-*.txt

# ── Build ─────────────────────────────────────────────────────────────

build-backend: ## Build backend application
	@echo "==> Building backend..." && go build -o bin/server cmd/server/main.go

build-frontend: ## Build frontend application
	@if [ -f frontend/package.json ]; then \
		echo "==> Building frontend..."; \
		cd frontend && bun run build; \
	else \
		echo "==> Skipping frontend build: not scaffolded yet."; \
	fi

build: build-backend build-frontend ## Build both backend and frontend

# ── Run ───────────────────────────────────────────────────────────────

run-backend: ## Run backend application
	@echo "==> Running backend..." && go run cmd/server/main.go

run-frontend: ## Run frontend application
	@if [ -f frontend/package.json ]; then \
		echo "==> Running frontend (Next.js)..."; \
		cd frontend && bun run dev; \
	else \
		echo "==> Running frontend (Legacy)..."; \
		go run cmd/client/main.go; \
	fi

run-all: ## Run backend and frontend concurrently
	@echo "==> Running backend and frontend..."
	@trap 'kill 0' EXIT; \
	$(MAKE) run-backend & \
	$(MAKE) run-frontend

# ── Docker ────────────────────────────────────────────────────────────

docker-build: ## Build Docker images
	docker compose build

docker-up: ## Start services in detached mode
	docker compose up -d

docker-down: ## Stop and remove containers
	docker compose down

docker-logs: ## Show service logs
	docker compose logs -f

docker-restart: ## Restart services
	docker compose restart

docker-ps: ## Show running containers
	docker compose ps

docker-clean: ## Remove containers, volumes, and images
	docker compose down -v --rmi local

docker-dev: ## Start development environment with hot-reload
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up

docker-dev-build: ## Build and start development environment
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

docker-db: ## Access SQLite database inside Docker container
	docker exec -it forum-app sqlite3 -line -header db/data/forum.db

# ── Database ──────────────────────────────────────────────────────────

db-clean: ## Clean database files
	@echo "==> Cleaning database..." && rm -rf db/data

db-reset: db-clean ## Reset database
	@mkdir -p db/data

seed: ## Seed database with test data
	@echo "==> Seeding database..." && \
	sqlite3 db/data/forum.db < db/migrations/schema.sql && \
	sqlite3 db/data/forum.db < db/migrations/indexes.sql && \
	sqlite3 db/data/forum.db < db/seeds/dev_data.sql

# ── Cleanup ───────────────────────────────────────────────────────────

clean: ## Remove generated coverage and profiling artifacts
	rm -f coverage.out

# ── Help ──────────────────────────────────────────────────────────────

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'

.PHONY: install env dev setup tools setup-hooks bench-tools \
	format check-format staticcheck golangci-lint vulncheck lint \
	test test-short \
	ci-mod be-ci fe-ci ci review-gates review-gates-fast review-gates-all check-arch \
	ci-bench bench-compare bench-profile bench-flame bench-clean \
	build-backend build-frontend build \
	run-backend run-frontend run-all \
	docker-build docker-up docker-down docker-logs docker-restart docker-ps docker-clean docker-dev docker-dev-build docker-db \
	db-clean db-reset seed clean help
