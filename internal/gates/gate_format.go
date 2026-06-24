/*
FormatGate validates Go code formatting (gofumpt and goimports).
If formatting violations are found, it directs the user to run 'make format'.
*/
package gates

import (
	"strings"
)

// FormatGate checks Go code formatting.
type FormatGate struct{}

func (g *FormatGate) Name() string { return "format" }

func (g *FormatGate) Run() Result {
	var errors []string

	// Check gofumpt formatting
	if toolAvailable("gofumpt") {
		// Run gofumpt -l cmd internal
		// #nosec G204
		cmd := ExecCommand("gofumpt", "-l", "cmd", "internal")
		out, err := cmd.CombinedOutput()
		if err != nil || len(strings.TrimSpace(string(out))) > 0 {
			errors = append(errors, "Run 'gofumpt -l -w cmd internal' or 'make format' to check/fix formatting details.")
		}
	}

	// Check goimports formatting
	if toolAvailable("goimports") {
		// Run goimports -l -local social-network cmd internal
		// #nosec G204
		cmd := ExecCommand("goimports", "-l", "-local", "social-network", "cmd", "internal")
		out, err := cmd.CombinedOutput()
		if err != nil || len(strings.TrimSpace(string(out))) > 0 {
			errors = append(errors, "Run 'goimports -w -local social-network cmd internal' or 'make format' to check/fix imports details.")
		}
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: "gate did not pass. " + strings.Join(errors, " "),
		}
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: "code formatting OK",
	}
}
