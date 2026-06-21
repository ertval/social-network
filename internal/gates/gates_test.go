/*
Package gates tests verify the correct behavior of all gate checks,
using mock directories, AST fixtures, and command-line execution overrides.
*/
package gates

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStackGate_Pass(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module social-network\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &StackGate{GoModPath: gomod}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestStackGate_WrongVersion(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module social-network\n\ngo 1.22\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &StackGate{GoModPath: gomod}
	result := g.Run()
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL, got %s: %s", result.Status, result.Message)
	}
}

func TestStackGate_WrongModule(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module wrong-name\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &StackGate{GoModPath: gomod}
	result := g.Run()
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL, got %s: %s", result.Status, result.Message)
	}
}

func TestLayoutGate_Empty(t *testing.T) {
	dir := t.TempDir()
	// No feature dirs → PASS (or SKIP)
	g := &LayoutGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS for empty dir, got %s: %s", result.Status, result.Message)
	}
}

func TestLayoutGate_MissingStructure(t *testing.T) {
	dir := t.TempDir()
	// Create a feature dir without required structure
	featureDir := filepath.Join(dir, "user")
	if err := os.MkdirAll(featureDir, 0o700); err != nil {
		t.Fatal(err)
	}

	g := &LayoutGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL for missing structure, got %s: %s", result.Status, result.Message)
	}
}

func TestLayoutGate_CompleteStructure(t *testing.T) {
	dir := t.TempDir()
	featureDir := filepath.Join(dir, "user")
	for _, sub := range []string{"commands", "queries", "transport", "store"} {
		if err := os.MkdirAll(filepath.Join(featureDir, sub), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(featureDir, "user.go"), []byte("package user\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &LayoutGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestLayoutGate_SkipsDirs(t *testing.T) {
	dir := t.TempDir()
	// core/ should be skipped even without structure
	if err := os.MkdirAll(filepath.Join(dir, "core"), 0o700); err != nil {
		t.Fatal(err)
	}

	g := &LayoutGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS (core skipped), got %s: %s", result.Status, result.Message)
	}
}

func TestMigrationsGate_NoDirPass(t *testing.T) {
	g := &MigrationsGate{MigrationDir: "/nonexistent"}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS for missing dir, got %s: %s", result.Status, result.Message)
	}
}

func TestMigrationsGate_MissingDown(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000001_init.up.sql"), []byte("CREATE TABLE t(id INT);"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &MigrationsGate{MigrationDir: dir}
	result := g.Run()
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL for missing down migration, got %s: %s", result.Status, result.Message)
	}
}

func TestMigrationsGate_ValidPair(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000001_init.up.sql"), []byte("CREATE TABLE t(id INT);"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000001_init.down.sql"), []byte("DROP TABLE t;"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &MigrationsGate{MigrationDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestRunnerRunAll(t *testing.T) {
	runner := NewRunner()
	runner.Register(&StackGate{GoModPath: createTempGoMod(t, "module social-network\n\ngo 1.24\n")})

	report := runner.RunAll()
	if report.Overall != "PASS" {
		t.Errorf("expected PASS overall, got %s", report.Overall)
	}
	if len(report.Gates) != 1 {
		t.Errorf("expected 1 gate result, got %d", len(report.Gates))
	}
}

func TestRunnerRunOne(t *testing.T) {
	runner := NewRunner()
	runner.Register(&StackGate{GoModPath: createTempGoMod(t, "module social-network\n\ngo 1.24\n")})

	result, err := runner.RunOne("stack")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s", result.Status)
	}

	_, err = runner.RunOne("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent gate")
	}
}

func createTempGoMod(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func mockExecCommand(command string, args ...string) *exec.Cmd {
	var script string

	switch command {
	case "golangci-lint":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "echo 'depguard violation: transport importing store'; exit 1"
		} else {
			script = "echo 'no violations'; exit 0"
		}
	case "go-arch-lint":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "echo 'cycle detected'; exit 1"
		} else {
			script = "echo 'clean'; exit 0"
		}
	case "gosec":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "echo 'gosec violation: hardcoded password'; exit 1"
		} else {
			script = "echo 'clean'; exit 0"
		}
	case "go":
		if len(args) > 0 {
			switch args[0] {
			case "list":
				if os.Getenv("MOCK_FAIL") == "1" {
					script = "exit 1"
				} else if os.Getenv("MOCK_CYCLE") == "1" {
					pkg := args[len(args)-1]
					if strings.Contains(pkg, "/user/...") {
						script = `echo '{"Imports": ["social-network/internal/follow"]}'`
					} else if strings.Contains(pkg, "/follow/...") {
						script = `echo '{"Imports": ["social-network/internal/user"]}'`
					}
				} else if os.Getenv("MOCK_NOTIF") == "1" {
					pkg := args[len(args)-1]
					if strings.Contains(pkg, "/user/...") {
						script = `echo '{"Imports": ["social-network/internal/notification"]}'`
					}
				} else {
					script = `echo '{"Imports": []}'`
				}
			case "test":
				if os.Getenv("MOCK_FAIL") == "1" {
					script = "exit 1"
				} else {
					// Locate coverprofile arg and touch it
					covPath := ""
					for _, arg := range args {
						if strings.HasPrefix(arg, "-coverprofile=") {
							covPath = strings.TrimPrefix(arg, "-coverprofile=")
						}
					}
					if covPath != "" {
						script = fmt.Sprintf("touch %s; exit 0", covPath)
					} else {
						script = "exit 0"
					}
				}
			case "tool":
				if len(args) > 1 && args[1] == "cover" {
					if os.Getenv("MOCK_FAIL") == "1" {
						script = "exit 1"
					} else if os.Getenv("MOCK_COVER_MALFORMED") == "1" {
						script = "echo 'total'"
					} else {
						script = "echo 'total: (statements) 92.5%'"
					}
				}
			}
		}
	case "git":
		if os.Getenv("MOCK_FAIL") == "1" {
			script = "exit 1"
		} else if len(args) > 0 {
			switch args[0] {
			case "merge-base":
				script = "echo 'abcdef123456'"
			case "log":
				if os.Getenv("MOCK_GIT_EMPTY") == "1" {
					script = "echo ''"
				} else if os.Getenv("MOCK_COMMIT_FAIL") == "1" {
					script = "echo 'Fixing code'"
				} else {
					script = "echo 'feat(user): add auth handler'"
				}
			case "rev-parse":
				if os.Getenv("MOCK_REV_MAIN") == "1" {
					script = "echo 'main'"
				} else if os.Getenv("MOCK_REV_FAIL") == "1" {
					script = "exit 1"
				} else {
					script = "echo 'user/S1-BE-01-add-auth'"
				}
			case "diff":
				if os.Getenv("MOCK_GIT_EMPTY") == "1" {
					script = "echo ''"
				} else {
					script = "echo 'internal/user/user.go'"
				}
			case "worktree":
				if len(args) > 3 && args[1] == "add" {
					script = fmt.Sprintf("mkdir -p %s; exit 0", args[3])
				} else {
					script = "exit 0"
				}
			}
		}
	default:
		script = "exit 0"
	}

	return exec.Command("sh", "-c", script)
}

