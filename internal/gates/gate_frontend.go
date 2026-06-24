/*
FrontendGate runs frontend lint, format, typecheck, and test checks.
It checks for the presence of frontend-next or frontend directories, and runs
the corresponding package manager (bun) commands.
*/
package gates

import (
	"fmt"
	"os"
	"strings"
)

// FrontendGate runs frontend CI scripts.
type FrontendGate struct{}

func (g *FrontendGate) Name() string { return "frontend" }

func (g *FrontendGate) Run() Result {
	var dir string
	if _, err := os.Stat("frontend-next/package.json"); err == nil {
		dir = "frontend-next"
	} else if _, err := os.Stat("frontend/package.json"); err == nil {
		dir = "frontend"
	}

	if dir == "" {
		return Result{
			Gate:    g.Name(),
			Status:  "SKIP",
			Message: "no frontend scaffolded yet",
		}
	}

	steps := []struct {
		name    string
		command []string
	}{
		{"lint", []string{"bun", "run", "lint"}},
		{"format:check", []string{"bun", "run", "format:check"}},
		{"typecheck", []string{"bun", "x", "tsc", "--noEmit"}},
		{"test", []string{"bun", "run", "test"}},
	}

	for _, step := range steps {
		// #nosec G204
		cmd := ExecCommand(step.command[0], step.command[1:]...)
		cmd.Dir = dir
		if _, err := cmd.CombinedOutput(); err != nil {
			return Result{
				Gate:    g.Name(),
				Status:  "FAIL",
				Message: fmt.Sprintf("gate did not pass. Run 'cd %s && %s' to check details.", dir, strings.Join(step.command, " ")),
			}
		}
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: "frontend CI checks passed",
	}
}
