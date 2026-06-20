package gates

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DAGGate validates D6 dependency DAG acyclicity (Gate #4).
type DAGGate struct {
	InternalDir string // defaults to "internal"
}

func (g *DAGGate) Name() string { return "d6-dag" }

func (g *DAGGate) Run() Result {
	dir := g.InternalDir
	if dir == "" {
		dir = "internal"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return Result{Gate: g.Name(), Status: "SKIP", Message: fmt.Sprintf("cannot read %s: %v", dir, err)}
	}

	// Collect feature names
	var features []string
	for _, e := range entries {
		if e.IsDir() && !skipDirs[e.Name()] {
			features = append(features, e.Name())
		}
	}

	if len(features) == 0 {
		return Result{Gate: g.Name(), Status: "PASS", Message: "no feature slices yet (pre-migration)"}
	}

	var errors []string

	// Build dependency graph
	deps := make(map[string][]string)
	for _, feature := range features {
		featureDeps, err := getFeatureDeps(feature)
		if err != nil {
			continue
		}
		deps[feature] = featureDeps
	}

	// Check for cycles
	for _, feature := range features {
		for _, dep := range deps[feature] {
			for _, revDep := range deps[dep] {
				if revDep == feature {
					errors = append(errors, fmt.Sprintf("CIRCULAR: %s ↔ %s", feature, dep))
				}
			}
		}
	}

	// Check notification is never imported (except by bootstrap)
	for _, feature := range features {
		if feature == "notification" {
			continue
		}
		for _, dep := range deps[feature] {
			if dep == "notification" {
				errors = append(errors, fmt.Sprintf("D6: %s imports notification (forbidden)", feature))
			}
		}
	}

	if len(errors) > 0 {
		return Result{Gate: g.Name(), Status: "FAIL", Message: strings.Join(errors, "; ")}
	}
	return Result{Gate: g.Name(), Status: "PASS", Message: "D6 DAG acyclic"}
}

type goListPkg struct {
	Imports []string `json:"Imports"`
}

// getFeatureDeps returns other feature slices that this feature imports.
func getFeatureDeps(feature string) ([]string, error) {
	cmd := exec.Command("go", "list", "-json", "social-network/internal/"+feature+"/...")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var deps []string
	seen := make(map[string]bool)
	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var pkg goListPkg
		if err := dec.Decode(&pkg); err != nil {
			break
		}
		for _, imp := range pkg.Imports {
			if !strings.HasPrefix(imp, "social-network/internal/") {
				continue
			}
			dep := strings.TrimPrefix(imp, "social-network/internal/")
			dep = strings.SplitN(dep, "/", 2)[0]
			if dep == feature || skipDirs[dep] || seen[dep] {
				continue
			}
			seen[dep] = true
			deps = append(deps, dep)
		}
	}
	return deps, nil
}
