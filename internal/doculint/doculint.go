// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: See package comment for this one file package.

// Package doculint contains the necessary logic for the doculint linter. The doculint
// linter ensures proper documentation on various types, functions, variables, constants,
// etc. in the form of comments.
package doculint

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/getoutreach/lintroller/internal/common"
	"github.com/getoutreach/lintroller/internal/reporter"
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
func NewAnalyzerWithOptions(
	_minFunLen int, _validatePackages, _validateFunctions, _validateVariables, _validateConstants, _validateTypes bool) *analysis.Analyzer {
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

func init() { //nolint:gochecknoinits // Why: This is necessary to grab flags.
	Analyzer.Flags.IntVar(
		&minFunLen, "minFunLen", 10, "the minimum function length that doculint will report on if said function has no related documentation")
	Analyzer.Flags.BoolVar(
		&validatePackages, "validatePackages", true, "a boolean flag that denotes whether or not to validate package comments")
	Analyzer.Flags.BoolVar(
		&validateFunctions, "validateFunctions", true, "a boolean flag that denotes whether or not to validate function comments")
	Analyzer.Flags.BoolVar(
		&validateVariables, "validateVariables", true, "a boolean flag that denotes whether or not to validate variable comments")
	Analyzer.Flags.BoolVar(
		&validateConstants, "validateConstants", true, "a boolean flag that denotes whether or not to validate constant comments")
	Analyzer.Flags.BoolVar(
		&validateTypes, "validateTypes", true, "a boolean flag that denotes whether or not to validate type comments")

	if minFunLen == 0 {
		minFunLen = 10
	}
}

// doculint is the function that gets passed to the Analyzer which runs the actual
// analysis for the doculint linter on a set of files.
func doculint(_pass *analysis.Pass) (interface{}, error) { //nolint:funlen // Why: Doesn't make sense to break this function up anymore.
	// Ignore test packages.
	if common.IsTestPackage(_pass) {
		return nil, nil
	}

	// Wrap _pass with reporter.Pass to take nolint directives into account.
	pass := reporter.NewPass(name, _pass)

	// Variable to keep track of whether or not this current package has a file with
	// the same name as the package. This is where the package comment should exist.
	var packageHasFileWithSameName bool

	// allGenerated is a flag to denote whether or not an entire package was generated.
	// This will bypass the package comment reporting.
	allGenerated := true

	for _, file := range pass.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(file) || common.IsTestFile(pass.Pass, file) {
			continue
		}

		// We've made it past the generated check, make sure to denote that at least one file in the
		// package was not generated.
		allGenerated = false

		if pass.Pkg.Name() == common.PackageMain || !validatePackages {
			// Ignore the main package, it doesn't need a package comment, and ignore package comment
			// checks if the validatePackages flag was set to false.
			packageHasFileWithSameName = true
		} else {
			// Get current filepath.
			fp := pass.Fset.PositionFor(file.Package, false).Filename
			splitPath := strings.Split(fp, string(os.PathSeparator))

			// Extract filename from path and remove the ".go" suffix.
			fn := strings.TrimSuffix(splitPath[len(splitPath)-1], ".go")

			// If the current file name matches the package name, examine the comment
			// that should exist within it.
			if fn == pass.Pkg.Name() || fn == common.DocFilenameWithoutPath {
				packageHasFileWithSameName = true

				// This is the file we'd report the bad package name on, so run the validation here.
				validatePackageName(pass, file.Package, pass.Pkg.Name())

				if file.Doc == nil {
					pass.Reportf(
						file.Package,
						"package \"%s\" has no comment associated with it in \"%s.go\"", pass.Pkg.Name(), pass.Pkg.Name())
				} else {
					expectedPrefix := fmt.Sprintf("Package %s", pass.Pkg.Name())
					if !strings.HasPrefix(strings.TrimSpace(file.Doc.Text()), expectedPrefix) {
						pass.Reportf(
							file.Package,
							"comment for package \"%s\" should begin with \"%s\"", pass.Pkg.Name(), expectedPrefix)
					}
				}
			}
		}

		// funcStart and funcEnd keep track of the most recently encountered function start and
		// end locations.
		var funcStart, funcEnd int

		var stack []ast.Node
		ast.Inspect(file, func(n ast.Node) bool {
			// Taken from: https://stackoverflow.com/a/66810485
			// Manage the stack. Inspect calls a function like this:
			//   f(node)
			//   for each child {
			//      f(child) // and recursively for child's children
			//   }
			//   f(nil)
			if n == nil {
				// Done with node's children. Pop.
				stack = stack[:len(stack)-1]
			} else {
				// Push the current node for children.
				stack = append(stack, n)
			}

			switch expr := n.(type) {
			case *ast.FuncDecl:
				funcStart = pass.Fset.PositionFor(expr.Pos(), false).Line
				funcEnd = pass.Fset.PositionFor(expr.End(), false).Line

				if !validateFunctions {
					// validateFunctions flag was set to false, ignore all functions.
					return true
				}

				if pass.Pkg.Name() == common.PackageMain && expr.Name.Name == common.FuncMain {
					// Ignore func main in main package.
					return true
				}

				// The reason a 1 is added is to account for single-line functions (edge case).
				// This also doesn't affect non-single line functions, it will just account for
				// the trailing } which is what most people would expect anyways when providing
				// a minimum function length to validate against.
				if (funcEnd - funcStart + 1) >= minFunLen {
					// Run through function declaration validation rules if the minimum function
					// length is met or exceeded.
					validateFuncDecl(pass, expr)
				}
			case *ast.GenDecl:
				if pos := pass.Fset.PositionFor(expr.Pos(), false).Line; pos >= funcStart && pos <= funcEnd {
					// Ignore general declarations that are within a function.
					return true
				}

				// Run through general declaration validation rules, currently these
				// only apply to constants, type, and variable declarations, as you
				// will see if you dig into the proceeding function call.
				validateGenDecl(pass, expr, stack)
			default:
				return true
			}

			return true
		})
	}

	if !allGenerated {
		if !packageHasFileWithSameName {
			pass.Reportf(0, "package \"%s\" has no file with the same name containing package comment", pass.Pkg.Name())
		}
	}

	return nil, nil
}

