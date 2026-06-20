package gates

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// BoundariesGate validates D5 import boundary rules via AST (Gate #3).
type BoundariesGate struct {
	InternalDir string // defaults to "internal"
}

func (g *BoundariesGate) Name() string { return "d5-boundaries" }

func (g *BoundariesGate) Run() Result {
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
		featureDir := filepath.Join(dir, feature)

		// Rule: transport/ must not import store/
		errors = append(errors, checkImports(featureDir, "transport", []string{"/store"})...)

		// Rule: store/ must not import transport/, commands/, queries/
		errors = append(errors, checkImports(featureDir, "store", []string{"/transport", "/commands", "/queries"})...)

		// Rule: commands/ must not import store/ or transport/
		errors = append(errors, checkImports(featureDir, "commands", []string{"/store", "/transport"})...)

		// Rule: queries/ must not import store/ or transport/
		errors = append(errors, checkImports(featureDir, "queries", []string{"/store", "/transport"})...)
	}

	if len(errors) > 0 {
		return Result{Gate: g.Name(), Status: "FAIL", Message: strings.Join(errors, "; ")}
	}
	return Result{Gate: g.Name(), Status: "PASS", Message: "D5 boundaries OK"}
}

// checkImports parses Go files in a subdirectory and checks for forbidden import path fragments.
func checkImports(featureDir, subDir string, forbidden []string) []string {
	dir := filepath.Join(featureDir, subDir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	var violations []string
	fset := token.NewFileSet()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return nil // skip unparseable files
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if !strings.Contains(importPath, "internal/") {
				continue
			}
			for _, frag := range forbidden {
				if strings.Contains(importPath, frag) {
					pos := fset.Position(imp.Pos())
					violations = append(violations, fmt.Sprintf("D5: %s:%d imports %s", pos.Filename, pos.Line, importPath))
				}
			}
		}
		return nil
	})
	if err != nil {
		// Silently skip walk errors
	}

	return violations
}
