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

	for _, file := range pass.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(file) || common.IsTestFile(pass.Pass, file) {
			continue
		}
		lintMessageStrings(file, pass)
	}

	return nil, nil
}

// lintMessageStrings examines error/trace/log message strings for capitalization and valid ending
func lintMessageStrings(file *ast.File, pass *reporter.Pass) {
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isNotErrorPackage(call.Fun) || len(call.Args) < 1 {
			return true
		}

		msgIndex := 1
		if isDotInPkg(call.Fun, "errors", "New") || isDotInPkg(call.Fun, "fmt", "Errorf") ||
			isDotInPkg(call.Fun, "errors", "Errorf") {
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
		pkgName := getPkgName(call.Fun)
		errormsg := getErrorMessages(msgString)
		for _, msg := range errormsg {
			pass.Reportf(node.Pos(), "%s "+msg, pkgName)
		}
		return true
	})
}

// isNotErrorPackage checks if the ast.Expr package matches the error/fmt/trace/log packages for linter
func isNotErrorPackage(expr ast.Expr) bool {
	return !isDotInPkg(expr, "errors", "New") && !isDotInPkg(expr, "errors", "Wrap") &&
		!isDotInPkg(expr, "errors", "Wrapf") && !isDotInPkg(expr, "log", "Warn") &&
		!isDotInPkg(expr, "log", "Info") && !isDotInPkg(expr, "log", "Error") &&
		!isDotInPkg(expr, "trace", "StartSpan") && !isDotInPkg(expr, "trace", "StartCall") &&
		!isDotInPkg(expr, "fmt", "Errorf") && !isDotInPkg(expr, "errors", "Errorf") &&
		!isDotInPkg(expr, "errors", "WithMessage") && !isDotInPkg(expr, "errors", "WithMessagef")
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
func isIdent(expr ast.Expr, ident string) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == ident
}

// hasDotInPkg checks if pkg.function format is followed
func isDotInPkg(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	return ok && isIdent(sel.X, pkg) && isIdent(sel.Sel, name)
}

// getPkgName returns package name errors/fmt/log/trace
func getPkgName(expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	for _, pkg := range []string{"errors", "log", "fmt", "trace"} {
		if isIdent(sel.X, pkg) {
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
	return last == '.' || last == ':' || last == '!' || last == '\n'
}
