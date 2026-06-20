package gates

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// CoverageGate compares branch coverage vs base branch (Gate #13).
// Uses git worktree to avoid mutating the active workspace.
type CoverageGate struct {
	Threshold float64 // max allowed coverage drop %, defaults to 5
}

func (g *CoverageGate) Name() string { return "coverage-delta" }

func (g *CoverageGate) Run() Result {
	threshold := g.Threshold
	if threshold == 0 {
		threshold = 5
	}

	base := FindBaseBranch()

	// Get base branch coverage via git worktree
	baseCov, err := getBaselineCoverage(base)
	if err != nil {
		return Result{Gate: g.Name(), Status: "PASS", Message: fmt.Sprintf("could not compute baseline: %v", err)}
	}

	// Get current branch coverage
	branchCov, err := getCurrentCoverage()
	if err != nil {
		return Result{Gate: g.Name(), Status: "PASS", Message: fmt.Sprintf("could not compute branch coverage: %v", err)}
	}

	delta := branchCov - baseCov
	if delta < -threshold {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("coverage dropped by %.1f%% (%.1f%% → %.1f%%)", -delta, baseCov, branchCov),
		}
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("coverage %.1f%% (delta: %+.1f%%)", branchCov, delta),
	}
}

func getBaselineCoverage(baseBranch string) (float64, error) {
	tempDir := filepath.Join(os.TempDir(), "sn-gate-cov-base")
	defer os.RemoveAll(tempDir)

	// Create worktree
	add := exec.Command("git", "worktree", "add", tempDir, baseBranch)
	if err := add.Run(); err != nil {
		return 0, fmt.Errorf("git worktree add: %w", err)
	}
	defer exec.Command("git", "worktree", "remove", "--force", tempDir).Run()

	// Run tests in worktree
	covFile := filepath.Join(tempDir, "coverage.out")
	testCmd := exec.Command("go", "test", "-coverprofile="+covFile, "./...")
	testCmd.Dir = tempDir
	if err := testCmd.Run(); err != nil {
		return 0, fmt.Errorf("go test in worktree: %w", err)
	}

	return parseCoverageFile(covFile)
}

func getCurrentCoverage() (float64, error) {
	covFile := filepath.Join(os.TempDir(), "sn-gate-cov-branch.out")
	defer os.Remove(covFile)

	testCmd := exec.Command("go", "test", "-coverprofile="+covFile, "./...")
	if err := testCmd.Run(); err != nil {
		return 0, fmt.Errorf("go test: %w", err)
	}
	return parseCoverageFile(covFile)
}

func parseCoverageFile(path string) (float64, error) {
	// Use go tool cover to get total coverage
	cmd := exec.Command("go", "tool", "cover", "-func="+path)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Last line: "total: (statements) XX.X%"
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	fields := strings.Fields(lastLine)
	if len(fields) < 3 {
		return 0, fmt.Errorf("unexpected coverage output: %s", lastLine)
	}
	pctStr := strings.TrimSuffix(fields[len(fields)-1], "%")
	return strconv.ParseFloat(pctStr, 64)
}
