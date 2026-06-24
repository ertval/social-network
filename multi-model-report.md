# Multi-Model Analysis Report: Gates-vs-Legacy Plan

## Date
2026-06-24

## Problem
`make review-gates` fails on legacy code (`internal/app/`, `internal/domain/`, `internal/infra/`, `internal/pkg/`) because `make ci` runs format/lint/test on ALL Go code before custom gates execute. Custom gates (`cmd/gates/main.go`) already correctly skip legacy via `isFeatureSlice()` + `skipDirs`, but never get reached.

## Consensus Findings (3+ models)

### F1 [BUG]: `make ci` blocks custom gates from executing
- **Models**: DeepSeek, Mimo, Big Pickle, Kimi, GLM
- **Detail**: `make review-gates` = `make ci` + `go run cmd/gates/main.go --all`. The `make ci` step runs `check-format`, `lint` (staticcheck+golangci-lint+govulncheck+gosec), `test` on all code. Legacy ~200+ files fail formatting/lint. Custom gates (which correctly exclude legacy) never run.
- **Fix**: Decouple `make review-gates` from `make ci`.

### F2 [SUGGESTION]: Decouple `make review-gates` from `make ci`
- **Models**: DeepSeek, Mimo, Big Pickle, Kimi, GLM
- **Detail**: Change `make review-gates` target to run custom gates directly (already filter by feature slice) + compilation check. `make ci` becomes a full-system check (informational, not blocking PRs).
- **Rationale**: Strangler Fig means legacy code is in transitional state — enforcing new-code rules on it is waste. PRs should gate only new/changed code.

### F3 [SUGGESTION]: Keep `go build ./...` on all code
- **Models**: DeepSeek, Mimo, Big Pickle, Kimi, GLM
- **Detail**: Compilation check on full codebase (including legacy) is low-cost, high-signal. Ensures migration doesn't break builds. Separate from formatting/lint/test enforcement.
- **Implementation**: Add `go build ./...` step to `make review-gates` (or new `make build` target).

### F4 [SUGGESTION]: Define explicit new-code directory list in Makefile
- **Models**: DeepSeek, Mimo, Big Pickle, Kimi, GLM
- **Detail**: Create `NEW_DIRS` or `NEW_PKGS` variable listing vertical slice paths: `internal/user`, `internal/follow`, `internal/topic`, `internal/comment`, `internal/group`, `internal/event`, `internal/chat`, `internal/notification`, `internal/oauth`, `internal/core`, `internal/platform`, `internal/bootstrap`, `internal/config`, `internal/gates`, `cmd/gates`, `cmd/server`.
- **Rationale**: `gofmt`, `staticcheck`, `go vet`, `go test` don't have skip-dir flags. Need explicit path targeting.

### F5 [INSIGHT]: Custom gates already solve the problem at Go level
- **Models**: DeepSeek, Mimo, Big Pickle, Kimi, GLM
- **Detail**: `internal/gates/features.go` `isFeatureSlice()` + `skipDirs` map correctly excludes `domain`, `app`, `infra`, `core`, `platform`, `pkg`, `config`, `bootstrap`, `gates`. The gates (layout, boundaries, DAG, TDD, security) already only scan feature slices. The bottleneck is the Makefile layer.
- **Action**: No changes needed to custom Go gates. Focus on Makefile and tool invocation.

### F6 [BUG]: Branch gate fails on non-conventional existing commits
- **Models**: DeepSeek, Mimo, Big Pickle, Kimi, GLM
- **Detail**: `internal/gates/gate_branch.go` scans all commits on branch. Historical commits before gate adoption fail pattern check.
- **Fix options**: (a) Limit scan to commits since branch base vs main, (b) Exempt commits before a cutoff date, (c) Accept as known noise.

## Unique Findings (1-2 models)

### F7 [SUGGESTION]: Test strategy — intentional violation assertions
- **Models**: Kimi, GLM, Big Pickle
- **Detail**: Create fixture files in legacy dir (`internal/app/.../bad.go`) with intentional formatting/lint errors. Verify `make review-gates` passes. Create fixture in new dir (`internal/user/.../bad.go`) with same violations. Verify `make review-gates` fails. This proves the filter works in both directions.

