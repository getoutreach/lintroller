// Package doculint contains the necessary logic for the doculint linter.
package doculint

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/getoutreach/lintroller/internal/common"
	"github.com/getoutreach/lintroller/internal/nolint"
	"golang.org/x/tools/go/analysis"
)

// name defines the name of the doculint linter.
const name = "doculint"

// doc defines the help text for the doculint linter.
const doc = `Checks for proper function, type, package, constant, and string and numeric 
literal documentation in accordance with godoc standards.`

// Analyzer exports the doculint analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: name,
	Doc:  doc,
	Run:  doculint,
}

// NewAnalyzerWithOptions returns the Analyzer package-level variable, with the options
// that would have been defined via flags if this was ran as a vet tool. This is so the
// analyzers can be ran outside of the context of a vet tool and config can be gathered
// from elsewhere.
func NewAnalyzerWithOptions(_minFunLen int, _validatePackages, _validateFunctions, _validateVariables, _validateConstants, _validateTypes bool) *analysis.Analyzer {
	minFunLen = _minFunLen
	validatePackages = _validatePackages
	validateFunctions = _validateFunctions
	validateVariables = _validateVariables
	validateConstants = _validateConstants
	validateTypes = _validateTypes

	return &Analyzer
}

// Variable block to keep track of flags whose values are collected at runtime. See the
// init function that immediately proceeds this block to see more.
var (
	// minFunLen is a variable that gets collected via flags. This variable contains a
	// the minimum function length that doculint will report on if said function has no
	// related documentation.
	minFunLen int

	// validatePackages is a variable that gets collected via flags. This variable contains
	// a flag that denotes whether or not the linter should validate that packages have
	// satisfactory comments.
	validatePackages bool

	// validateFunctions is a variable that gets collected via flags. This variable contains
	// a flag that denotes whether or not the linter should validate that functions have
	// satisfactory comments.
	validateFunctions bool

	// validateVariables is a variable that gets collected via flags. This variable contains
	// a flag that denotes whether or not the linter should validate that variables have
	// satisfactory comments.
	validateVariables bool

	// validateConstants is a variable that gets collected via flags. This variable contains
	// a flag that denotes whether or not the linter should validate that constants have
	// satisfactory comments.
	validateConstants bool

	// validateTypes is a variable that gets collected via flags. This variable contains a
	// flag that denotes whether or not the linter should validate that types have
	// satisfactory comments.
	validateTypes bool
)

func init() { //nolint:gochecknoinits
	Analyzer.Flags.IntVar(&minFunLen, "minFunLen", 10, "the minimum function length that doculint will report on if said function has no related documentation")
	Analyzer.Flags.BoolVar(&validatePackages, "validatePackages", true, "a boolean flag that denotes whether or not to validate package comments")
	Analyzer.Flags.BoolVar(&validateFunctions, "validateFunctions", true, "a boolean flag that denotes whether or not to validate function comments")
	Analyzer.Flags.BoolVar(&validateVariables, "validateVariables", true, "a boolean flag that denotes whether or not to validate variable comments")
	Analyzer.Flags.BoolVar(&validateConstants, "validateConstants", true, "a boolean flag that denotes whether or not to validate constant comments")
	Analyzer.Flags.BoolVar(&validateTypes, "validateTypes", true, "a boolean flag that denotes whether or not to validate type comments")

	if minFunLen == 0 {
		minFunLen = 10
	}
}

