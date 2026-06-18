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

setup: tools

dev: docker-dev

tools:
	@echo "==> Installing tools..."
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)

# Benchmarking tools (local use only)
bench-tools:
	@echo "==> Installing benchmark tools..."
	go install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_VERSION)
	@echo "For flame graphs, install Graphviz:"
	@echo "  macOS: brew install graphviz"
	@echo "  Ubuntu: sudo apt-get install graphviz"
	@echo "  Windows: choco install graphviz"

# Verify Go modules
ci-mod:
	@echo "==> Verifying Go modules..."
	go mod tidy
	git diff --exit-code go.mod go.sum || \
		(echo "Error: go.mod or go.sum are out of date. Run 'go mod tidy' and commit changes."; exit 1)

# Format Go code
format:
	@echo "==> Formatting code..."
	golangci-lint run --fix --timeout=5m

# Verify formatting
check-format: format
	@echo "==> Checking code formatting..."
	git diff --exit-code || \
		(echo "Error: Files are not formatted. Run 'make format' and commit changes."; exit 1)

# Run staticcheck
staticcheck:
	@echo "==> Running staticcheck..."
	staticcheck ./...

# Run golangci-lint
golangci-lint:
	@echo "==> Running golangci-lint..."
	golangci-lint run --timeout=5m

# Run govulncheck
vulncheck:
	@echo "==> Running govulncheck..."
	govulncheck ./...

lint: staticcheck golangci-lint vulncheck

## -- Testing -- ##
# Run tests with coverage
test:
	@echo "==> Running tests..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

# Run quick tests
test-short:
	go test -short ./...

# Backend CI pipeline
be-ci: ci-mod check-format lint test 

# Frontend CI pipeline
fe-ci:
	@if [ -f frontend/package.json ]; then \
		echo "==> Running frontend CI..."; \
		cd frontend && bun run lint && bun run format:check && tsc --noEmit && bun run test; \
	else \
		echo "==> Skipping frontend CI: frontend not scaffolded yet."; \
	fi

# Full CI pipeline (backend + frontend)
ci: be-ci fe-ci

# Clean artifacts
clean:
	rm -f coverage.out

# Full CI pipeline for benchmarks
ci-bench:
	go test -run=NONE -bench=. -benchmem ./...

# Compare current vs main branch benchmarks
bench-compare:
	@echo "==> Comparing benchmarks (current vs main)..."
	git stash -u
	git checkout main && go test -run=NONE -bench=. -benchmem -count=5 ./... > bench-base.txt
	git checkout - && go test -run=NONE -bench=. -benchmem -count=5 ./... > bench-head.txt
	benchstat bench-base.txt bench-head.txt
	git stash pop

# Run benchmarks with CPU/memory profiling
bench-profile:
	@echo "==> Running benchmarks with profiling..."
	go test -run=NONE -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo "CPU profile: cpu.prof"
	@echo "Memory profile: mem.prof"

# Generate interactive CPU flame graph (requires Graphviz)
bench-flame:
	@echo "==> Generating CPU flame graph..."
	go tool pprof -http=:8080 cpu.prof

# Clean profiling files
bench-clean:
	rm -f *.prof bench-*.txt

# Clean database
db-clean:
	@echo "==> Cleaning database..."
	rm -rf db/data

# Reset database
db-reset: db-clean
	@mkdir -p db/data

# Seed database with test data
seed: db-reset
	@echo "==> Seeding database..."
	sqlite3 db/data/forum.db < db/migrations/schema.sql
	sqlite3 db/data/forum.db < db/migrations/indexes.sql
	sqlite3 db/data/forum.db < db/seeds/dev_data.sql

# Local run and development commands
run-backend:
	@echo "==> Running backend..."
	go run cmd/server/main.go

run-frontend:
	@if [ -f frontend/package.json ]; then \
		echo "==> Running frontend (Next.js)..."; \
		cd frontend && bun run dev; \
	else \
		echo "==> Running frontend (Legacy Client Server)..."; \
		go run cmd/client/main.go; \
	fi

