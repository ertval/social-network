/*
BranchGate validates the current Git branch name against the pattern
'<username>/<ticket-ID>-<detail>' and checks that all commit messages
on the branch follow the Conventional Commits format.
*/
package gates

import (
	"fmt"
	"regexp"
	"strings"
)

// BranchGate validates branch naming and conventional commits (Gate #9).
type BranchGate struct{}

func (g *BranchGate) Name() string { return "branch" }

var (
	branchPattern = regexp.MustCompile(`^[a-z]+/[A-Za-z0-9-]+-[A-Za-z0-9-]+$`)
	commitPattern = regexp.MustCompile(`^(feat|fix|test|refactor|chore|docs|style|perf|ci|build|revert)\((user|topic|follow|group|event|chat|notification|oauth|core|platform|comment)\):`)
)

func (g *BranchGate) Run() Result {
	branch := GitBranch()

	if branch == "main" || branch == "HEAD" {
		return Result{Gate: g.Name(), Status: "PASS", Message: "on main or detached HEAD"}
	}

	var errors []string

	if !branchPattern.MatchString(branch) {
		errors = append(errors, fmt.Sprintf("branch '%s' doesn't match <username>/<ticket-ID>-<detail>", branch))
	}

	base := FindBaseBranch()
	commits, err := GitLog(base)
	if err == nil {
		for _, msg := range commits {
			if strings.HasPrefix(msg, "Merge ") {
				continue // skip auto-generated merge commits
			}
			if !commitPattern.MatchString(msg) {
				errors = append(errors, fmt.Sprintf("non-conventional commit: '%s'", msg))
			}
		}
	}

	if len(errors) > 0 {
		return Result{Gate: g.Name(), Status: "FAIL", Message: strings.Join(errors, "; ")}
	}
	return Result{Gate: g.Name(), Status: "PASS", Message: "branch and commits OK"}
}