// doculint is the function that gets passed to the Analyzer which runs the actual
// analysis for the doculint linter on a set of files.
func doculint(pass *analysis.Pass) (interface{}, error) { //nolint:funlen
	// Wrap pass with nolint.Pass to take nolint directives into account.
	passWithNoLint := nolint.PassWithNoLint(name, pass)

	// Variable to keep track of whether or not this current package has a file with
	// the same name as the package. This is where the package comment should exist.
	var packageHasFileWithSameName bool

	// Validate the package name of the current pass, which is a single go package.
	validatePackageName(passWithNoLint, passWithNoLint.Pkg.Name())

	for _, file := range passWithNoLint.Files {
		// Pull file into a local variable so it can be passed as a parameter safely.
		file := file

		if passWithNoLint.Pkg.Name() == common.PackageMain || !validatePackages {
			// Ignore the main package, it doesn't need a package comment, and ignore package comment
			// checks if the validatePackages flag was set to false.
			packageHasFileWithSameName = true
		} else {
			// Get current filepath.
			fp := passWithNoLint.Fset.PositionFor(file.Package, false).Filename
			splitPath := strings.Split(fp, string(os.PathSeparator))

			// Extract filename from path and remove the ".go" suffix.
			fn := strings.TrimSuffix(splitPath[len(splitPath)-1], ".go")

			// If the current file name matches the package name, examine the comment
			// that should exist within it.
			if fn == passWithNoLint.Pkg.Name() {
				packageHasFileWithSameName = true

				if file.Doc == nil {
					passWithNoLint.Reportf(0, "package \"%s\" has no comment associated with it in \"%s.go\"", passWithNoLint.Pkg.Name(), passWithNoLint.Pkg.Name())
				} else {
					expectedPrefix := fmt.Sprintf("Package %s", passWithNoLint.Pkg.Name())
					if !strings.HasPrefix(strings.TrimSpace(file.Doc.Text()), expectedPrefix) {
						passWithNoLint.Reportf(0, "comment for package \"%s\" should begin with \"%s\"", passWithNoLint.Pkg.Name(), expectedPrefix)
					}
				}
			}
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch expr := n.(type) {
			case *ast.FuncDecl:
				if !validateFunctions {
					// validateFunctions flag was set to false, ignore all functions.
					return true
				}

				if passWithNoLint.Pkg.Name() == common.PackageMain && expr.Name.Name == common.FuncMain {
					// Ignore func main in main package.
					return true
				}

				start := passWithNoLint.Fset.PositionFor(expr.Pos(), false).Line
				end := passWithNoLint.Fset.PositionFor(expr.End(), false).Line

				// The reason a 1 is added is to account for single-line functions (edge case).
				// This also doesn't affect non-single line functions, it will just account for
				// the trailing } which is what most people would expect anyways when providing
				// a minimum function length to validate against.
				fmt.Println("------------>>>>>>>>>>> using minFunLen", minFunLen)
				if (end - start + 1) >= minFunLen {
					// Run through function declaration validation rules if the minimum function
					// length is met or exceeded.
					validateFuncDecl(passWithNoLint, expr)
				}
			case *ast.GenDecl:
				// Run through general declaration validation rules, currently these
				// only apply to constants, type, and variable declarations, as you
				// will see if you dig into the proceeding function call.
				validateGenDecl(passWithNoLint, expr)
			default:
				return true
			}

			return true
		})
	}

	if !packageHasFileWithSameName {
		passWithNoLint.Reportf(0, "package \"%s\" has no file with the same name containing package comment", passWithNoLint.Pkg.Name())
	}

	return nil, nil
}

// validateGenDecl validates an *ast.GenDecl to ensure it is up to doculint standards.
// Currently this function only looks for constants, type, and variable declarations
// then further validates them.
func validateGenDecl(reporter nolint.Reporter, expr *ast.GenDecl) {
	switch expr.Tok { //nolint:exhaustive
	case token.CONST:
		if validateConstants {
			// validateConstants flag was set to true, go ahead and validate constants.
			validateGenDeclConstants(reporter, expr)
		}
	case token.TYPE:
		if validateTypes {
			// validateTypes flag was set to true, go ahead and validate types.
			validateGenDeclTypes(reporter, expr)
		}
	case token.VAR:
		if validateVariables {
			// validateVariables flag was set to true, go ahead and validate variables.
			validateGenDeclVariables(reporter, expr)
		}
	}
}