// validateGenDecl validates an *ast.GenDecl to ensure it is up to doculint standards.
// Currently this function only looks for constants, type, and variable declarations
// then further validates them.
func validateGenDecl(r reporter.Reporter, expr *ast.GenDecl, stack []ast.Node) {
	switch expr.Tok { //nolint:exhaustive // Why: We don't need to take into account anything else.
	case token.CONST:
		if validateConstants {
			// validateConstants flag was set to true, go ahead and validate constants.
			validateGenDeclConstants(r, expr, stack)
		}
	case token.TYPE:
		if validateTypes {
			// validateTypes flag was set to true, go ahead and validate types.
			validateGenDeclTypes(r, expr)
		}
	case token.VAR:
		if validateVariables {
			// validateVariables flag was set to true, go ahead and validate variables.
			validateGenDeclVariables(r, expr)
		}
	}
}

// validateGenDeclConstants validates an *ast.GenDecl that is a constant type. It ensures
// that if it is a constant block that the block itself has a comment, and each constant
// within it also has a comment. If it is a standalone constant it ensures that it has a
// comment associated with it.
func validateGenDeclConstants(r reporter.Reporter, expr *ast.GenDecl, stack []ast.Node) {
	if expr.Lparen.IsValid() {
		// Constant block
		if expr.Doc == nil {
			// We are creating an explicit exception to the commenting rule for constants if a block is what we declare to
			// be "enum-like".  In these cases, the commentingi of the enum type declaration itself should be comprehensively
			// descriptive of the enum's usage, and the values should be well-named to be self-descriptive, though comments
			// may also be added to specific value types if it is helpful -- we simply want to stop mandating them.  The
			// entire block has to match an explicit set of criteria to exclude itself (and all of its members) from the
			// commenting rule:
			// 1. Immediately above the constant block must be a (commented) simple type declaration (i.e. `type MyEnum int`)
			// 2. The constant block must have ONLY simple value declarations in it (comments are also allowed)
			// 3. Every single value declaration in the comment block must explicitly declare the value as typed with the
			//    same type as the type block above the enum.
			//
			// Example of valid block:
			// // MyEnum is a wonderful enum
			// type MyEnum int
			//
			// const (
			//     ValA MyEnum = 1
			//     ValB MyEnum = 4
			//     ValC MyEnum = iota
			// )

			// See if it's enum-like -- are we at the top level of the doc?
			if len(stack) == 2 {
				if parent, worked := stack[0].(*ast.File); worked {
					// Okay, now find our previous sibling to see if it's a type declaration
					for i, decl := range parent.Decls {
						if decl != expr {
							continue
						}

						if i == 0 {
							// No prev sibling somehow, stop checking -- it won't be an enum
							break
						}

						prevSib, is := parent.Decls[i-1].(*ast.GenDecl)
						if !is || prevSib.Tok != token.TYPE || len(prevSib.Specs) != 1 {
							// Previous sibling isn't a simple type declaration that looks like maybe an enum
							break
						}

						typeSpec, is := prevSib.Specs[0].(*ast.TypeSpec)
						if !is {
							// Not a type spec, won't be an enum
							break
						}

						// Looking good!  Now check all the elements inside the const block to make sure they're the same type.
						foundInvalid := false
						for i := range expr.Specs {
							vs, ok := expr.Specs[i].(*ast.ValueSpec)
							if !ok {
								// Const block needs all elements to be value declarations to be an enum -- fail
								foundInvalid = true
								break
							}

							ns, ok := vs.Type.(*ast.Ident)
							if !ok {
								// Declaration needs to have a type associated with it to be an enum -- fail
								foundInvalid = true
								break
							}

							if ns.Name != typeSpec.Name.Name {
								// Declaration's type needs to match the type from the declaration right above the block -- fail
								foundInvalid = true
								break
							}
						}

						if !foundInvalid {
							// All members of the const block were using the same type as the element above it.  Call it an enum,
							// which doesn't need a comment, and move on!
							return
						}

						break
					}
				}
			}

			r.Reportf(expr.Pos(), "constant block has no comment associated with it")
		}
	}

	for i := range expr.Specs {
		vs, ok := expr.Specs[i].(*ast.ValueSpec)
		if ok {
			if len(vs.Names) > 1 {
				var names []string
				for j := range vs.Names {
					names = append(names, fmt.Sprintf("%q", vs.Names[j].Name))
				}

				r.Reportf(vs.Pos(), "constants %s should be separated and each have a comment associated with them", strings.Join(names, ", "))
				continue
			}

			name := vs.Names[0].Name

			doc := vs.Doc
			if !expr.Lparen.IsValid() {
				// If this constant isn't apart of a constant block it's comment is stored in the *ast.GenDecl type.
				doc = expr.Doc
			}

			if doc == nil {
				r.Reportf(vs.Pos(), "constant \"%s\" has no comment associated with it", name)
				continue
			}

			if !strings.HasPrefix(strings.TrimSpace(doc.Text()), name) {
				r.Reportf(vs.Pos(), "comment for constant \"%s\" should begin with \"%s\"", name, name)
			}
		}
	}
}

