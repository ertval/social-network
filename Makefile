# Ensure GOPATH/bin is in PATH so installed tools are available
GOBIN := $(shell go env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)

MODULE := $(shell go list -m)

NEW_DIRS := internal/user internal/follow internal/topic internal/comment \
            internal/group internal/event internal/chat internal/notification \
            internal/oauth internal/core internal/platform internal/bootstrap \
            internal/config internal/gates cmd/gates cmd/server

NEW_PKGS := $(addprefix $(MODULE)/, $(NEW_DIRS))

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

env:
	@echo "=== System ===" && uname -a
	@echo "=== Go ===" && go version && go env
	@echo "=== Module ===" && echo "$(MODULE)"
	@echo "=== Packages ===" && go list ./... | tr '\n' ' ' && echo ""

dev: ## Start development environment using Docker Compose with hot-reload
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

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

setup: tools setup-hooks

tools:
	@echo "==> Installing Go tools..."
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	go install github.com/fe3dback/go-arch-lint@$(GOARCHLINT_VERSION)

setup-hooks:
	@echo "==> Installing Lefthook hooks..."
	go install github.com/evilmartians/lefthook/v2@latest
	lefthook install

bench-tools:
	@echo "==> Installing benchmark tools..."
	go install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_VERSION)
	@echo "For flame graphs: macOS: brew install graphviz | Ubuntu: sudo apt-get install graphviz"

# ── Code Formatting ───────────────────────────────────────────────────

format: ## Format all Go files using gofumpt and goimports
	@echo "==> Formatting code..."
	@goimports -w -local $(MODULE) cmd internal
	@gofumpt -w cmd internal
	@golangci-lint run --fix --timeout=5m || true

check-format:
	@echo "==> Checking code formatting..."
	@UNFORMATTED=$$(gofumpt -l cmd internal || true); \
	UNFORMATTED_IMPORTS=$$(goimports -l -local $(MODULE) cmd internal || true); \
	if [ -n "$$UNFORMATTED" ] || [ -n "$$UNFORMATTED_IMPORTS" ]; then \
		[ -n "$$UNFORMATTED" ] && echo "gofumpt errors:" && echo "$$UNFORMATTED"; \
		[ -n "$$UNFORMATTED_IMPORTS" ] && echo "goimports errors:" && echo "$$UNFORMATTED_IMPORTS"; \
		exit 1; \
	fi

check-format-new:
	@echo "==> Checking code formatting (new code)..." && \
	UNFORMATTED=$$(gofumpt -l $(NEW_DIRS) || true); \
	UNFORMATTED_IMPORTS=$$(goimports -l -local $(MODULE) $(NEW_DIRS) || true); \
	if [ -n "$$UNFORMATTED" ] || [ -n "$$UNFORMATTED_IMPORTS" ]; then \
		[ -n "$$UNFORMATTED" ] && echo "gofumpt errors (new code):" && echo "$$UNFORMATTED"; \
		[ -n "$$UNFORMATTED_IMPORTS" ] && echo "goimports errors (new code):" && echo "$$UNFORMATTED_IMPORTS"; \
		exit 1; \
	fi

# ── Linting & Static Analysis ─────────────────────────────────────────

staticcheck:
	@echo "==> Running staticcheck..." && staticcheck ./...

golangci-lint:
	@echo "==> Running golangci-lint..." && golangci-lint run --timeout=5m

vulncheck:
	@echo "==> Running govulncheck..." && govulncheck ./... || true

gosec:
	@echo "==> Running gosec..." && gosec -quiet ./...

lint: staticcheck golangci-lint vulncheck gosec

staticcheck-new:
	@echo "==> Running staticcheck (new code)..." && staticcheck $(NEW_PKGS)

golangci-lint-new:
	@echo "==> Running golangci-lint (new code)..." && golangci-lint run --timeout=5m $(addsuffix /..., $(NEW_DIRS))

vet-new:
	@echo "==> Running go vet (new code)..." && go vet $(NEW_PKGS)

vulncheck-new:
	@echo "==> Running govulncheck (new code)..." && govulncheck $(NEW_PKGS) || true

gosec-new:
	@echo "==> Running gosec (new code)..." && gosec -quiet $(addsuffix /..., $(NEW_DIRS))

lint-new: staticcheck-new golangci-lint-new vet-new vulncheck-new gosec-new

# ── Testing ───────────────────────────────────────────────────────────

test:
	@echo "==> Running tests..."
	@if go test -race -coverprofile=coverage.out -covermode=atomic ./... > test.log 2>&1; then \
		rm -f test.log; \
	else \
		cat test.log; \
		rm -f test.log; \
		exit 1; \
	fi
	@go tool cover -func=coverage.out | grep total

