package gates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TDDGate verifies test files exist for feature code (Gate #6).
type TDDGate struct {
	InternalDir string // defaults to "internal"
}

func (g *TDDGate) Name() string { return "tdd" }

func (g *TDDGate) Run() Result {
	dir := g.InternalDir
	if dir == "" {
		dir = "internal"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return Result{Gate: g.Name(), Status: "SKIP", Message: fmt.Sprintf("cannot read %s: %v", dir, err)}
	}

	var errors []string
	for _, e := range entries {
		if !e.IsDir() || skipDirs[e.Name()] {
			continue
		}
		feature := e.Name()
		cmdDir := filepath.Join(dir, feature, "commands")
		errors = append(errors, checkTestCoverage(cmdDir)...)
	}

	if len(errors) > 0 {
		return Result{Gate: g.Name(), Status: "FAIL", Message: strings.Join(errors, "; ")}
	}
	return Result{Gate: g.Name(), Status: "PASS", Message: "TDD OK"}
}

// checkTestCoverage verifies a directory has test files if it has Go source files.
func checkTestCoverage(dir string) []string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	hasGoFiles := false
	hasTestFiles := false

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, "_test.go") {
			hasTestFiles = true
		} else if strings.HasSuffix(name, ".go") {
			hasGoFiles = true
		}
	}

	if hasGoFiles && !hasTestFiles {
		return []string{fmt.Sprintf("%s: has Go files but no test files", dir)}
	}
	return nil
}
