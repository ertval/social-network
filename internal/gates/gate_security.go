/*
SecurityGate runs static analysis security checks (Gate #8).
It executes 'gosec ./...' if available to find potential vulnerabilities, and
performs custom AST-based verification of bcrypt costs (must be >= 12)
and potential SQL string concatenation patterns.
*/
package gates

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SecurityGate checks security patterns (Gate #8).
// Primary: gosec. Custom AST checks always run (bcrypt cost, SQL concat).
type SecurityGate struct {
	InternalDir string // defaults to "internal"
}

func (g *SecurityGate) Name() string { return "security" }

//nolint:nestif
func (g *SecurityGate) Run() Result {
	var errors []string

	// Run gosec if available
	if toolAvailable("gosec") {
		args := []string{"-quiet"}
		// Include new command entries
		args = append(args, "./cmd/server/...", "./cmd/gates/...")

		dir := g.InternalDir
		if dir == "" {
			dir = "internal"
		}
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				if e.IsDir() && isFeatureSlice(dir, e.Name()) {
					args = append(args, fmt.Sprintf("./%s/%s/...", dir, e.Name()))
				}
			}
		}
		// #nosec G204
		cmd := ExecCommand("gosec", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			errors = append(errors, "gosec findings:\n"+string(out))
		}
	}

	// Always run custom AST checks (gosec doesn't cover bcrypt cost)
	errors = append(errors, g.runASTChecks()...)

	if len(errors) > 0 {
		return Result{Gate: g.Name(), Status: "FAIL", Message: strings.Join(errors, "; ")}
	}

	suffix := "gosec + AST"
	if !toolAvailable("gosec") {
		suffix = "AST only"
	}
	return Result{Gate: g.Name(), Status: "PASS", Message: fmt.Sprintf("security OK (%s)", suffix)}
}

func (g *SecurityGate) runASTChecks() []string {
	dir := g.InternalDir
	if dir == "" {
		dir = "internal"
	}

	var errors []string
	fset := token.NewFileSet()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []string{fmt.Sprintf("cannot read %s: %v", dir, err)}
	}

	inspectFile := func(path string) {
		f, parseErr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if parseErr != nil || f == nil {
			return
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				errors = append(errors, checkSQLConcat(fset, node, path)...)
				errors = append(errors, checkBcryptCost(fset, f, node, path)...)
			case *ast.KeyValueExpr, *ast.AssignStmt, *ast.FuncDecl:
				errors = append(errors, checkCheckOrigin(fset, node, path)...)
			}
			return true
		})
	}

	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if !e.IsDir() {
			if strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
				inspectFile(path)
			}
			continue
		}

		if !isFeatureSlice(dir, e.Name()) {
			continue
		}

		_ = filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(walkPath, ".go") || strings.HasSuffix(walkPath, "_test.go") {
				return nil
			}
			inspectFile(walkPath)
			return nil
		})
	}

	return errors
}

// checkSQLConcat detects fmt.Sprintf with SQL keywords (potential injection).
func checkSQLConcat(fset *token.FileSet, call *ast.CallExpr, path string) []string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok || ident.Name != "fmt" || sel.Sel.Name != "Sprintf" {
		return nil
	}
	if len(call.Args) == 0 {
		return nil
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return nil
	}

	sqlKeywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
	val := lit.Value
	for _, kw := range sqlKeywords {
		if strings.Contains(strings.ToUpper(val), kw) {
			pos := fset.Position(call.Pos())
			return []string{fmt.Sprintf("%s:%d potential SQL injection (fmt.Sprintf with %s)", path, pos.Line, kw)}
		}
	}
	return nil
}

// checkBcryptCost detects bcrypt.GenerateFromPassword calls with cost < 12 or bcrypt.DefaultCost.
func checkBcryptCost(fset *token.FileSet, f *ast.File, call *ast.CallExpr, path string) []string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "GenerateFromPassword" {
		return nil
	}
	// bcrypt.GenerateFromPassword(password, cost)
	if len(call.Args) < 2 {
		return nil
	}

	costExpr := call.Args[1]

	// Explicitly check for bcrypt.DefaultCost or bcrypt.MinCost
	if isBcryptDefaultCost(costExpr) {
		pos := fset.Position(call.Pos())
		selExpr := costExpr.(*ast.SelectorExpr) //nolint:forcetypeassert
		return []string{fmt.Sprintf("%s:%d bcrypt cost %s (%d) < 12", path, pos.Line, selExpr.Sel.Name, 10)}
	}

	cost, resolved := resolveCostExpr(f, costExpr)
	if !resolved {
		return nil // cost is a variable/constant we couldn't resolve
	}

	if cost < 12 {
		pos := fset.Position(call.Pos())
		return []string{fmt.Sprintf("%s:%d bcrypt cost %d < 12", path, pos.Line, cost)}
	}
	return nil
}

