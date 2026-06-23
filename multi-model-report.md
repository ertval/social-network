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

## Per-Model Summary

| Model | #Findings | Key Insight |
|-------|-----------|-------------|
| DeepSeek V4 Flash | 11 | Custom gates correct; Makefile is the bottleneck. `review-gates` should bypass `make ci`. |
| Mimo 2.5 | 7 | Static tools (gofmt, staticcheck) lack skip-dir flags — explicit path lists are necessary. |
| Big Pickle | 12 | Strangler Fig demands gate adaptation. `go build ./...` is safe baseline; everything else scoped. |
| Kimi K2.6 | 10 | Two-phase gate: Phase 1 = custom gates (already correct), Phase 2 = scoped CI on new dirs only. |
| GLM 5.1 | 10 | `NEW_PKGS` Makefile variable drives all scoped tools. Test fixtures prove both directions. |

## Recommended Plan

### Step 1: Fix `make review-gates` (Makefile change)
Replace `make ci` dependency with targeted steps:

```makefile
review-gates: build ## Run PR gates on new code only
	@echo "==> Compiling all code (legacy + new)..."
	go build ./...
	@echo "==> Running custom verification gates..."
	go run cmd/gates/main.go --all
	@echo "==> Checking new-code formatting..."
	@gofumpt -l $(NEW_DIRS) | xargs -r echo "Formatting issues in:" && \
	  goimports -l -local $(MODULE) $(NEW_DIRS) | xargs -r echo "Import issues in:" && \
	  test -z "$(shell gofumpt -l $(NEW_DIRS))" || (echo "Formatting failures"; exit 1)
	@echo "==> Running new-code linters..."
	staticcheck $(NEW_PKGS)
	golangci-lint run --timeout=5m ./$(NEW_DIRS)...
	go vet $(NEW_PKGS)
```

### Step 2: Define `NEW_DIRS` / `NEW_PKGS` in Makefile
```makefile
NEW_DIRS := internal/user internal/follow internal/topic internal/comment \
            internal/group internal/event internal/chat internal/notification \
            internal/oauth internal/core internal/platform internal/bootstrap \
            internal/config internal/gates cmd/gates cmd/server
NEW_PKGS := $(addprefix $(MODULE)/, $(NEW_DIRS))
```

### Step 3: Keep legacy as optional target
```makefile
be-ci-legacy: ## Check legacy code (advisory, not in PR gates)
	@echo "==> Running legacy checks..."
	gofumpt -l internal/app internal/domain internal/infra
	golangci-lint run ./internal/app/... ./internal/domain/... ./internal/infra/...
```

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
    // Assert: exactly
}

func TestNewCodeGatedByFormatCheck(t *testing.T) {
    // Given: a fixture with bad formatting in new dir (internal/user/bad_test_stub.go)
    // Then: make review-gates should FAIL (new code is enforced)
    // Assert: exactly
}
```

### Step 6: Update docs (general-instructions.md, DEVELOPMENT.md)
Update Q2 gate commands to reflect new scoping. Update R5 code review checklist.

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