test-short:
	@if go test -short ./... > test.log 2>&1; then \
		rm -f test.log; \
		echo "All tests passed."; \
	else \
		cat test.log; \
		rm -f test.log; \
		exit 1; \
	fi

test-new:
	@echo "==> Running tests (new code)..." && \
	if go test -race -coverprofile=coverage.out -covermode=atomic $(NEW_PKGS) > test.log 2>&1; then \
		rm -f test.log; \
	else \
		cat test.log; \
		rm -f test.log; \
		exit 1; \
	fi
	@go tool cover -func=coverage.out | grep total

# ── CI Pipeline ───────────────────────────────────────────────────────

ci-mod:
	@echo "==> Verifying Go modules..."
	go mod tidy
	git diff --exit-code go.mod go.sum || \
		(echo "Error: go.mod/go.sum out of date. Run 'go mod tidy'."; exit 1)

be-ci: ci-mod check-format lint test

be-ci-new: ci-mod check-format-new lint-new test-new

FE_NEXT_DIR := frontend-next

fe-ci:
	@if [ -d $(FE_NEXT_DIR) ] && [ -f $(FE_NEXT_DIR)/package.json ]; then \
		echo "==> Running frontend-next CI..."; \
		cd $(FE_NEXT_DIR) && bun run lint && bun run format:check && tsc --noEmit && bun run test; \
	elif [ -f frontend/package.json ]; then \
		echo "==> [legacy] Running frontend CI..."; \
		cd frontend && bun run lint && bun run format:check && tsc --noEmit && bun run test; \
	else \
		echo "==> Skipping frontend CI: no frontend scaffolded yet."; \
	fi

ci: be-ci fe-ci

ci-new: be-ci-new fe-ci
gates: ## Run quality gates (build all, verification gates)
	go run cmd/gates/main.go --all

check-arch:
	@echo "==> Running go-arch-lint..." && go-arch-lint check

# ── Performance & Benchmarking ────────────────────────────────────────

ci-bench:
	go test -run=NONE -bench=. -benchmem ./...

bench-compare: ## Compare benchmarks against main branch
	@echo "==> Comparing benchmarks (current vs main)..."
	@git worktree remove -f .git-worktree-main 2>/dev/null || true
	@rm -rf .git-worktree-main
	@git worktree add -d .git-worktree-main main
	@cd .git-worktree-main && go test -run=NONE -bench=. -benchmem -count=5 ./... > ../bench-base.txt
	@git worktree remove -f .git-worktree-main
	@go test -run=NONE -bench=. -benchmem -count=5 ./... > bench-head.txt
	@benchstat bench-base.txt bench-head.txt
	@rm -f bench-base.txt bench-head.txt

bench-profile:
	go test -run=NONE -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo "Profiles: cpu.prof mem.prof"

bench-flame:
	go tool pprof -http=:8080 cpu.prof

bench-clean:
	rm -f *.prof bench-*.txt

# ── Build ─────────────────────────────────────────────────────────────

build-backend:
	@echo "==> Building backend..." && go build -o bin/server cmd/server/main.go

build-frontend:
	@if [ -f frontend/package.json ]; then \
		echo "==> Building frontend..."; \
		cd frontend && bun run build; \
	else \
		echo "==> Skipping frontend build: not scaffolded yet."; \
	fi

build: build-backend build-frontend

# ── Run ───────────────────────────────────────────────────────────────

run-backend:
	@echo "==> Running backend..." && go run cmd/server/main.go

run-frontend:
	@if [ -f frontend/package.json ]; then \
		echo "==> Running frontend (Next.js)..."; \
		cd frontend && bun run dev; \
	else \
		echo "==> Running frontend (Legacy)..."; \
		go run cmd/client/main.go; \
	fi

run: ## Run backend and frontend concurrently (native local development)
	@echo "==> Running backend and frontend..."
	@trap 'kill 0' EXIT; \
	$(MAKE) -s run-backend & \
	$(MAKE) -s run-frontend

run-all: run

# ── Docker ────────────────────────────────────────────────────────────

docker-clean:
	docker compose down -v --rmi local

docker-db:
	docker exec -it forum-app sqlite3 -line -header db/data/forum.db

# ── Database ──────────────────────────────────────────────────────────

db-clean:
	@echo "==> Cleaning database..." && rm -rf db/data

db-reset: db-clean ## Reset and seed SQLite database
	@mkdir -p db/data
	@$(MAKE) -s seed

seed:
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
	format check-format staticcheck golangci-lint vulncheck gosec lint \
	test test-short \
	ci-mod be-ci fe-ci ci review-gates gates check-arch \
	ci-bench bench-compare bench-profile bench-flame bench-clean \
	build-backend build-frontend build \
	run-backend run-frontend run run-all \
	docker-clean docker-db \
	db-clean db-reset seed clean help
