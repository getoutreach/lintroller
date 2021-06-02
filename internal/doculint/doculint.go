// Package doculint contains the necessary logic for the doculint linter.
package doculint

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer exports the doculint analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: "doculint",
	Doc:  "checks for proper function, type, package, constant, and string and numeric literal documentation",
	Run:  doculint,
}

// packageMain is a constant denoting the name of the "main" package in a Go program.
const packageMain = "main"

// funcMain is a constant denoting the name of the "main" function that exists in
// packageMain of a Go program.
const funcMain = "main"

// funcInit is a constant denoting the name of the "init" function that can exist in
// any package of a Go program.
const funcInit = "init"

// doculint is the function that gets passed to the Analyzer which runs the actual
// analysis for the doculint linter on a set of files.
func doculint(pass *analysis.Pass) (interface{}, error) { //nolint:funlen
	// packageWithSameNameFile keep track of which packages have a file with the same
	// name as the package and which do not (the convention is that this file will
	// contain the package documentation).
	packageWithSameNameFile := make(map[string]bool)

	// Validate the package name of the current pass, which is a single go package.
	validatePackageName(pass, pass.Pkg.Name())

	for _, file := range pass.Files {
		if pass.Pkg.Name() != packageMain {
			// Ignore the main package, it doesn't need a package comment.

			// Add this package to packageWithSameNameFile if it does not already
			// exist.
			if _, exists := packageWithSameNameFile[pass.Pkg.Name()]; !exists {
				packageWithSameNameFile[pass.Pkg.Name()] = false
			}

			// If the current file name matches the package name, examine the comment
			// that should exist within it.
			if file.Name.Name == pass.Pkg.Name() {
				packageWithSameNameFile[pass.Pkg.Name()] = true

				if file.Doc == nil {
					pass.Reportf(0, "package \"%s\" has no comment associated with it in \"%s.go\"", pass.Pkg.Name(), pass.Pkg.Name())
				} else {
					expectedPrefix := fmt.Sprintf("Package %s", pass.Pkg.Name())
					if !strings.HasPrefix(strings.TrimSpace(file.Doc.Text()), expectedPrefix) {
						pass.Reportf(0, "comment for package \"%s\" should begin with \"%s\"", pass.Pkg.Name(), expectedPrefix)
					}
				}
			}
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch expr := n.(type) {
			case *ast.FuncDecl:
				if pass.Pkg.Name() == packageMain && expr.Name.Name == funcMain {
					// Ignore func main in main package.
					return true
				}

				// Run through function declaration validation rules.
				validateFuncDecl(pass, expr)
			case *ast.IfStmt:
				// Run through if statement validation rules.
				validateIfStmt(pass, expr)
			case *ast.GenDecl:
				// Run through general declaration validation rules, currently these
				// only apply to constants and type declarations, as you will see if
				// you dig into the proceeding function call.
				validateGenDecl(pass, expr)
			default:
				return true
			}

			return true
		})
	}

	for pkg := range packageWithSameNameFile {
		if !packageWithSameNameFile[pkg] {
			pass.Reportf(0, "package \"%s\" has no file with the same name containing package comment", pkg)
		}
	}

	return nil, nil
}

// validateGenDecl validates an *ast.GenDecl to ensure it is up to doculint standards.
// Currently this function only looks for constants and type declarations and further
// validates them.
func validateGenDecl(pass *analysis.Pass, expr *ast.GenDecl) {
	if expr.Tok == token.CONST {
		validateGenDeclConstants(pass, expr)
	} else if expr.Tok == token.TYPE {
		validateGenDeclTypes(pass, expr)
	}
}

