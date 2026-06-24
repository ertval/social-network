/*
UnitTestGate runs Go unit tests (go test -race ./...) to ensure code correctness.
*/
package gates

// UnitTestGate runs the unit test suite.
type UnitTestGate struct{}

func (g *UnitTestGate) Name() string { return "go-test" }

func (g *UnitTestGate) Run() Result {
	// #nosec G204
	cmd := ExecCommand("go", "test", "-race", "./...")
	if _, err := cmd.CombinedOutput(); err != nil {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: "gate did not pass. Run 'go test -race ./...' to check details.",
		}
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: "all unit tests passed",
	}
}
