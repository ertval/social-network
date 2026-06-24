/*
LintGate runs Go linter tools (golangci-lint, staticcheck, go vet) to ensure code hygiene.
*/
package gates

import (
	"fmt"
	"os"
	"strings"
)

// LintGate runs linters.
type LintGate struct{}

func (g *LintGate) Name() string { return "lint" }

func getLintDirs() []string {
	dirs := []string{"./cmd/server/...", "./cmd/gates/..."}
	if entries, err := os.ReadDir("internal"); err == nil {
		for _, e := range entries {
			if e.IsDir() && e.Name() != "app" && e.Name() != "infra" {
				dirs = append(dirs, fmt.Sprintf("./internal/%s/...", e.Name()))
			}
		}
	}
	return dirs
}

//nolint:nestif
func (g *LintGate) Run() Result {
	var errors []string
	targets := getLintDirs()

	if toolAvailable("golangci-lint") {
		// Run golangci-lint on target directories
		args := append([]string{"run", "--timeout=5m"}, targets...)
		// #nosec G204
		cmd := ExecCommand("golangci-lint", args...)
		if _, err := cmd.CombinedOutput(); err != nil {
			errors = append(errors, "Run 'golangci-lint run' to check details.")
		}
	} else {
		// Fallbacks: staticcheck and go vet
		if toolAvailable("staticcheck") {
			args := append([]string{}, targets...)
			// #nosec G204
			cmd := ExecCommand("staticcheck", args...)
			if _, err := cmd.CombinedOutput(); err != nil {
				errors = append(errors, "Run 'staticcheck ./...' to check details.")
			}
		}

		args := append([]string{"vet"}, targets...)
		// #nosec G204
		cmd := ExecCommand("go", args...)
		if _, err := cmd.CombinedOutput(); err != nil {
			errors = append(errors, "Run 'go vet ./...' to check details.")
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