// validateGenDeclConstants validates an *ast.GenDecl that is a constant type. It ensures
// that if it is a constant block that the block itself has a comment, and each constant
// within it also has a comment. If it is a standalone constant it ensures that it has a
// comment associated with it.
func validateGenDeclConstants(pass *analysis.Pass, expr *ast.GenDecl) {
	if expr.Lparen.IsValid() {
		// Constant block
		if expr.Doc == nil {
			pass.Reportf(expr.Pos(), "constant block has no comment associated with it")
		}
	}

	for i := range expr.Specs {
		vs, ok := expr.Specs[i].(*ast.ValueSpec)
		if ok {
			if len(vs.Names) > 1 {
				var names []string
				for j := range vs.Names {
					names = append(names, vs.Names[j].Name)
				}

				pass.Reportf(vs.Pos(), "constants \"%s\" should be separated and each have a comment associated with them", strings.Join(names, ", "))
				continue
			}

			name := vs.Names[0].Name

			doc := vs.Doc
			if !expr.Lparen.IsValid() {
				// If this constant isn't apart of a constant block it's comment is stored in the *ast.GenDecl type.
				doc = expr.Doc
			}

			if doc == nil {
				pass.Reportf(vs.Pos(), "constant \"%s\" has no comment associated with it", name)
				continue
			}

			if !strings.HasPrefix(strings.TrimSpace(doc.Text()), name) {
				pass.Reportf(vs.Pos(), "comment for constant \"%s\" should begin with \"%s\"", name, name)
			}
		}
	}
}

// validateGenDeclTypes validates an *ast.GenDecl that is a type declaration. It ensures
// that if it is a type declaration block that the block itself has a comment, and each
// type declaration within it also has a comment. If it is a standalone type declaration
// it ensures that it has a comment associated with it.
func validateGenDeclTypes(pass *analysis.Pass, expr *ast.GenDecl) {
	if expr.Lparen.IsValid() {
		// Type block
		if expr.Doc == nil {
			pass.Reportf(expr.Pos(), "type block has no comment associated with it")
		}
	}

	for i := range expr.Specs {
		ts, ok := expr.Specs[i].(*ast.TypeSpec)
		if ok {
			doc := ts.Doc
			if !expr.Lparen.IsValid() {
				// If this type isn't apart of a type block it's comment is stored in the *ast.GenDecl type.
				doc = expr.Doc
			}

			if doc == nil {
				pass.Reportf(ts.Pos(), "type \"%s\" has no comment associated with it", ts.Name.Name)
				continue
			}

			if !strings.HasPrefix(strings.TrimSpace(doc.Text()), ts.Name.Name) {
				pass.Reportf(ts.Pos(), "comment for type \"%s\" should begin with \"%s\"", ts.Name.Name, ts.Name.Name)
			}
		}
	}
}

// validateIfStmt validates that an *ast.IfStmt upholds doculint standards. What this
// currently means is that it doesn't contain a condition statement that uses literals.
func validateIfStmt(pass *analysis.Pass, expr *ast.IfStmt) {
	be, ok := expr.Cond.(*ast.BinaryExpr)
	if !ok {
		// Ignore non-binary expressions in the conditional.
		return
	}

	if literal, ok := be.X.(*ast.BasicLit); ok {
		pass.Reportf(literal.Pos(), "literal found in conditional")
	}

	if literal, ok := be.Y.(*ast.BasicLit); ok {
		pass.Reportf(literal.Pos(), "literal found in conditional")
	}
}

// validateFuncDecl ensures that an *ast.FuncDecl upholds doculint standards by ensuring
// it has a corresponding comment that starts with the name of the function.
func validateFuncDecl(pass *analysis.Pass, expr *ast.FuncDecl) {
	if expr.Name.Name == funcInit {
		// Ignore init functions.
		return
	}

	if expr.Doc == nil {
		pass.Reportf(expr.Pos(), "function \"%s\" has no comment associated with it", expr.Name.Name)
	}

	if !strings.HasPrefix(strings.TrimSpace(expr.Doc.Text()), expr.Name.Name) {
		pass.Reportf(expr.Pos(), "comment for function \"%s\" should begin with \"%s\"", expr.Name.Name, expr.Name.Name)
	}
}

// validatePackageName ensures that a given package name follows the conventions that can
// be read about here: https://blog.golang.org/package-names
func validatePackageName(pass *analysis.Pass, pkg string) {
	if strings.ContainsAny(pkg, "_-") {
		pass.Reportf(0, "package \"%s\" should not contain - or _ in name", pkg)
	}

	if pkg != strings.ToLower(pkg) {
		pass.Reportf(0, "package \"%s\" should be all lowercase", pkg)
	}
}
