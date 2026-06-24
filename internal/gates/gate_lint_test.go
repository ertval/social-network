package gates

import (
	"os"
	"strings"
	"testing"
)

func TestLintGate_Run(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand

	g := &LintGate{MaxLines: 10000, Dirs: []string{"."}}

	// 1. Tool available (golangci-lint), PASS
	lookPath = func(name string) (string, error) {
		if name == "golangci-lint" {
			return name, nil
		}
		return "", os.ErrNotExist
	}
	t.Setenv("MOCK_FAIL", "0")
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected lint PASS (golangci-lint), got: %s (%s)", res.Status, res.Message)
	}

	// 2. Tool available (golangci-lint), FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected lint FAIL (golangci-lint), got: %s (%s)", res.Status, res.Message)
	}

	// 3. Fallbacks (staticcheck + go vet), PASS
	lookPath = func(name string) (string, error) {
		if name == "staticcheck" || name == "go" {
			return name, nil
		}
		return "", os.ErrNotExist
	}
	t.Setenv("MOCK_FAIL", "0")
	res = g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected lint PASS (fallbacks), got: %s (%s)", res.Status, res.Message)
	}

	// 4. Fallbacks, FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected lint FAIL (fallbacks), got: %s (%s)", res.Status, res.Message)
	}
}

func TestLintGate_LineLimit(t *testing.T) {
	// Use a small limit so that stack gate (59 lines) triggers failure
	g := &LintGate{MaxLines: 10, Dirs: []string{"."}}

	// Mock lookPath to do nothing
	oldLook := lookPath
	defer func() { lookPath = oldLook }()
	lookPath = func(name string) (string, error) {
		return "", os.ErrNotExist
	}

	res := g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL due to line limit, got: %s (%s)", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "exceeding maximum line limit") {
		t.Errorf("expected error message to mention line limit, got: %s", res.Message)
	}
}
