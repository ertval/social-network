package gates

import (
	"os"
	"path/filepath"
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
