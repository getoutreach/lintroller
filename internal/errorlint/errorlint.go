// Copyright 2023 Outreach Corporation. All Rights Reserved.

// Description: See package comment for this one file package.

// Package errorlint contains the necessary logic for the error/log/trace linter. This validates that error
// messages follow Google's go error guidelines
// (https://google.github.io/styleguide/go/decisions.html#error-strings).
// Specifically, this requires that error messages start with a lower-case letter, and do not end in
// punctuation.
//
// This validates the following functions:
// errors.New, errors.WithMessage, errors.WithMessagef, trace.StartCall, trace.StartSpan, log.Warn,
// log.Error, log.Info, errors.Wrap, errors.Wrapf, fmt.Errorf, errors.Errorf
package errorlint

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/getoutreach/lintroller/internal/common"
	"github.com/getoutreach/lintroller/internal/reporter"
	"golang.org/x/tools/go/analysis"
)

// name defines the name for the errorlint linter.
const name = "errorlint"

// doc defines the help text for the error linter.
const doc = `Ensures that each error message starts with a lower-case letter and does not end in punctuation.
// Bad:
err := fmt.Errorf("Something bad happened.")
// Good:
err := fmt.Errorf("something bad happened")`

// Analyzer exports the errorlint analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: name,
	Doc:  doc,
	Run:  errorlint,
}

// NewAnalyzerWithOptions returns the Analyzer package-level variable, with the options
// that would have been defined via flags if this was ran as a vet tool. This is so the
// analyzers can be ran outside of the context of a vet tool and config can be gathered
// from elsewhere.
func NewAnalyzerWithOptions(_warn bool) *analysis.Analyzer {
	warn = _warn
	return &Analyzer
}

// Variable block to keep track of flags whose values are collected at runtime. See the
// init function that immediately proceeds this block to see more.
var (
	// warn denotes whether or not lint reports from this linter will result in warnings or
	// errors.
	warn bool
)

func init() { //nolint:gochecknoinits // Why: This is necessary to grab flags.
	Analyzer.Flags.BoolVar(&warn,
		"warn", false, "controls whether or not reports from this linter will result in errors or warnings")
}

// file represents a file being linted.
type file struct {
	f           *ast.File
	importSpecs map[*ast.File]map[*ast.Ident]string
}

// ImportDecls creates map of import specs for a *ast.File
func (file *file) ImportDecls() {
	for _, is := range file.f.Imports {
		if is.Name == nil {
			continue
		}
		arr := strings.Split(strings.Trim(is.Path.Value, "\""), "/")
		pkgName := arr[len(arr)-1]
		file.importSpecs = make(map[*ast.File]map[*ast.Ident]string)
		file.importSpecs[file.f] = make(map[*ast.Ident]string)
		file.importSpecs[file.f][is.Name] = pkgName
	}
}

func (file *file) walk(fn func(ast.Node) bool) {
	ast.Walk(walker(fn), file.f)
}

// walker adapts a function to satisfy the ast.Visitor interface.
// The function returns whether the walk should proceed into the node's children.
type walker func(ast.Node) bool

func (w walker) Visit(node ast.Node) ast.Visitor {
	if w(node) {
		return w
	}

	return nil
}

// errorlint defines linter for error/trace/log messages
func errorlint(_pass *analysis.Pass) (interface{}, error) {
	// Ignore test packages.
	if common.IsTestPackage(_pass) {
		return nil, nil
	}

	var opts []reporter.PassOption
	if warn {
		opts = append(opts, reporter.Warn())
	}

	// Wrap _pass with reporter.Pass to take nolint directives into account and potentially
	// warn instead of error.
	pass := reporter.NewPass(name, _pass, opts...)

	for _, astFile := range pass.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(astFile) || common.IsTestFile(pass.Pass, astFile) {
			continue
		}
		newFile := &file{f: astFile}
		newFile.ImportDecls()
		newFile.lintMessageStrings(pass)
	}

	return nil, nil
}

