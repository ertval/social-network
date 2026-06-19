# Ensure GOPATH/bin is in PATH so installed tools are available
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

MODULE := $(shell go list -m)
GO_PACKAGES := $(shell go list ./...)
GO_FOLDERS := $(shell find . -type d -not -path '*/\.*' -not -path './vendor*')
GO_FILES := $(shell find . -type f -name '*.go' -not -path '*/\.*' -not -path './vendor*')
TOOLS_BIN := $(shell ls $(shell go env GOPATH)/bin)

env:  ## Show environment information
	@echo "=== System Environment ==="
	@uname -a
	@echo ""

	@echo "=== Go Environment ==="
	@go version
	@go env
	@echo ""

	@echo "=== Module ==="
	@echo "$(MODULE)"
	@echo ""

	@echo "=== Packages ==="
	@echo "$(GO_PACKAGES)" | tr ' ' '\n'
	@echo ""

	@echo "=== Folders ==="
	@echo "$(GO_FOLDERS)" | tr ' ' '\n'
	@echo ""

	@echo "=== Go Files ==="
	@echo "$(GO_FILES)" | tr ' ' '\n' | head -20
	@echo "... (showing first 20 files)"
	@echo ""

	@echo "=== Installed Tools ==="
	@echo "$(TOOLS_BIN)" | tr ' ' '\n'
	@echo ""

	@echo "=== PATH ==="
	@echo "$$PATH" | tr ':' '\n'
	@echo ""

	@echo "=== Shell Information ==="
	@echo "SHELL=$$SHELL"
	@echo "BASH=$$BASH"
	@echo "BASH_VERSION=$$BASH_VERSION"
	@echo ""

GOLANGCI_LINT_VERSION = v2.2.1
STATICCHECK_VERSION = latest
GOIMPORTS_VERSION = latest
BENCHSTAT_VERSION = latest
GOVULNCHECK_VERSION = latest
GOFUMPT_VERSION = latest

setup: tools ## Install development tools and dependencies

dev: docker-dev ## Start development environment (alias to docker-dev)

tools: ## Install Go development tools (gofumpt, goimports, etc.)
	@echo "==> Installing tools..."
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)

bench-tools: ## Install benchmarking tools (benchstat)
	@echo "==> Installing benchmark tools..."
	go install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_VERSION)
	@echo "For flame graphs, install Graphviz:"
	@echo "  macOS: brew install graphviz"
	@echo "  Ubuntu: sudo apt-get install graphviz"
	@echo "  Windows: choco install graphviz"

ci-mod: ## Verify that Go modules are tidy
	@echo "==> Verifying Go modules..."
	go mod tidy
	git diff --exit-code go.mod go.sum || \
		(echo "Error: go.mod or go.sum are out of date. Run 'go mod tidy' and commit changes."; exit 1)

# Format Go code
format: ## Format modified Go files using gofumpt and goimports
	@echo "==> Formatting code..."
	@MODIFIED_GO_FILES=$$(git status --porcelain | awk '{print $$2}' | grep '\.go$$' || true); \
	if [ -n "$$MODIFIED_GO_FILES" ]; then \
		echo "Formatting modified Go files: $$MODIFIED_GO_FILES"; \
		gofumpt -w $$MODIFIED_GO_FILES; \
		goimports -w -local $(MODULE) $$MODIFIED_GO_FILES; \
	else \
		echo "No modified Go files to format."; \
	fi

check-format: ## Verify formatting of modified Go files
	@echo "==> Checking code formatting..."
	@MODIFIED_GO_FILES=$$(git status --porcelain | awk '{print $$2}' | grep '\.go$$' || true); \
	if [ -n "$$MODIFIED_GO_FILES" ]; then \
		echo "Checking formatting on modified Go files: $$MODIFIED_GO_FILES"; \
		UNFORMATTED=$$(gofumpt -l $$MODIFIED_GO_FILES || true); \
		UNFORMATTED_IMPORTS=$$(goimports -l -local $(MODULE) $$MODIFIED_GO_FILES || true); \
		if [ -n "$$UNFORMATTED" ] || [ -n "$$UNFORMATTED_IMPORTS" ]; then \
			if [ -n "$$UNFORMATTED" ]; then \
				echo "Error: Go files not formatted with gofumpt:"; \
				echo "$$UNFORMATTED"; \
			fi; \
			if [ -n "$$UNFORMATTED_IMPORTS" ]; then \
				echo "Error: Go files not formatted with goimports:"; \
				echo "$$UNFORMATTED_IMPORTS"; \
			fi; \
			exit 1; \
		fi; \
	else \
		echo "No modified Go files to check."; \
	fi

staticcheck: ## Run staticcheck static analysis
	@echo "==> Running staticcheck..."
	staticcheck ./...

golangci-lint: ## Run golangci-lint static analysis
	@echo "==> Running golangci-lint..."
	golangci-lint run --timeout=5m

vulncheck: ## Run govulncheck security analysis
	@echo "==> Running govulncheck..."
	govulncheck ./... || (echo "Warning: govulncheck found vulnerabilities"; true)

lint: staticcheck golangci-lint vulncheck ## Run all static analysis checks (staticcheck + golangci-lint + govulncheck)

test: ## Run backend tests with race detector and coverage
	@echo "==> Running tests..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

test-short: ## Run quick/short backend tests
	go test -short ./...