### F8 [SUGGESTION]: Add `gosec` to custom Go gates `security` check
- **Models**: GLM
- **Detail**: Currently `gosec` runs in `make lint` (blanket, no skip-dirs). Move it into `internal/gates/gate_security.go` which already filters by `isFeatureSlice`.
- **Benefit**: Security scanning scoped to new code only. Legacy vulns are documented and will be fixed during porting.

### F9 [SUGGESTION]: Diff-based formatting scoping
- **Models**: Big Pickle, DeepSeek
- **Detail**: Only check formatting on files changed in current branch vs main (using `git diff --name-only`). Avoids touching legacy files.
- **Caveat**: More complex Makefile logic. Alternative: just exclude legacy dirs from format check.

### F10 [INFO]: Frontend gates already conditional on existence
- **Models**: Big Pickle, DeepSeek
- **Detail**: `Makefile` already gates `fe-ci` behind `if [ -f frontend/package.json ]`. Pattern can be reused for backend new-slice detection.

## Contradictions

| Topic | Model A | Model B | Resolution |
|-------|---------|---------|------------|
| `govulncheck` on legacy | Big Pickle/GLM: keep scanning | DeepSeek: skip legacy | Compromise: run in optional `make be-ci-legacy` or as separate weekly scan, not in PR gate |
| Diff-aware vs. path-targeted | Big Pickle: diff is cleaner | Kimi/Mimo: explicit paths simpler | Start with explicit paths (simpler, less error-prone). Evolve to diff-based if maintenance burden grows |
> WE DONT NEED TO SCAN LEGACY CODE, ONLY NEW CODE, IN NEW MODULES. OLD CODE WILL BE REMOVED IN THE END

## Per-Model Summary

| Model | #Findings | Key Insight |
|-------|-----------|-------------|
| DeepSeek V4 Flash | 11 | Custom gates correct; Makefile is the bottleneck. `review-gates` should bypass `make ci`. |
| Mimo 2.5 | 7 | Static tools (gofmt, staticcheck) lack skip-dir flags — explicit path lists are necessary. |
| Big Pickle | 12 | Strangler Fig demands gate adaptation. `go build ./...` is safe baseline; everything else scoped. |
| Kimi K2.6 | 10 | Two-phase gate: Phase 1 = custom gates (already correct), Phase 2 = scoped CI on new dirs only. |
| GLM 5.1 | 10 | `NEW_PKGS` Makefile variable drives all scoped tools. Test fixtures prove both directions. |

## Recommended Plan

### Step 1: Decouple `review-gates` from `make ci`

`make review-gates` no longer depends on `make ci`. It runs:
1. `go build ./...` — compiles all code (legacy + new), zero-cost safety net
2. Custom Go gates — `go run cmd/gates/main.go --all` (already skip legacy)
3. Format + lint + vet + test — scoped to `$(NEW_DIRS)`/`$(NEW_PKGS)` only

`make be-ci` becomes scoped exclusively to new-code dirs. `make ci` (full legacy+new) is available but NOT a gate dependency.

```makefile
NEW_DIRS := internal/user internal/follow internal/topic internal/comment \
            internal/group internal/event internal/chat internal/notification \
            internal/oauth internal/core internal/platform internal/bootstrap \
            internal/config internal/gates cmd/gates cmd/server

NEW_PKGS := $(addprefix $(MODULE)/, $(NEW_DIRS))

# ── Scoped CI (new code only) ─────────────────────────────────────────

be-ci-new: ci-mod check-format-new lint-new test-new
check-format-new:
	@UNFORMATTED=$$(gofumpt -l $(NEW_DIRS) || true); \
	UNFORMATTED_IMPORTS=$$(goimports -l -local $(MODULE) $(NEW_DIRS) || true); \
	if [ -n "$$UNFORMATTED" ] || [ -n "$$UNFORMATTED_IMPORTS" ]; then \
		[ -n "$$UNFORMATTED" ] && echo "gofumpt errors:" && echo "$$UNFORMATTED"; \
		[ -n "$$UNFORMATTED_IMPORTS" ] && echo "goimports errors:" && echo "$$UNFORMATTED_IMPORTS"; \
		exit 1; \
	fi
lint-new: staticcheck-new golangci-lint-new vet-new vulncheck-new gosec-new
staticcheck-new: ; staticcheck $(NEW_PKGS)
golangci-lint-new: ; golangci-lint run --timeout=5m $(NEW_PKGS)
vet-new: ; go vet $(NEW_PKGS)
vulncheck-new: ; govulncheck $(NEW_PKGS)
gosec-new: ; gosec -quiet $(NEW_DIRS)/...
test-new:
	@if go test -race -coverprofile=coverage.out -covermode=atomic $(NEW_PKGS) > test.log 2>&1; then \
		rm -f test.log; \
	else cat test.log; rm -f test.log; exit 1; fi

# ── Legacy (advisory, not in PR gates) ─────────────────────────────────

be-ci-legacy:
	@echo "==> Running legacy checks..."
	gofumpt -l internal/app internal/domain internal/infra
	golangci-lint run ./internal/app/... ./internal/domain/... ./internal/infra/...

# ── Full CI (legacy + new, informational) ──────────────────────────────

be-ci: ci-mod check-format lint test  # blanket ./... as before

# ── Review gates (PR blocker, new-code only) ───────────────────────────

review-gates: build
	@echo "==> Compiling all code (legacy + new)..."
	go build ./...
	@echo "==> Running custom verification gates..."
	go run cmd/gates/main.go --all
	@echo "==> Scoped CI on new code..."
	$(MAKE) be-ci-new
```

