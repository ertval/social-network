/*
LintGate runs Go linter tools (golangci-lint, staticcheck, go vet) to ensure code hygiene.
*/
package gates

import (
	"fmt"
	"strings"
)

// LintGate runs linters.
type LintGate struct{}

func (g *LintGate) Name() string { return "lint" }

//nolint:nestif
func (g *LintGate) Run() Result {
	var errors []string

	if toolAvailable("golangci-lint") {
		// golangci-lint uses directories with /... suffix
		var lintDirs []string
		for _, dir := range NewDirs {
			lintDirs = append(lintDirs, dir+"/...")
		}
		args := append([]string{"run", "--timeout=5m"}, lintDirs...)
		// #nosec G204
		cmd := ExecCommand("golangci-lint", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			errors = append(errors, fmt.Sprintf("golangci-lint error: %v (output: %q).", err, string(out)))
		}
	} else {
		// Fallbacks: staticcheck and go vet use package paths (NewPkgs)
		if toolAvailable("staticcheck") {
			args := append([]string{}, NewPkgs...)
			// #nosec G204
			cmd := ExecCommand("staticcheck", args...)
			if out, err := cmd.CombinedOutput(); err != nil {
				errors = append(errors, fmt.Sprintf("staticcheck error: %v (output: %q).", err, string(out)))
			}
		}

		args := append([]string{"vet"}, NewPkgs...)
		// #nosec G204
		cmd := ExecCommand("go", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			errors = append(errors, fmt.Sprintf("go vet error: %v (output: %q).", err, string(out)))
		}
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: "gate did not pass. " + strings.Join(errors, " "),
		}
	}

	suffix := "golangci-lint"
	if !toolAvailable("golangci-lint") {
		if toolAvailable("staticcheck") {
			suffix = "staticcheck + go vet"
		} else {
			suffix = "go vet"
		}
	}
	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: "lint OK (" + suffix + ")",
	}
}
