# Multi-Model Analysis Report

**Plan:** Integrated Gates Runner and Clean CLI Output  
**Date:** 2026-06-24  
**Models:** DeepSeek V4 Flash, Mimo 2.5, Big Pickle, GLM 5.1 (Kimi K2.6: no output)

---

## Consensus Findings (3+ models agreed)

### [BUG] Main README.md — Missing 4 gates from individual gate list
- DeepSeek, Mimo, Big Pickle
- Lines 277–286 only list 10 gates (`stack`, `layout`, `boundaries`, `dag`, `tdd`, `migrations`, `security`, `branch`, `coverage`, `scopedrift`).
- **Missing:** `format`, `lint`, `go-test`, `frontend`.
- Fix: Add `go run cmd/gates/main.go --gate=format`, `--gate=lint`, `--gate=go-test`, `--gate=frontend`.

### [BUG] Main README.md — Line 289 says "Output is JSON"
- DeepSeek, Mimo, Big Pickle
- Default output is human-readable text (`[PASS]/[FAIL]/[SKIP]`). JSON requires `--json` flag.
- Fix: Change to "Default output is human-readable text; use `--json` for JSON format."

### [BUG] Main README.md — Phantom steps 3-4 in review-gates description
- Mimo, Big Pickle
- Lines 258–261 list steps 3 (`make be-ci-new`) and 4 (`make fe-ci`) as part of `make review-gates`.
- Actual Makefile target only runs 2 steps: `go build ./...` + `go run cmd/gates/main.go --all`.
- Fix: Remove lines 258–261 or document as separate optional steps.

### [RISK] SecurityGate AST checks scan entire `internal/` dir
- Mimo, Big Pickle, GLM 5.1
- `runASTChecks()` in `gate_security.go:83` reads `internal/` and walks all feature slices.
- Not filtered by `NewDirs` — would check legacy code if present.
- Fix: Scope `InternalDir` to `NewDirs` or skip legacy feature slices.

### [RISK] CoverageGate uses `./...` for all packages
- Big Pickle, GLM 5.1
- `getCurrentCoverage()` in `gate_coverage.go` runs `go test -coverprofile=... ./...` — tests ALL packages, not just `NewPkgs`.
- Fix: Change `./...` to `NewPkgs`.

---

## Unique Findings (1-2 models)

### [BUG] Main README.md — Gate names mismatch actual `Gate.Name()` values
- Big Pickle
- README uses `--gate=layout`, `--gate=boundaries`, `--gate=dag`, `--gate=coverage`.
- Actual names: `d1-layout`, `d5-boundaries`, `d6-dag`, `coverage-delta`.
- Fix: Update README gate names to match `Gate.Name()`.

### [RISK] FormatGate silent-skip when tools missing
- DeepSeek V4 Flash
- `gofumpt` and `goimports` not installed → `toolAvailable` returns false → no check runs → gate returns PASS silently.
- User gets false positive on formatting.
- Fix: Either skip with clear message or enforce check with `go fmt` fallback.

### [RISK] CoverageGate worktree cleanup unsafe
- GLM 5.1
- `os.RemoveAll` + `git worktree add` — if process killed mid-run, stale worktree remains orphaned.

### [INSIGHT] 9 of 14 gates inherently operate at project level
- Big Pickle
- Stack, Layout, Boundaries, DAG, TDD, Migrations, Branch, ScopeDrift, Coverage are structural/git gates.
- Only the 5 new gates (Format, Lint, UnitTest, Frontend, Security gosec/govulncheck) use `NewDirs`/`NewPkgs`.
- This is likely intentional (architectural gates need full project view).

### [INSIGHT] All 14 gates exist, registered, and compile
- All models
- Source files present, registered in `cmd/gates/main.go`, tests pass (67 tests).

### [INFO] Output format matches spec
- All models
- `[PASS] name`, `[FAIL] name: message`, `[SKIP] name: reason` with exit code 1 on FAIL.

### [INFO] All 4 new test files present with meaningful tests
- All models
- `gate_format_test.go`, `gate_lint_test.go`, `gate_unittest_test.go`, `gate_frontend_test.go`.
- Tests cover PASS/FAIL/SKIP scenarios, mock via `helper_test.go`.
- `helper_test.go` mocks: gofumpt, goimports, govulncheck, bun, golangci-lint, gosec, staticcheck, go (list/test/vet), git.

### [INFO] Makefile `review-gates` matches plan exactly
- All models
- `go build ./...` + `go run cmd/gates/main.go --all`.

