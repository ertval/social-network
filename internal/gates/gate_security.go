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

// SecurityGate checks security patterns via AST (Gate #8).
type SecurityGate struct {
	InternalDir string // defaults to "internal"
}

func (g *SecurityGate) Name() string { return "security" }

func (g *SecurityGate) Run() Result {
	dir := g.InternalDir
	if dir == "" {
		dir = "internal"
	}

	var errors []string
	fset := token.NewFileSet()

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		f, parseErr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if parseErr != nil || f == nil {
			return nil
		}

		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				errors = append(errors, checkSQLConcat(fset, node, path)...)
				errors = append(errors, checkBcryptCost(fset, node, path)...)
			}
			return true
		})
		return nil
	})

	if len(errors) > 0 {
		return Result{Gate: g.Name(), Status: "FAIL", Message: strings.Join(errors, "; ")}
	}
	return Result{Gate: g.Name(), Status: "PASS", Message: "security OK"}
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

// checkBcryptCost detects bcrypt.GenerateFromPassword calls with cost < 12.
func checkBcryptCost(fset *token.FileSet, call *ast.CallExpr, path string) []string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "GenerateFromPassword" {
		return nil
	}
	// bcrypt.GenerateFromPassword(password, cost)
	if len(call.Args) < 2 {
		return nil
	}
	costArg, ok := call.Args[1].(*ast.BasicLit)
	if !ok || costArg.Kind != token.INT {
		return nil // cost is a variable/constant — can't statically check
	}
	cost, err := strconv.Atoi(costArg.Value)
	if err != nil {
		return nil
	}
	if cost < 12 {
		pos := fset.Position(call.Pos())
		return []string{fmt.Sprintf("%s:%d bcrypt cost %d < 12", path, pos.Line, cost)}
	}
	return nil
}