### Step 2: FE-CI scoped to `frontend-next/`

The new frontend lives exclusively in `frontend-next/`. FE-CI gates only that directory:

```makefile
FE_NEXT_DIR := frontend-next

fe-ci:
	@if [ -f $(FE_NEXT_DIR)/package.json ]; then \
		echo "==> Running frontend-next CI..."; \
		cd $(FE_NEXT_DIR) && bun run lint && bun run format:check && tsc --noEmit && bun run test; \
	elif [ -f frontend/package.json ]; then \
		echo "==> [legacy] Running frontend CI..."; \
		cd frontend && bun run lint && bun run format:check && tsc --noEmit && bun run test; \
	else \
		echo "==> Skipping frontend CI: no frontend scaffolded yet."; \
	fi
```

### Step 3: `frontend-next/` directory structure

```
frontend-next/
  src/
    app/
    components/
      ui/
      features/
    lib/
    styles/
    __tests__/
```

Create with `.gitkeep` files in each empty directory.

### Step 4: Fix branch gate

Either:
- Update `GitLog()` in `internal/gates/git.go` to only check commits since base branch (`git log main..HEAD`)
- Or rebase offending commit before proceeding

### Step 5: Add test fixtures + gates tests

Create test file `internal/gates/gate_legacy_scope_test.go`:

```go
func TestLegacyCodeNotGatedByFormatCheck(t *testing.T) {
    // Given: a fixture with bad formatting in legacy dir (internal/app/.../bad.go)
    // Then: make review-gates should PASS (legacy is excluded)
}

func TestNewCodeGatedByFormatCheck(t *testing.T) {
    // Given: a fixture with bad formatting in new dir (internal/user/bad_test_stub.go)
    // Then: make review-gates should FAIL (new code is enforced)
}
```

### Step 6: Update docs

Update `conventions.md`, `architecture.md`, `DEVELOPMENT.md`, `general-instructions.md`, `sds.md`, `target-architecture-with-phases.md` to reflect new CI scoping:
- `be-ci-new` / `be-ci-legacy` / `be-ci` (full) targets
- `fe-ci` scoped to `frontend-next/`
- `review-gates` decoupled from `make ci`
- `NEW_DIRS` / `NEW_PKGS` variables documented

## Raw Reports

### Report: DeepSeek V4 Flash
Insight: Custom gates correct at Go level. Makefile layer bottlenecks. Split be-ci into legacy/new. Keep go build on all. Branch gate fix separate.

### Report: Mimo 2.5
Insight: Static tools (gofmt, staticcheck) don't support skip-dirs. Explicit path lists in Makefile required. `go build ./...` safe for legacy.

### Report: Big Pickle
Insight: Strangler Fig demands gate adapt to code state. `go build` on all. Everything else scoped to new dirs. Test fixtures prove both directions. Branch gate is noise.

### Report: Kimi K2.6
Insight: Two-phase gate: custom gates first (already correct), then scoped CI. `go build` all code. Test: legacy violations pass, new-code violations fail.

### Report: GLM 5.1
Insight: `NEW_PKGS` variable drives all scoped tools. Decouple `review-gates` from `ci`. Test proves both directions. Keep govulncheck on legacy as optional.
