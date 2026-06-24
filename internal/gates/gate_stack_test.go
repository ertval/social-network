package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createMockStack(t *testing.T, goModContent, envContent string, createPlatform bool, frontendSubdir string, pkgJSONContent string, createBunLock bool) string {
	t.Helper()
	dir := t.TempDir()

	if goModContent != "" {
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if envContent != "" {
		if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if createPlatform {
		if err := os.MkdirAll(filepath.Join(dir, "internal", "platform"), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	if frontendSubdir == "" {
		return dir
	}

	fePath := filepath.Join(dir, frontendSubdir)
	if err := os.MkdirAll(fePath, 0o700); err != nil {
		t.Fatal(err)
	}
	if pkgJSONContent != "" {
		if err := os.WriteFile(filepath.Join(fePath, "package.json"), []byte(pkgJSONContent), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if createBunLock {
		if err := os.WriteFile(filepath.Join(fePath, "bun.lockb"), []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestStackGate_Pass_NoFrontend(t *testing.T) {
	goMod := "module social-network\ngo 1.25\n"
	env := "DB_DRIVER=sqlite3\nDB_PRAGMA=_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000\n"
	dir := createMockStack(t, goMod, env, true, "", "", false)

	g := &StackGate{
		GoModPath: filepath.Join(dir, "go.mod"),
		EnvPath:   filepath.Join(dir, ".env"),
		RootDir:   dir,
	}

	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", res.Status, res.Message)
	}

	if !strings.Contains(res.Message, "status: OK") || !strings.Contains(res.Message, "frontend is not scaffolded yet") {
		t.Errorf("expected descriptive PASS message, got: %s", res.Message)
	}
}

func TestStackGate_Pass_WithValidFrontend(t *testing.T) {
	goMod := "module social-network\ngo 1.25\n"
	env := "DB_DRIVER=sqlite3\nDB_PRAGMA=_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000\n"
	pkg := `{"dependencies": {"next": "^14.0.0", "tailwindcss": "^3.0.0"}}`
	dir := createMockStack(t, goMod, env, true, "frontend-next", pkg, true)

	g := &StackGate{
		GoModPath:   filepath.Join(dir, "go.mod"),
		EnvPath:     filepath.Join(dir, ".env"),
		FrontendDir: "frontend-next",
		RootDir:     dir,
	}

	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", res.Status, res.Message)
	}

	if !strings.Contains(res.Message, "Next.js/Tailwind/Bun verified") {
		t.Errorf("expected Next.js/Tailwind/Bun verified in output, got: %s", res.Message)
	}
}

func TestStackGate_Fail_WrongGoVersion(t *testing.T) {
	goMod := "module social-network\ngo 1.22\n"
	env := "DB_DRIVER=sqlite3\nDB_PRAGMA=_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000\n"
	dir := createMockStack(t, goMod, env, true, "", "", false)

	g := &StackGate{
		GoModPath: filepath.Join(dir, "go.mod"),
		EnvPath:   filepath.Join(dir, ".env"),
		RootDir:   dir,
	}

	res := g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL, got %s: %s", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "expected Go 1.25 in go.mod") {
		t.Errorf("expected Go 1.25 error in message, got: %s", res.Message)
	}
}

func TestStackGate_Fail_MissingWAL(t *testing.T) {
	goMod := "module social-network\ngo 1.25\n"
	env := "DB_DRIVER=sqlite3\nDB_PRAGMA=_foreign_keys=on&_busy_timeout=5000\n"
	dir := createMockStack(t, goMod, env, true, "", "", false)

	g := &StackGate{
		GoModPath: filepath.Join(dir, "go.mod"),
		EnvPath:   filepath.Join(dir, ".env"),
		RootDir:   dir,
	}

	res := g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL, got %s: %s", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "SQLite journal mode is not set to WAL") {
		t.Errorf("expected WAL error, got: %s", res.Message)
	}
}

func TestStackGate_Fail_MissingBusyTimeout(t *testing.T) {
	goMod := "module social-network\ngo 1.25\n"
	env := "DB_DRIVER=sqlite3\nDB_PRAGMA=_foreign_keys=on&_journal_mode=WAL\n"
	dir := createMockStack(t, goMod, env, true, "", "", false)

	g := &StackGate{
		GoModPath: filepath.Join(dir, "go.mod"),
		EnvPath:   filepath.Join(dir, ".env"),
		RootDir:   dir,
	}

	res := g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL, got %s: %s", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "SQLite busy timeout is not configured") {
		t.Errorf("expected busy timeout error, got: %s", res.Message)
	}
}

func TestStackGate_Fail_MissingPlatform(t *testing.T) {
	goMod := "module social-network\ngo 1.25\n"
	env := "DB_DRIVER=sqlite3\nDB_PRAGMA=_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000\n"
	dir := createMockStack(t, goMod, env, false, "", "", false)

	g := &StackGate{
		GoModPath: filepath.Join(dir, "go.mod"),
		EnvPath:   filepath.Join(dir, ".env"),
		RootDir:   dir,
	}

	res := g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL, got %s: %s", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "missing internal/platform directory") {
		t.Errorf("expected platform dir error, got: %s", res.Message)
	}
}

func TestStackGate_Fail_InvalidFrontend(t *testing.T) {
	goMod := "module social-network\ngo 1.25\n"
	env := "DB_DRIVER=sqlite3\nDB_PRAGMA=_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000\n"
	// Missing tailwind, missing bun lockfile
	pkg := `{"dependencies": {"next": "^14.0.0"}}`
	dir := createMockStack(t, goMod, env, true, "frontend", pkg, false)

	g := &StackGate{
		GoModPath:   filepath.Join(dir, "go.mod"),
		EnvPath:     filepath.Join(dir, ".env"),
		FrontendDir: "frontend",
		RootDir:     dir,
	}

	res := g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL, got %s: %s", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "Tailwind CSS dependency missing") || !strings.Contains(res.Message, "Bun lockfile") {
		t.Errorf("expected missing tailwind/lockfile errors, got: %s", res.Message)
	}
}
