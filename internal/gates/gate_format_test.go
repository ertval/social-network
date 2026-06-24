package gates

import (
	"testing"
)

func TestFormatGate_Run(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand

	g := &FormatGate{}

	// 1. Tools available, PASS
	lookPath = func(name string) (string, error) { return name, nil }
	t.Setenv("MOCK_FAIL", "0")
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected format PASS, got: %s (%s)", res.Status, res.Message)
	}

	// 2. Tools available, FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected format FAIL, got: %s (%s)", res.Status, res.Message)
	}
}
