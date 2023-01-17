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
const doc = `Ensures that each error message starts with a lower-case letter and does not end in puctuation.
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

// file represents a file being linted.
type file struct {
	f *ast.File
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

	// Wrap _pass with reporter.Pass to take nolint directives into account.
	pass := reporter.NewPass(name, _pass, reporter.Warn())
	for _, astFile := range pass.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(astFile) || common.IsTestFile(pass.Pass, astFile) {
			continue
		}
		newFile := &file{f: astFile}
		lintMessageStrings(newFile, pass)
	}

	return nil, nil
}

// lintMessageStrings examines error/trace/log message strings for capitalization and valid ending
func lintMessageStrings(file *file, pass *reporter.Pass) {
	file.walk(func(node ast.Node) bool {
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

		if msgString == "" {
			return true
		}
		errormsg := getErrorMessage(msgString)
		if errormsg != "" {
			pkgName := getPkgName(call.Fun)
			pass.Reportf(node.Pos(), "%s "+errormsg, pkgName)
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

// getErrorMessage returns message based on whether it has capitalization, punctuation or not
func getErrorMessage(msg string) string {
	isCap, isPunct := isStringFormatted(msg)
	var errormsg string
	switch {
	case isCap && isPunct:
		errormsg = "message should not be capitalized and should not end with punctuation"
	case isCap:
		errormsg = "message should not be capitalized"
	case isPunct:
		errormsg = "message should not end with punctuation"
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
	for _, pkg := range []string{"errors", "log", "fmt", "trace"} {
		if ok && isIdent(sel.X, pkg) {
			return pkg
		}
	}

	return ""
}

// isStringFormatted examines error/trace/log strings for incorrect ending and capitalization
func isStringFormatted(msg string) (isCap, isPunct bool) {
	last, _ := utf8.DecodeLastRuneInString(msg)
	isPunct = unicode.IsPunct(last) || last == '\n'
	for _, ch := range msg {
		if unicode.IsUpper(ch) {
			isCap = true
		}
	}

	return
}