---

## Contradictions

- **CoverageGate scoping:** Plan says "all gates run only on new codebase". Models agree CoverageGate uses `./...`. But Big Pickle notes comparison inherently needs full project scope — may be intentional.
- **SecurityGate AST scoping:** Plan says "new codebase only". AST checks scan all of `internal/`. But in greenfield, NewDirs ≈ all of `internal/`. Disagreement: bug vs acceptable in practice.

---

## Per-Model Summary

| Model | #Findings | Key Insight |
|-------|-----------|-------------|
| DeepSeek V4 Flash | 5 | README drift (missing gates, JSON claim). FormatGate silent skip. |
| Mimo 2.5 | 6 | README phantom steps (lines 258-261). SecurityGate AST scoping. |
| Big Pickle | 9 | README gate name mismatches. 9/14 gates not scoped to new code. |
| GLM 5.1 | 6 | CoverageGate `./...` bug. SecurityGate AST scoping. CoverageGate crash safety. |
| Kimi K2.6 | — | No output returned. |

---

## Raw Reports

### DeepSeek V4 Flash

#### Findings
- [BUG] README.md individual gate command list missing `format`, `lint`, `go-test`, `frontend` gates (only 10 of 14 listed)
- [BUG] README.md line 289 says "Output is JSON with exit codes" — default is human-readable text, JSON requires `--json` flag
- [RISK] FormatGate passes silently when gofumpt/goimports not installed (no check runs, no warning)
- [INSIGHT] All 14 gates exist, registered in `cmd/gates/main.go`, pass `make review-gates`
- [INSIGHT] FormatGate/LintGate/UnitTestGate/FrontendGate/SecurityGate all correctly scope to NewDirs/NewPkgs only
- [INSIGHT] SecurityGate govulncheck error message suggests `govulncheck ./...` but gate only runs on NewPkgs (minor inconsistency)
- [INFO] Output format matches spec: `[PASS] name`, `[FAIL] name: msg`, `[SKIP] name: reason` with exit code 1 on FAIL
- [INFO] All 4 test files exist (gate_format, gate_lint, gate_unittest, gate_frontend test) with meaningful PASS/FAIL/SKIP tests — all pass
- [INFO] `internal/gates/README.md` has complete 14-gate catalog table with correct output format docs
- [INFO] `make review-gates` works correctly: `go build ./...` then `go run cmd/gates/main.go --all`

#### Reasoning
- Main README line 275-286 lists only original 10 gates; format/lint/go-test/frontend omitted despite being registered and functional.
- Line 289 says "Output is JSON" which contradicts default text output; JSON is opt-in via `--json`.
- FormatGate returns PASS with "code formatting OK" when gofumpt/goimports missing — user gets false positive on formatting quality.
- Gates targeting new code correctly use `NewDirs`/`NewPkgs` vars defined in `features.go`. 
- All source code compiles and 67 tests pass. `make review-gates` produces correct output for all 14 gates.
- SecurityGate `gate_security.go` line 41: `"Run 'govulncheck ./...' to check details."` should say `NewPkgs` not `./...`.
- `internal/gates/README.md` is the authoritative gate docs and is fully correct.

#### Uncertainties
- Not sure if omitting 4 gates from main README command list is intentional (plan says update but unclear which README).
- Not sure if FormatGate silent-skip on missing tools is by design (defense-in-depth vs strict enforcement).

---

### Mimo 2.5

#### Findings
- [INSIGHT] All 14 gates correctly registered in main.go — StackGate through FrontendGate, exact match
- [INSIGHT] Default output format correct — [PASS]/[FAIL] with message, [SKIP] with reason (lines 81-93)
- [INSIGHT] Format/Lint/UnitTest/Frontend/Security gates all scope to NewDirs/NewPkgs correctly
- [INSIGHT] All test files present (format, lint, unittest, frontend) with meaningful PASS/FAIL/SKIP scenarios
- [INSIGHT] Tests pass: 67 passed in 1 packages
- [BUG] README.md (lines 255-261) describes 4 steps for `review-gates` (build + gates + be-ci-new + fe-ci) but actual Makefile only runs 2 steps (build + gates). Steps 3-4 are phantom documentation.
- [BUG] README.md line 289 says "Output is JSON" but default is text. JSON requires `--json` flag. Misleading for users.
- [BUG] README.md gate catalog (lines 277-286) lists only 10 gates, missing: format, lint, go-test, frontend. 4 gates undocumented.
- [RISK] SecurityGate AST checks (runASTChecks) scan entire `internal/` directory tree, not scoped to NewDirs. If legacy code existed under internal/, it would be checked. Greenfield codebase now, but gate is fragile to future legacy dirs.
- [SUGGESTION] gate_security.go: AST checks should filter by NewDirs or accept dir parameter for consistency with gosec/govulncheck scoping.