// isBcryptDefaultCost checks if an expression is bcrypt.DefaultCost or bcrypt.MinCost
func isBcryptDefaultCost(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "bcrypt" && (sel.Sel.Name == "DefaultCost" || sel.Sel.Name == "MinCost")
}

// resolveCostExpr attempts to determine the numeric bcrypt cost from the expression.
func resolveCostExpr(f *ast.File, expr ast.Expr) (int, bool) {
	// If it is literal
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT {
		val, err := strconv.Atoi(lit.Value)
		if err == nil {
			return val, true
		}
	}
	// If it is identifier, look up in the file's scope
	if ident, ok := expr.(*ast.Ident); ok {
		return findValueInFile(f, ident.Name)
	}
	// If it is bcrypt.DefaultCost or bcrypt.MinCost
	if isBcryptDefaultCost(expr) {
		sel := expr.(*ast.SelectorExpr) //nolint:forcetypeassert
		if sel.Sel.Name == "DefaultCost" {
			return 10, true
		}
		if sel.Sel.Name == "MinCost" {
			return 4, true
		}
	}
	return 0, false
}

// findValueInFile searches const and var declarations in f to find the value of name.
//
//nolint:gocognit,nestif
func findValueInFile(f *ast.File, name string) (int, bool) {
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || (gen.Tok != token.CONST && gen.Tok != token.VAR) {
			continue
		}
		for _, spec := range gen.Specs {
			vspec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, n := range vspec.Names {
				if n.Name == name {
					if i < len(vspec.Values) {
						valExpr := vspec.Values[i]
						// Check if it is a literal
						if lit, ok := valExpr.(*ast.BasicLit); ok && lit.Kind == token.INT {
							val, err := strconv.Atoi(lit.Value)
							if err == nil {
								return val, true
							}
						}
						// Check if it is bcrypt.DefaultCost or bcrypt.MinCost
						if isBcryptDefaultCost(valExpr) {
							sel := valExpr.(*ast.SelectorExpr) //nolint:forcetypeassert
							if sel.Sel.Name == "DefaultCost" {
								return 10, true
							}
							if sel.Sel.Name == "MinCost" {
								return 4, true
							}
						}
						// Check if it is another identifier (constant chain)
						if ident, ok := valExpr.(*ast.Ident); ok {
							return findValueInFile(f, ident.Name)
						}
					}
				}
			}
		}
	}
	return 0, false
}

// checkCheckOrigin detects WebSocket CheckOrigin returning true unconditionally.
//
//nolint:gocognit
func checkCheckOrigin(fset *token.FileSet, node ast.Node, path string) []string {
	var errors []string
	var funcLit *ast.FuncLit

	switch expr := node.(type) {
	case *ast.KeyValueExpr:
		keyIdent, ok := expr.Key.(*ast.Ident)
		if ok && keyIdent.Name == "CheckOrigin" {
			funcLit, _ = expr.Value.(*ast.FuncLit)
		}
	case *ast.AssignStmt:
		for _, lhs := range expr.Lhs {
			if isCheckOrigin(lhs) {
				for _, rhs := range expr.Rhs {
					if fl, ok := rhs.(*ast.FuncLit); ok {
						funcLit = fl
					}
				}
			}
		}
	case *ast.FuncDecl:
		if strings.Contains(expr.Name.Name, "CheckOrigin") && expr.Body != nil {
			for _, stmt := range expr.Body.List {
				ret, ok := stmt.(*ast.ReturnStmt)
				if !ok {
					continue
				}
				for _, res := range ret.Results {
					ident, ok := res.(*ast.Ident)
					if ok && ident.Name == "true" {
						pos := fset.Position(expr.Pos())
						errors = append(errors, fmt.Sprintf("%s:%d WebSocket CheckOrigin returns true unconditionally", path, pos.Line))
					}
				}
			}
		}
	}

	if funcLit != nil && funcLit.Body != nil {
		for _, stmt := range funcLit.Body.List {
			ret, ok := stmt.(*ast.ReturnStmt)
			if !ok {
				continue
			}
			for _, res := range ret.Results {
				ident, ok := res.(*ast.Ident)
				if ok && ident.Name == "true" {
					pos := fset.Position(funcLit.Pos())
					errors = append(errors, fmt.Sprintf("%s:%d WebSocket CheckOrigin returns true unconditionally", path, pos.Line))
				}
			}
		}
	}

	return errors
}

func isCheckOrigin(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name == "CheckOrigin"
	case *ast.SelectorExpr:
		return e.Sel.Name == "CheckOrigin"
	}
	return false
}