be-ci: ci-mod check-format lint test ## Run full backend CI pipeline (check modules, format, lint, test)

fe-ci: ## Run frontend CI pipeline (lint, check format, type-check, test)
	@if [ -f frontend/package.json ]; then \
		echo "==> Running frontend CI..."; \
		cd frontend && bun run lint && bun run format:check && tsc --noEmit && bun run test; \
	else \
		echo "==> Skipping frontend CI: frontend not scaffolded yet."; \
	fi

ci: be-ci fe-ci ## Run complete CI pipeline (backend + frontend)

clean: ## Remove generated coverage and profiling artifacts
	rm -f coverage.out

ci-bench: ## Run performance benchmarks
	go test -run=NONE -bench=. -benchmem ./...

bench-compare: ## Compare local benchmarks against main branch
	@echo "==> Comparing benchmarks (current vs main)..."
	git stash -u
	git checkout main && go test -run=NONE -bench=. -benchmem -count=5 ./... > bench-base.txt
	git checkout - && go test -run=NONE -bench=. -benchmem -count=5 ./... > bench-head.txt
	benchstat bench-base.txt bench-head.txt
	git stash pop



# Docker commands
docker-build:  ## Build Docker image
	@echo "==> Building Docker image..."
	docker-compose build

docker-up:  ## Start services in detached mode
	@echo "==> Starting services..."
	docker-compose up -d

docker-down:  ## Stop and remove containers
	@echo "==> Stopping services..."
	docker-compose down

docker-logs:  ## Show service logs
	docker-compose logs -f

docker-restart:  ## Restart services
	@echo "==> Restarting services..."
	docker-compose restart

docker-ps:  ## Show running containers
	docker-compose ps

docker-clean:  ## Remove containers, volumes, and images
	@echo "==> Cleaning Docker resources..."
	docker-compose down -v --rmi local

docker-dev:  ## Start development environment
	@echo "==> Starting development environment..."
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

docker-dev-build:  ## Build and start development environment
	@echo "==> Building and starting development environment..."
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build

docker-install-sqlite: ## Install SQLite inside Docker container
	@echo "==> Installing SQLite inside Docker container..."
	docker exec -it -u root forum-app apk add --no-cache sqlite

docker-db: ## Access SQLite database inside Docker container
	@echo "==> Seeing users in the database..."
	docker exec -it forum-app sqlite3 -line -header db/data/forum.db

bench-profile: ## Run benchmarks with CPU and memory profiling
	@echo "==> Running benchmarks with profiling..."
	go test -run=NONE -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo "CPU profile: cpu.prof"
	@echo "Memory profile: mem.prof"

bench-flame: ## Generate interactive CPU flame graph
	@echo "==> Generating CPU flame graph..."
	go tool pprof -http=:8080 cpu.prof

bench-clean: ## Clean profiling files
	rm -f *.prof bench-*.txt

db-clean: ## Clean database
	@echo "==> Cleaning database..."
	rm -rf db/data

db-reset: db-clean ## Reset database
	@mkdir -p db/data

seed: db-reset ## Seed database with test data
	@echo "==> Seeding database..."
	sqlite3 db/data/forum.db < db/migrations/schema.sql
	sqlite3 db/data/forum.db < db/migrations/indexes.sql
	sqlite3 db/data/forum.db < db/seeds/dev_data.sql

run-backend: ## Run backend application
	@echo "==> Running backend..."
	go run cmd/server/main.go

run-frontend: ## Run frontend application (Next.js or Legacy)
	@if [ -f frontend/package.json ]; then \
		echo "==> Running frontend (Next.js)..."; \
		cd frontend && bun run dev; \
	else \
		echo "==> Running frontend (Legacy Client Server)..."; \
		go run cmd/client/main.go; \
	fi

run-all: ## Run backend and frontend concurrently
	@echo "==> Running backend and frontend concurrently..."
	@go run cmd/server/main.go & BACKEND_PID=$$!; \
	$(MAKE) run-frontend; \
	kill $$BACKEND_PID 2>/dev/null || true

build-backend: ## Build backend application
	@echo "==> Building backend..."
	go build -o bin/server cmd/server/main.go

build-frontend: ## Build frontend application
	@if [ -f frontend/package.json ]; then \
		echo "==> Building frontend..."; \
		cd frontend && bun run build; \
	else \
		echo "==> Skipping frontend build: frontend not scaffolded yet."; \
	fi

build: build-backend build-frontend ## Build both backend and frontend

# ── Deterministic Review Gates ──────────────────────────────────────

REVIEW_GATES_FAST := ci-mod check-format staticcheck

review-gates: $(REVIEW_GATES_FAST) ## Run all deterministic quality gates (fast subset)
	@echo "  Docs: docs/review-gates-reference.md"
	@echo "✅ All review gates passed"

review-gates-all: review-gates vulncheck ## Run all gates including slower ones (coverage, vulns)
	@echo "✅ All review gates (incl. slow) passed"

# Show help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'

.PHONY: env setup dev tools bench-tools ci-mod format check-format staticcheck golangci-lint vulncheck lint test test-short ci-bench be-ci fe-ci ci clean \
        bench-compare bench-profile bench-flame bench-clean db-clean db-reset seed run-backend run-frontend run-all help \
        build-backend build-frontend build \
        docker-build docker-up docker-down docker-logs docker-restart docker-ps docker-clean docker-dev docker-dev-build