#### Reasoning
- All 14 gates from plan exist in code with correct names and correct scoping to new code only (NewDirs/NewPkgs).
- Output format matches spec: [PASS] name, [FAIL] name: message, [SKIP] name: reason.
- All 4 test files present with correct scenarios. helper_test.go mocks all required tools.
- Makefile `review-gates` matches plan exactly: `go build ./...` then `go run cmd/gates/main.go --all`.
- README documentation has 3 distinct drift issues: phantom steps, wrong default format description, incomplete gate catalog.
- SecurityGate AST section (lines 83-146) reads raw `internal/` dir, not filtered by NewDirs. All other gate tools (gosec, govulncheck, gofumpt, goimports, golangci-lint, go test) correctly scope to new code.

#### Uncertainties
- Whether conventions.md commit scopes table also needs updating (another model flagged this drift).
- Whether general-instructions.md references review-gates with the phantom 4-step description.

---

### Big Pickle

#### Findings
- [PASS] All 14 gate types exist as files, registered in main.go, with matching test files
- [PASS] FormatGate uses NewDirs for gofumpt/goimports checks
- [PASS] LintGate uses NewDirs for golangci-lint, NewPkgs for fallback
- [PASS] UnitTestGate uses NewPkgs for go test -race
- [PASS] FrontendGate checks frontend-next/ then frontend/, runs bun steps
- [PASS] SecurityGate uses NewDirs for gosec, NewPkgs for govulncheck
- [PASS] Output format matches spec: [PASS]/[FAIL]+msg/[SKIP]+reason, --json flag
- [PASS] make review-gates matches plan: go build ./... && go run cmd/gates/main.go --all
- [PASS] Test coverage comprehensive: PASS/FAIL/SKIP scenarios, AST edge cases, git helpers
- [BUG] Main README.md lists 4 incorrect gate names: layout→d1-layout, boundaries→d5-boundaries, dag→d6-dag, coverage→coverage-delta
- [BUG] Main README.md missing 4 gates from examples: format, lint, go-test, frontend
- [BUG] Main README.md line 289 says "Output is JSON" but default is human-readable text
- [RISK] 9 of 14 gates not scoped to NewDirs/NewPkgs: Stack, Layout, Boundaries, DAG, TDD, Migrations, Branch, ScopeDrift, Coverage operate at project level
- [RISK] CoverageGate runs go test `./...` on ALL packages (legacy+new), not NewPkgs
- [RISK] SecurityGate AST checks scan entire internal/ dir, not NewDirs
- [RISK] BoundariesGate golangci-lint depguard path scans whole project
- [RISK] DAGGate go-arch-lint path scans whole project, not new code
- [INSIGHT] internal/gates/README.md is correct and comprehensive — only main README.md has drifts
- [INSIGHT] `make review-gates` compiles legacy+new code before running new-code gates, creating a mixed-scope pipeline
- [INSIGHT] Gate Name() values differ from struct types (e.g., UnitTestGate→"go-test", CoverageGate→"coverage-delta") – consistent internally but affects --gate= CLI usage
- [SUGGESTION] Add format, lint, go-test, frontend gate examples to main README.md `--gate=` list
- [SUGGESTION] Fix main README.md gate names: layout→d1-layout, boundaries→d5-boundaries, dag→d6-dag, coverage→coverage-delta
- [SUGGESTION] Fix main README.md line 289: clarify default is text, --json for JSON
- [SUGGESTION] Consider scoping CoverageGate and BoundariesGate fallback to NewPkgs/NewDirs for consistency