// lintMessageStrings examines error/trace/log message strings for capitalization and valid ending
func (file *file) lintMessageStrings(pass *reporter.Pass) {
	file.walk(func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if file.isNotErrorPackage(call.Fun) || len(call.Args) < 1 {
			return true
		}

		msgIndex := 1
		if file.isDotInPkg(call.Fun, "errors", "New") || file.isDotInPkg(call.Fun, "fmt", "Errorf") ||
			file.isDotInPkg(call.Fun, "errors", "Errorf") {
			msgIndex = 0
		}

		msg, ok := call.Args[msgIndex].(*ast.BasicLit)
		if !ok || msg.Kind != token.STRING {
			return true
		}

		msgString, err := strconv.Unquote(msg.Value)
		if err != nil {
			return false
		}
		pkgName := file.getPkgName(call.Fun)
		errormsg := getErrorMessages(msgString)
		for _, msg := range errormsg {
			pass.Reportf(node.Pos(), "%s "+msg, pkgName)
		}
		return true
	})
}

// isNotErrorPackage checks if the ast.Expr package matches the error/fmt/trace/log packages for linter
func (file *file) isNotErrorPackage(expr ast.Expr) bool {
	return !file.isDotInPkg(expr, "errors", "New") && !file.isDotInPkg(expr, "errors", "Wrap") &&
		!file.isDotInPkg(expr, "errors", "Wrapf") && !file.isDotInPkg(expr, "log", "Warn") &&
		!file.isDotInPkg(expr, "log", "Info") && !file.isDotInPkg(expr, "log", "Error") &&
		!file.isDotInPkg(expr, "trace", "StartSpan") && !file.isDotInPkg(expr, "trace", "StartCall") &&
		!file.isDotInPkg(expr, "fmt", "Errorf") && !file.isDotInPkg(expr, "errors", "Errorf") &&
		!file.isDotInPkg(expr, "errors", "WithMessage") && !file.isDotInPkg(expr, "errors", "WithMessagef")
}

// getErrorMessages returns messages based on whether error message is empty, is capitalized, or punctuated
func getErrorMessages(msg string) []string {
	var errormsg []string
	if msg == "" {
		errormsg = append(errormsg, "message should not be empty")
	} else {
		if isUpperCase(msg) {
			errormsg = append(errormsg, "message should not be capitalized")
		}
		if hasPunctuation(msg) {
			errormsg = append(errormsg, "message should not end with punctuation")
		}
	}

	return errormsg
}

// isIdent checks if ident string is equal to ast.Ident name
func (file *file) isIdent(expr ast.Expr, ident, identIdentifier string) (string, bool) {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return "", false
	}
	originalPackage := ""
	if identIdentifier == "pkg" {
		pkgInfo := file.importSpecs[file.f]
		for ident, pkg := range pkgInfo {
			if id.Name == ident.Name {
				originalPackage = pkg
				break
			}
		}
	}
	return originalPackage, id.Name == ident
}

// isDotInPkg checks if pkg.function format is followed
func (file *file) isDotInPkg(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	originalPackage, isPackageMatching := file.isIdent(sel.X, pkg, "pkg")
	_, isFunctionMatching := file.isIdent(sel.Sel, name, "func")
	return (originalPackage != "" || isPackageMatching) && isFunctionMatching
}

// getPkgName returns package name errors/fmt/log/trace
func (file *file) getPkgName(expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	for _, pkg := range []string{"errors", "log", "fmt", "trace"} {
		originalPackage, isPackageMatching := file.isIdent(sel.X, pkg, "pkg")
		if originalPackage == pkg || isPackageMatching {
			return pkg
		}
	}

	return ""
}

// isUpperCase examines error/trace/log strings for capitalization
func isUpperCase(msg string) bool {
	for _, ch := range msg {
		if unicode.IsUpper(ch) {
			return true
		}
	}

	return false
}

// hasPunctuation examines error/trace/log strings for ending in punctuation
func hasPunctuation(msg string) bool {
	last, _ := utf8.DecodeLastRuneInString(msg)
	return unicode.IsPunct(last) || last == '\n'
}
