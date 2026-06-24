/*
FormatGate validates Go code formatting (gofumpt and goimports).
If formatting violations are found, it directs the user to run 'make format'.
*/
package gates

import (
	"fmt"
	"strings"
)

// FormatGate checks Go code formatting.
type FormatGate struct{}

func (g *FormatGate) Name() string { return "format" }

func (g *FormatGate) Run() Result {
	var errors []string

	// Check gofumpt formatting
	if toolAvailable("gofumpt") {
		args := append([]string{"-l"}, NewDirs...)
		// #nosec G204
		cmd := ExecCommand("gofumpt", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			errors = append(errors, fmt.Sprintf("gofumpt error: %v (output: %q).", err, string(out)))
		} else if len(strings.TrimSpace(string(out))) > 0 {
			errors = append(errors, "gofumpt found unformatted files. Run 'gofumpt -w <new_dirs>' or 'make format' to fix.")
		}
	}

	// Check goimports formatting
	if toolAvailable("goimports") {
		args := append([]string{"-l", "-local", "social-network"}, NewDirs...)
		// #nosec G204
		cmd := ExecCommand("goimports", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			errors = append(errors, fmt.Sprintf("goimports error: %v (output: %q).", err, string(out)))
		} else if len(strings.TrimSpace(string(out))) > 0 {
			errors = append(errors, "goimports found import issues. Run 'goimports -w -local social-network <new_dirs>' or 'make format' to fix.")
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