run-all:
	@echo "==> Running backend and frontend concurrently..."
	@go run cmd/server/main.go & BACKEND_PID=$$!; \
	if [ -f frontend/package.json ]; then \
		(cd frontend && bun run dev); \
	else \
		go run cmd/client/main.go; \
	fi; \
	kill $$BACKEND_PID 2>/dev/null || true

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

# Show help
help:
	@echo "\033[1mCI Commands:\033[0m"
	@echo "  \033[36msetup\033[0m           Install development tools (alias to tools)"
	@echo "  \033[36mdev\033[0m             Start development environment (alias to docker-dev)"
	@echo "  \033[36mci\033[0m              Run full CI pipeline (BE + FE)"
	@echo "  \033[36mbe-ci\033[0m           Run backend CI pipeline (format, lint, test)"
	@echo "  \033[36mtools\033[0m           Install CI tools (goimports, staticcheck, golangci-lint, govulncheck, gofumpt)"
	@echo "  \033[36mci-mod\033[0m          Verify Go modules"
	@echo "  \033[36mformat\033[0m          Format Go code"
	@echo "  \033[36mcheck-format\033[0m    Verify code formatting"
	@echo "  \033[36mstaticcheck\033[0m     Run staticcheck"
	@echo "  \033[36mgolangci-lint\033[0m   Run golangci-lint"
	@echo "  \033[36mlint\033[0m            Run all linters"
	@echo "  \033[36mtest\033[0m            Run tests with coverage"
	@echo "  \033[36mtest-short\033[0m      Run quick tests"
	@echo "  \033[36mclean\033[0m           Clean artifacts"
	@echo "  \033[36mdb-clean\033[0m        Remove database"
	@echo "  \033[36mdb-reset\033[0m        Wipe local database files"
	@echo "  \033[36mseed\033[0m            Seed database with test data"
	
	@echo "\n\033[1mDevelopment / Running (Local):\033[0m"
	@echo "  \033[36mrun-backend\033[0m     Run backend application"
	@echo "  \033[36mrun-frontend\033[0m    Run frontend application (Next.js or Legacy)"
	@echo "  \033[36mrun-all\033[0m         Run both backend and frontend concurrently"

	@echo "\n\033[1mDocker Commands:\033[0m"
	@echo "  \033[36mdocker-build\033[0m      Build Docker image"
	@echo "  \033[36mdocker-up\033[0m         Start services"
	@echo "  \033[36mdocker-down\033[0m       Stop services"
	@echo "  \033[36mdocker-logs\033[0m       Show logs"
	@echo "  \033[36mdocker-restart\033[0m    Restart services"
	@echo "  \033[36mdocker-ps\033[0m         Show running containers"
	@echo "  \033[36mdocker-clean\033[0m      Remove all Docker resources"
	@echo "  \033[36mdocker-dev\033[0m        Start development environment"
	@echo "  \033[36mdocker-dev-build\033[0m  Build and start dev environment"
	
	@echo "\n\033[1mBenchmarking & Profiling (Local):\033[0m"
	@echo "  \033[36mbench-tools\033[0m     Install benchmark tools (benchstat)"
	@echo "  \033[36mci-bench\033[0m           Run benchmarks"
	@echo "  \033[36mbench-compare\033[0m   Compare benchmarks vs main branch"
	@echo "  \033[36mbench-profile\033[0m   Run benchmarks with CPU/mem profiling"
	@echo "  \033[36mbench-flame\033[0m     Generate CPU flame graph (requires Graphviz)"
	@echo "  \033[36mbench-clean\033[0m     Clean profiling files"
	@echo "\n\033[3mNote: Benchmark commands require 'make bench-tools' and Graphviz for flame graphs\033[0m"

.PHONY: env setup dev tools bench-tools ci-mod format check-format staticcheck golangci-lint vulncheck lint test test-short ci-bench be-ci fe-ci ci clean \
        bench-compare bench-profile bench-flame bench-clean db-clean db-reset seed run-backend run-frontend run-all help \
        docker-build docker-up docker-down docker-logs docker-restart docker-ps docker-clean docker-dev docker-dev-build