#### Reasoning
- Verified all 14 gate source files exist, all registered in main.go, all have _test.go
- Checked each gate's Run() for NewDirs/NewPkgs usage: only 5 use them; remaining 9 are structural/git gates operating at project root
- Compared main README.md gate examples against actual Gate.Name() values: 4 mismatches, 4 missing entries
- Reviewed Makefile review-gates target: matches plan exactly (go build ./... + go run cmd/gates/main.go --all)
- CoverageGate's getCurrentCoverage uses `go test ./...` (all packages), not NewPkgs
- SecurityGate runASTChecks defaults to "internal" dir, walks entire dir tree — broader than NewDirs
- BoundariesGate Run() with golangci-lint runs on project root, not NewDirs
- DAGGate Run() with go-arch-lint runs on project root, not NewDirs
- internal/gates/README.md is fully correct (14 gates, correct names, correct output format docs)
- Tests cover PASS/FAIL paths for all gates, AST edge cases for security, git helper edge cases
- No runtime bugs found in gate logic; helper_test.go mocks all tool binaries properly

#### Uncertainties
- Not sure if 9 non-scoped gates is intentional design (architectural gates inherently need full project view) or should be constrained to NewDirs/NewPkgs
- README.md might describe overall PR pipeline (steps 1-4) rather than just `make review-gates` — ambiguous wording
- SecurityGate AST checks scanning entire internal/ may be acceptable since NewDirs covers most of internal/, but technically broader

---

### GLM 5.1

#### Findings
- [BUG] CoverageGate uses `./...` — runs on ALL code (legacy+new), not scoped to NewDirs/NewPkgs
- [BUG] SecurityGate AST checks scan whole `internal/` dir — not scoped to NewDirs at all
- [BUG] UnitTestGate comment/docs say `./...` but implementation uses NewPkgs (good), however comment on line 2 is misleading
- [RISK] FrontendGate searches for `frontend-next/package.json` and `frontend/package.json` via `os.Stat` relative to CWD — breaks if run from non-project-root
- [RISK] CoverageGate `getBaselineCoverage` does `os.RemoveAll` + `git worktree add` — if crashed mid-run, stale worktree remains; defer only removes on success path
- [RISK] CoverageGate `getCurrentCoverage` uses `./...` (line 96) — tests ALL packages, not NewPkgs
- [SUGGESTION] SecurityGate `runASTChecks` iterates `os.ReadDir(dir)` then walks feature slices — should scope to NewDirs instead of scanning entire `internal/` tree
- [SUGGESTION] `gate_coverage.go` lines 82 and 96 both use `./...` — should use NewPkgs for consistency with other gates' new-code-only scoping
- [INSIGHT] All 14 gates registered correctly in `cmd/gates/main.go`, matching plan exactly
- [INSIGHT] Output format matches spec: `[PASS] gate`, `[FAIL] gate: message`, `[SKIP] gate: reason`
- [INFO] All 4 test files present with meaningful tests: format (PASS/FAIL), lint (golangci + fallback), unittest (PASS/FAIL), frontend (SKIP/PASS/FAIL)
- [INFO] Mock coverage in helper_test.go: gofumpt, goimports, govulncheck, bun, golangci-lint, gosec, staticcheck, go, git — matches plan
- [INFO] `make review-gates` matches spec: `go build ./...` then `go run cmd/gates/main.go --all`
- [INFO] `--json` flag works; default human-readable output with [PASS]/[FAIL]/[SKIP]

#### Reasoning
- CoverageGate line 82 (`getBaselineCoverage`) and line 96 (`getCurrentCoverage`) both use `go test -coverprofile=... ./...` — this runs ALL packages, not just NewPkgs. The plan requires gates scope to NEW code only. FormatGate, LintGate, UnitTestGate, SecurityGate (gosec/govulncheck) all correctly use NewDirs/NewPkgs, but CoverageGate doesn't.
- SecurityGate's `runASTChecks()` reads `internal/` directory and walks all feature slices — not filtered by NewDirs. While gosec/govulncheck are scoped, the AST checks are not. This is a scoping deviation.
- UnitTestGate line 2 comment says `go test -race ./...` but implementation on line 12 uses NewPkgs — the code is correct, the comment is misleading.
- FrontendGate uses relative paths `frontend-next/package.json` — fragile if CWD changes. Other gates don't have this pattern but it's a realistic risk.
- CoverageGate worktree cleanup: `os.RemoveAll` on tempDir before add, then defer removes after. If process killed between `worktree add` and completion, worktree remains orphaned. Minor risk.

#### Uncertainties
- Whether CoverageGate's `./...` usage is intentional (to compare total coverage baseline vs branch) vs a bug — the plan says "new code only" but coverage comparison inherently needs full project scope
- Whether SecurityGate AST checks intentionally scan all of `internal/` (since they check security patterns that should be enforced everywhere, not just new code)
- Whether FrontendGate's CWD-relative path is acceptable or if it should resolve project root explicitly
