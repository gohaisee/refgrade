package check

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const err01ID = "err-01"

// Err01 — _ = err и пустой if err != nil {} без обработки
type Err01 struct{}

func NewErr01() *Err01 {
	return &Err01{}
}

func (c *Err01) ID() string {
	return err01ID
}

func (c *Err01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	var findings []Finding
	fset := token.NewFileSet()

	for _, gf := range mod.GoSourceFiles() {
		if skipErr01File(gf.RelPath) {
			continue
		}

		src, err := os.ReadFile(gf.Path)
		if err != nil {
			return nil, err
		}

		file, err := parser.ParseFile(fset, gf.Path, src, 0)
		if err != nil {
			return nil, err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.AssignStmt:
				if isIgnoredErrAssign(node) {
					pos := fset.Position(node.Pos())
					findings = append(findings, Finding{
						ID:       err01ID,
						Severity: SeverityWarn,
						File:     gf.RelPath,
						Line:     pos.Line,
					})
				}
			case *ast.IfStmt:
				if isEmptyErrNilCheck(node) {
					pos := fset.Position(node.Pos())
					findings = append(findings, Finding{
						ID:       err01ID,
						Severity: SeverityWarn,
						File:     gf.RelPath,
						Line:     pos.Line,
					})
				}
			}
			return true
		})
	}
	return findings, nil
}

func skipErr01File(rel string) bool {
	return strings.HasSuffix(filepath.Base(rel), "_test.go")
}

func isIgnoredErrAssign(stmt *ast.AssignStmt) bool {
	if stmt.Tok != token.ASSIGN {
		return false
	}
	if len(stmt.Lhs) != 1 || len(stmt.Rhs) != 1 {
		return false
	}
	lhs, ok := stmt.Lhs[0].(*ast.Ident)
	if !ok || lhs.Name != "_" {
		return false
	}
	rhs, ok := stmt.Rhs[0].(*ast.Ident)
	if !ok || rhs.Name != "err" {
		return false
	}
	return true
}

func isEmptyErrNilCheck(stmt *ast.IfStmt) bool {
	if stmt.Body == nil || len(stmt.Body.List) != 0 {
		return false
	}
	return isErrNotNilExpr(stmt.Cond)
}

func isErrNotNilExpr(expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}
	return isErrIdent(bin.X) && isNilIdent(bin.Y) || isErrIdent(bin.Y) && isNilIdent(bin.X)
}

func isErrIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "err"
}

func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}
