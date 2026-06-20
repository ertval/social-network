package gates

import (
	"fmt"
	"strings"
)

// ScopeDriftGate detects scope drift — advisory only (Gate #10).
type ScopeDriftGate struct{}

func (g *ScopeDriftGate) Name() string { return "scope-drift" }

func (g *ScopeDriftGate) Run() Result {
	branch := GitBranch()
	if branch == "main" || branch == "HEAD" {
		return Result{Gate: g.Name(), Status: "PASS", Message: "on main or detached HEAD"}
	}

	base := FindBaseBranch()
	files, err := GitDiffFiles(base)
	if err != nil || len(files) == 0 {
		return Result{Gate: g.Name(), Status: "PASS", Message: fmt.Sprintf("no changes from %s", base)}
	}

	// Advisory: always PASS, but report file count
	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("%d files changed (review for scope drift): %s", len(files), strings.Join(files[:min(len(files), 5)], ", ")),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