// validateGenDeclTypes validates an *ast.GenDecl that is a type declaration. It ensures
// that if it is a type declaration block that the block itself has a comment, and each
// type declaration within it also has a comment. If it is a standalone type declaration
// it ensures that it has a comment associated with it.
func validateGenDeclTypes(r reporter.Reporter, expr *ast.GenDecl) {
	if expr.Lparen.IsValid() {
		// Type block
		if expr.Doc == nil {
			r.Reportf(expr.Pos(), "type block has no comment associated with it")
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
				r.Reportf(ts.Pos(), "type \"%s\" has no comment associated with it", ts.Name.Name)
				continue
			}

			if !strings.HasPrefix(strings.TrimSpace(doc.Text()), ts.Name.Name) {
				r.Reportf(ts.Pos(), "comment for type \"%s\" should begin with \"%s\"", ts.Name.Name, ts.Name.Name)
			}
		}
	}
}

// validateGenDeclVariables validates an *ast.GenDecl that is a variable type. It ensures
// that if it is a variable block that the block itself has a comment, and each variable
// within it also has a comment. If it is a standalone variable it ensures that it has a
// comment associated with it.
func validateGenDeclVariables(r reporter.Reporter, expr *ast.GenDecl) {
	if expr.Lparen.IsValid() {
		// Variable block
		if expr.Doc == nil {
			r.Reportf(expr.Pos(), "variable block has no comment associated with it")
		}
	}

	for i := range expr.Specs {
		vs, ok := expr.Specs[i].(*ast.ValueSpec)
		if ok {
			if len(vs.Names) > 1 {
				var names []string
				for j := range vs.Names {
					names = append(names, fmt.Sprintf("%q", vs.Names[j].Name))
				}

				r.Reportf(vs.Pos(), "variables %s should be separated and each have a comment associated with them", strings.Join(names, ", "))
				continue
			}

			name := vs.Names[0].Name
			if name == "_" {
				continue // skip underscore variables.
			}

			doc := vs.Doc
			if !expr.Lparen.IsValid() {
				// If this variable isn't apart of a variable block it's comment is stored in the *ast.GenDecl type.
				doc = expr.Doc
			}

			if doc == nil {
				r.Reportf(vs.Pos(), "variable %q has no comment associated with it", name)
				continue
			}

			if !strings.HasPrefix(strings.TrimSpace(doc.Text()), name) {
				r.Reportf(vs.Pos(), "comment for variable \"%s\" should begin with \"%s\"", name, name)
			}
		}
	}
}

// validateFuncDecl ensures that an *ast.FuncDecl upholds doculint standards by ensuring
// it has a corresponding comment that starts with the name of the function.
func validateFuncDecl(r reporter.Reporter, expr *ast.FuncDecl) {
	if expr.Name.Name == common.FuncInit {
		// Ignore init functions.
		return
	}

	if expr.Doc == nil {
		r.Reportf(expr.Pos(), "function \"%s\" has no comment associated with it", expr.Name.Name)
		return
	}

	// Enforce a space after the function name.
	if !strings.HasPrefix(strings.TrimSpace(expr.Doc.Text()), expr.Name.Name+" ") {
		r.Reportf(expr.Pos(), "comment for function \"%s\" should be a sentence that starts with \"%s \"", expr.Name.Name, expr.Name.Name)
	}
}

// validatePackageName ensures that a given package name follows the conventions that can
// be read about here: https://blog.golang.org/package-names
func validatePackageName(r reporter.Reporter, pos token.Pos, pkg string) {
	if strings.ContainsAny(pkg, "_-") {
		r.Reportf(pos, "package \"%s\" should not contain - or _ in name", pkg)
	}

	if pkg != strings.ToLower(pkg) {
		r.Reportf(pos, "package \"%s\" should be all lowercase", pkg)
	}
}
