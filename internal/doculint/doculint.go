// Package doculint contains the necessary logic for the doculint linter.
package doculint

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/getoutreach/lintroller/internal/common"
	"golang.org/x/tools/go/analysis"
)

// Analyzer exports the doculint analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: "doculint",
	Doc:  "checks for proper function, type, package, constant, and string and numeric literal documentation",
	Run:  doculint,
}

// Variable block to keep track of flags whose values are collected at runtime. See the
// init function that immediately proceeds this block to see more.
var (
	// minFunLen is a variable that gets collected via flags. This variable contains a
	// the minimum function length that doculint will report on if said function has no
	// related documentation.
	minFunLen int
)

func init() { //nolint:gochecknoinits
	Analyzer.Flags.IntVar(&minFunLen, "minFunLen", 10, "the minimum function length that doculint will report on if said function has no related documentation")
}

// doculint is the function that gets passed to the Analyzer which runs the actual
// analysis for the doculint linter on a set of files.
func doculint(pass *analysis.Pass) (interface{}, error) { //nolint:funlen
	// Variable to keep track of whether or not this current package has a file with
	// the same name as the package. This is where the package comment should exist.
	var packageHasFileWithSameName bool

	// Validate the package name of the current pass, which is a single go package.
	validatePackageName(pass, pass.Pkg.Name())

	for _, file := range pass.Files {
		// Pull file into a local variable so it can be passed as a parameter safely.
		file := file

		if pass.Pkg.Name() == common.PackageMain {
			// Ignore the main package, it doesn't need a package comment.
			packageHasFileWithSameName = true
		} else {
			// Get current filepath.
			fp := pass.Fset.PositionFor(file.Package, false).Filename
			splitPath := strings.Split(fp, string(os.PathSeparator))

			// Extract filename from path and remove the ".go" suffix.
			fn := strings.TrimSuffix(splitPath[len(splitPath)-1], ".go")

			// If the current file name matches the package name, examine the comment
			// that should exist within it.
			if fn == pass.Pkg.Name() {
				packageHasFileWithSameName = true

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
				if pass.Pkg.Name() == common.PackageMain && expr.Name.Name == common.FuncMain {
					// Ignore func main in main package.
					return true
				}

				start := pass.Fset.PositionFor(expr.Pos(), false).Line
				end := pass.Fset.PositionFor(expr.End(), false).Line

				// The reason a 1 is added is to account for single-line functions (edge case).
				// This also doesn't affect non-single line functions, it will just account for
				// the trailing } which is what most people would expect anyways when providing
				// a minimum function length to validate against.
				if (end - start + 1) >= minFunLen {
					// Run through function declaration validation rules if the minimum function
					// length is met or exceeded.
					validateFuncDecl(pass, expr)
				}
			case *ast.GenDecl:
				// Run through general declaration validation rules, currently these
				// only apply to constants, type, and variable declarations, as you
				// will see if you dig into the proceeding function call.
				validateGenDecl(pass, expr)
			default:
				return true
			}

			return true
		})
	}

	if !packageHasFileWithSameName {
		pass.Reportf(0, "package \"%s\" has no file with the same name containing package comment", pass.Pkg.Name())
	}

	return nil, nil
}

// validateGenDecl validates an *ast.GenDecl to ensure it is up to doculint standards.
// Currently this function only looks for constants, type, and variable declarations
// then further validates them.
func validateGenDecl(pass *analysis.Pass, expr *ast.GenDecl) {
	switch expr.Tok { //nolint:exhaustive
	case token.CONST:
		validateGenDeclConstants(pass, expr)
	case token.TYPE:
		validateGenDeclTypes(pass, expr)
	case token.VAR:
		validateGenDeclVariables(pass, expr)
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

// validateGenDeclVariables validates an *ast.GenDecl that is a variable type. It ensures
// that if it is a variable block that the block itself has a comment, and each variable
// within it also has a comment. If it is a standalone variable it ensures that it has a
// comment associated with it.
func validateGenDeclVariables(pass *analysis.Pass, expr *ast.GenDecl) {
	if expr.Lparen.IsValid() {
		// Variable block
		if expr.Doc == nil {
			pass.Reportf(expr.Pos(), "variable block has no comment associated with it")
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

				pass.Reportf(vs.Pos(), "variables \"%s\" should be separated and each have a comment associated with them", strings.Join(names, ", "))
				continue
			}

			name := vs.Names[0].Name

			doc := vs.Doc
			if !expr.Lparen.IsValid() {
				// If this variable isn't apart of a variable block it's comment is stored in the *ast.GenDecl type.
				doc = expr.Doc
			}

			if doc == nil {
				pass.Reportf(vs.Pos(), "variable \"%s\" has no comment associated with it", name)
				continue
			}

			if !strings.HasPrefix(strings.TrimSpace(doc.Text()), name) {
				pass.Reportf(vs.Pos(), "comment for variable \"%s\" should begin with \"%s\"", name, name)
			}
		}
	}
}

// validateFuncDecl ensures that an *ast.FuncDecl upholds doculint standards by ensuring
// it has a corresponding comment that starts with the name of the function.
func validateFuncDecl(pass *analysis.Pass, expr *ast.FuncDecl) {
	if expr.Name.Name == common.FuncInit {
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
