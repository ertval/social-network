package gates

import (
	"testing"
)

func TestCoverageGate_Run(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	g := &CoverageGate{Threshold: 5}

	// 1. Delta within limit -> PASS
	t.Setenv("MOCK_FAIL", "0")
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected coverage PASS, got: %s (%s)", res.Status, res.Message)
	}

	// 2. Command fail -> PASS (warns but does not fail pipeline)
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected coverage PASS with error message on run failure, got: %s (%s)", res.Status, res.Message)
	}
}

func TestCoverageGate_Errors(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	g := &CoverageGate{}

	t.Run("getCurrentCoverage command fail", func(t *testing.T) {
		t.Setenv("MOCK_FAIL", "1")
		_, err := getCurrentCoverage()
		if err == nil {
			t.Error("expected error for go test command failure, got none")
		}
	})

	t.Run("parseCoverageFile malformed cover", func(t *testing.T) {
		t.Setenv("MOCK_COVER_MALFORMED", "1")
		res := g.Run()
		if res.Status != "PASS" { // advisory PASS on error
			t.Errorf("expected PASS on parsing error, got: %s (%s)", res.Status, res.Message)
		}
	})
}
