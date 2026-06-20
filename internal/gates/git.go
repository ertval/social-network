package gates

import (
	"os/exec"
	"strings"
)

// FindBaseBranch resolves the correct base branch for comparison.
// Tries local main first, falls back to origin/main for CI/Gitea environments.
func FindBaseBranch() string {
	cmd := exec.Command("git", "merge-base", "main", "HEAD")
	if err := cmd.Run(); err == nil {
		return "main"
	}
	return "origin/main"
}

// GitLog returns commit subjects between base..HEAD.
func GitLog(base string) ([]string, error) {
	// #nosec G204
	cmd := exec.Command("git", "log", base+"..HEAD", "--format=%s")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

// GitBranch returns the current branch name.
func GitBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// GitDiffFiles returns filenames changed between base..HEAD.
func GitDiffFiles(base string) ([]string, error) {
	// #nosec G204
	cmd := exec.Command("git", "diff", base+"..HEAD", "--name-only")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}