func TestBranchGate_Run(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	t.Run("on main branch", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "1")
		g := &BranchGate{}
		res := g.Run()
		if res.Status != "PASS" || res.Message != "on main or detached HEAD" {
			t.Errorf("expected PASS for main branch, got: %s (%s)", res.Status, res.Message)
		}
	})

	t.Run("conventional branch and commit", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "0")
		t.Setenv("MOCK_COMMIT_FAIL", "0")
		g := &BranchGate{}
		res := g.Run()
		if res.Status != "PASS" {
			t.Errorf("expected branch check PASS, got: %s (%s)", res.Status, res.Message)
		}
	})

	t.Run("non-conventional commit", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "0")
		t.Setenv("MOCK_COMMIT_FAIL", "1")
		g := &BranchGate{}
		res := g.Run()
		if res.Status != "FAIL" {
			t.Errorf("expected branch check FAIL for bad commit, got: %s", res.Status)
		}
	})
}

func TestScopeDriftGate_Run(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	t.Run("on main", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "1")
		g := &ScopeDriftGate{}
		res := g.Run()
		if res.Status != "PASS" {
			t.Errorf("expected scope-drift PASS on main, got: %s", res.Status)
		}
	})

	t.Run("on branch", func(t *testing.T) {
		t.Setenv("MOCK_REV_MAIN", "0")
		g := &ScopeDriftGate{}
		res := g.Run()
		if res.Status != "PASS" || !strings.Contains(res.Message, "changed") {
			t.Errorf("expected scope-drift PASS on branch with log, got: %s (%s)", res.Status, res.Message)
		}
	})
}

func TestWriteJSON(t *testing.T) {
	report := Report{
		Overall: "PASS",
		Gates: []Result{
			{Gate: "stack", Status: "PASS", Message: "OK"},
		},
	}
	var buf bytes.Buffer
	err := WriteJSON(&buf, report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"overall": "PASS"`) {
		t.Errorf("unexpected JSON: %s", buf.String())
	}
}

func TestGitHelpers(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	// FindBaseBranch
	base := FindBaseBranch()
	if base != "main" {
		t.Errorf("expected base main, got: %s", base)
	}

	// GitLog
	log, err := GitLog("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(log) != 1 || log[0] != "feat(user): add auth handler" {
		t.Errorf("unexpected git log: %v", log)
	}

	// GitBranch
	branch := GitBranch()
	if branch != "user/S1-BE-01-add-auth" {
		t.Errorf("expected user/S1-BE-01-add-auth, got: %s", branch)
	}

	// GitDiffFiles
	files, err := GitDiffFiles("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "internal/user/user.go" {
		t.Errorf("unexpected files: %v", files)
	}
}

func TestGitHelpers_Empty(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	t.Setenv("MOCK_GIT_EMPTY", "1")

	log, err := GitLog("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log != nil {
		t.Errorf("expected nil log for empty output, got: %v", log)
	}

	files, err := GitDiffFiles("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if files != nil {
		t.Errorf("expected nil files for empty output, got: %v", files)
	}
}

func TestLayoutGate_Errors(t *testing.T) {
	g := &LayoutGate{InternalDir: "/nonexistent"}
	res := g.Run()
	if res.Status != "SKIP" {
		t.Errorf("expected SKIP for nonexistent dir, got: %s", res.Status)
	}
}

func TestMigrationsGate_ReadDirError(t *testing.T) {
	// Pass a file path instead of a directory to force os.ReadDir failure
	dir := t.TempDir()
	filePath := filepath.Join(dir, "migrations.txt")
	if err := os.WriteFile(filePath, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	g := &MigrationsGate{MigrationDir: filePath}
	res := g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL for file path as migration dir, got: %s (%s)", res.Status, res.Message)
	}
}

func TestScopeDriftGate_Error(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	g := &ScopeDriftGate{}
	t.Setenv("MOCK_FAIL", "1")
	res := g.Run()
	if res.Status != "PASS" || !strings.Contains(res.Message, "no changes") {
		t.Errorf("expected PASS with no changes message on error, got: %s (%s)", res.Status, res.Message)
	}
}
