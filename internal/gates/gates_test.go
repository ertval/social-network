package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStackGate_Pass(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module social-network\n\ngo 1.24\n"), 0644); err != nil {
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
	if err := os.WriteFile(gomod, []byte("module social-network\n\ngo 1.22\n"), 0644); err != nil {
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
	if err := os.WriteFile(gomod, []byte("module wrong-name\n\ngo 1.24\n"), 0644); err != nil {
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
	if err := os.MkdirAll(featureDir, 0755); err != nil {
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
		if err := os.MkdirAll(filepath.Join(featureDir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(featureDir, "user.go"), []byte("package user\n"), 0644); err != nil {
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
	if err := os.MkdirAll(filepath.Join(dir, "core"), 0755); err != nil {
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
	if err := os.WriteFile(filepath.Join(dir, "000001_init.up.sql"), []byte("CREATE TABLE t(id INT);"), 0644); err != nil {
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
	if err := os.WriteFile(filepath.Join(dir, "000001_init.up.sql"), []byte("CREATE TABLE t(id INT);"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000001_init.down.sql"), []byte("DROP TABLE t;"), 0644); err != nil {
		t.Fatal(err)
	}

	g := &MigrationsGate{MigrationDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestTDDGate_NoFeatures(t *testing.T) {
	dir := t.TempDir()
	g := &TDDGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestTDDGate_MissingTests(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "user", "commands")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "create.go"), []byte("package commands\n"), 0644); err != nil {
		t.Fatal(err)
	}

	g := &TDDGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL for missing test files, got %s: %s", result.Status, result.Message)
	}
}

func TestTDDGate_WithTests(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "user", "commands")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "create.go"), []byte("package commands\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "create_test.go"), []byte("package commands\n"), 0644); err != nil {
		t.Fatal(err)
	}

	g := &TDDGate{InternalDir: dir}
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