// validateGenDeclConstants validates an *ast.GenDecl that is a constant type. It ensures
// that if it is a constant block that the block itself has a comment, and each constant
// within it also has a comment. If it is a standalone constant it ensures that it has a
// comment associated with it.
func validateGenDeclConstants(reporter nolint.Reporter, expr *ast.GenDecl) {
	if expr.Lparen.IsValid() {
		// Constant block
		if expr.Doc == nil {
			reporter.Reportf(expr.Pos(), "constant block has no comment associated with it")
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

				reporter.Reportf(vs.Pos(), "constants \"%s\" should be separated and each have a comment associated with them", strings.Join(names, ", "))
				continue
			}

			name := vs.Names[0].Name

			doc := vs.Doc
			if !expr.Lparen.IsValid() {
				// If this constant isn't apart of a constant block it's comment is stored in the *ast.GenDecl type.
				doc = expr.Doc
			}

			if doc == nil {
				reporter.Reportf(vs.Pos(), "constant \"%s\" has no comment associated with it", name)
				continue
			}

			if !strings.HasPrefix(strings.TrimSpace(doc.Text()), name) {
				reporter.Reportf(vs.Pos(), "comment for constant \"%s\" should begin with \"%s\"", name, name)
			}
		}
	}
}

// validateGenDeclTypes validates an *ast.GenDecl that is a type declaration. It ensures
// that if it is a type declaration block that the block itself has a comment, and each
// type declaration within it also has a comment. If it is a standalone type declaration
// it ensures that it has a comment associated with it.
func validateGenDeclTypes(reporter nolint.Reporter, expr *ast.GenDecl) {
	if expr.Lparen.IsValid() {
		// Type block
		if expr.Doc == nil {
			reporter.Reportf(expr.Pos(), "type block has no comment associated with it")
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
				reporter.Reportf(ts.Pos(), "type \"%s\" has no comment associated with it", ts.Name.Name)
				continue
			}

			if !strings.HasPrefix(strings.TrimSpace(doc.Text()), ts.Name.Name) {
				reporter.Reportf(ts.Pos(), "comment for type \"%s\" should begin with \"%s\"", ts.Name.Name, ts.Name.Name)
			}
		}
	}
}

// validateGenDeclVariables validates an *ast.GenDecl that is a variable type. It ensures
// that if it is a variable block that the block itself has a comment, and each variable
// within it also has a comment. If it is a standalone variable it ensures that it has a
// comment associated with it.
func validateGenDeclVariables(reporter nolint.Reporter, expr *ast.GenDecl) {
	if expr.Lparen.IsValid() {
		// Variable block
		if expr.Doc == nil {
			reporter.Reportf(expr.Pos(), "variable block has no comment associated with it")
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

				reporter.Reportf(vs.Pos(), "variables \"%s\" should be separated and each have a comment associated with them", strings.Join(names, ", "))
				continue
			}

			name := vs.Names[0].Name

			doc := vs.Doc
			if !expr.Lparen.IsValid() {
				// If this variable isn't apart of a variable block it's comment is stored in the *ast.GenDecl type.
				doc = expr.Doc
			}

			if doc == nil {
				reporter.Reportf(vs.Pos(), "variable \"%s\" has no comment associated with it", name)
				continue
			}

			if !strings.HasPrefix(strings.TrimSpace(doc.Text()), name) {
				reporter.Reportf(vs.Pos(), "comment for variable \"%s\" should begin with \"%s\"", name, name)
			}
		}
	}
}

// validateFuncDecl ensures that an *ast.FuncDecl upholds doculint standards by ensuring
// it has a corresponding comment that starts with the name of the function.
func validateFuncDecl(reporter nolint.Reporter, expr *ast.FuncDecl) {
	if expr.Name.Name == common.FuncInit {
		// Ignore init functions.
		return
	}

	if expr.Doc == nil {
		reporter.Reportf(expr.Pos(), "function \"%s\" has no comment associated with it", expr.Name.Name)
		return
	}

	if !strings.HasPrefix(strings.TrimSpace(expr.Doc.Text()), expr.Name.Name) {
		reporter.Reportf(expr.Pos(), "comment for function \"%s\" should begin with \"%s\"", expr.Name.Name, expr.Name.Name)
	}
}

// validatePackageName ensures that a given package name follows the conventions that can
// be read about here: https://blog.golang.org/package-names
func validatePackageName(reporter nolint.Reporter, pkg string) {
	if strings.ContainsAny(pkg, "_-") {
		reporter.Reportf(0, "package \"%s\" should not contain - or _ in name", pkg)
	}

	if pkg != strings.ToLower(pkg) {
		reporter.Reportf(0, "package \"%s\" should be all lowercase", pkg)
	}
}
