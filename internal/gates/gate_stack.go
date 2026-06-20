package gates

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// StackGate validates Go version and module path (Gate #1).
type StackGate struct {
	GoModPath string // path to go.mod, defaults to "go.mod"
}

func (g *StackGate) Name() string { return "stack" }

func (g *StackGate) Run() Result {
	path := g.GoModPath
	if path == "" {
		path = "go.mod"
	}

	f, err := os.Open(path)
	if err != nil {
		return Result{Gate: g.Name(), Status: "FAIL", Message: fmt.Sprintf("cannot open go.mod: %v", err)}
	}
	defer f.Close()

	var goVersion, modulePath string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "go ") {
			goVersion = strings.TrimPrefix(line, "go ")
		}
		if strings.HasPrefix(line, "module ") {
			modulePath = strings.TrimPrefix(line, "module ")
		}
	}

	var errors []string
	if !strings.HasPrefix(goVersion, "1.24") {
		errors = append(errors, fmt.Sprintf("expected Go 1.24, got %s", goVersion))
	}
	if modulePath != "social-network" {
		errors = append(errors, fmt.Sprintf("expected module 'social-network', got '%s'", modulePath))
	}

	if len(errors) > 0 {
		return Result{Gate: g.Name(), Status: "FAIL", Message: strings.Join(errors, "; ")}
	}
	return Result{Gate: g.Name(), Status: "PASS", Message: "Go 1.24, module social-network"}
}
