// Copyright 2023 Outreach Corporation. All Rights Reserved.

// Description: See package comment for this one file package.

// Package errorlint contains the necessary logic for the error/log/trace linter.
package errorlint

import (
	"go/ast"
	"go/token"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/getoutreach/lintroller/internal/common"
	"github.com/getoutreach/lintroller/internal/nolint"
	"golang.org/x/tools/go/analysis"
)

// name defines the name for the errorlint linter.
const name = "errorlint"

// doc defines the help text for the error linter.
const doc = `Ensures that each static message in error/trace/log/ is lowercase.

A valid example is the following:

	func foo() { errors.New("org not found in launchdarkly rule")}`

// Analyzer exports the errorlint analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: name,
	Doc:  doc,
	Run:  errorlint,
}

// errorlint defines linter for error/trace/log messages
func errorlint(pass *analysis.Pass) (interface{}, error) {
	// Ignore test packages.
	if common.IsTestPackage(pass) {
		return nil, nil
	}

	// Wrap pass with nolint.Pass to take nolint directives into account.
	passWithNoLint := nolint.PassWithNoLint(name, pass)
	for _, file := range passWithNoLint.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(file) || common.IsTestFile(passWithNoLint.Pass, file) {
			continue
		}

		for expr := range pass.TypesInfo.Types {
			pkgName, isCleanString := lintMessageStrings(expr)
			if !isCleanString {
				passWithNoLint.Reportf(expr.Pos(), "%s message should be lowercase", pkgName)
			}
		}
	}

	return nil, nil
}

// lintMessageStrings examines error/trace/log strings for capitalization and valid ending
func lintMessageStrings(expr ast.Expr) (string, bool) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return "", true
	}

	if !isPkgDot(call.Fun, "errors", "New") && !isPkgDot(call.Fun, "errors", "Wrap") &&
		!isPkgDot(call.Fun, "errors", "Wrapf") && !isPkgDot(call.Fun, "log", "Warn") &&
		!isPkgDot(call.Fun, "log", "Info") && !isPkgDot(call.Fun, "log", "Error") &&
		!isPkgDot(call.Fun, "trace", "StartSpan") && !isPkgDot(call.Fun, "trace", "StartCall") {
		return "", true
	}

	if len(call.Args) < 1 {
		return "", true
	}

	msgIndex := 1
	if isPkgDot(call.Fun, "errors", "New") {
		msgIndex = 0
	}

	msg, ok := call.Args[msgIndex].(*ast.BasicLit)
	if !ok || msg.Kind != token.STRING {
		return "", true
	}

	msgString, err := strconv.Unquote(msg.Value)
	if err != nil {
		return "", false
	}

	if msgString == "" {
		return "", true
	}
	isClean := isStringFormatted(msgString)
	return getPkgName(call.Fun), isClean
}

// isIdent checks if ident string is equal to ast.Ident name
func isIdent(expr ast.Expr, ident string) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == ident
}

// isPkgDot checks if pkg.function format is followed
func isPkgDot(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	return ok && isIdent(sel.X, pkg) && isIdent(sel.Sel, name)
}

// getPkgName returns package name
func getPkgName(expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	for _, pkg := range []string{"errors", "log", "trace"} {
		if ok && isIdent(sel.X, pkg) {
			return pkg
		}
	}

	return ""
}

// isStringFormatted examines error/trace/log strings for incorrect ending and capitalization
func isStringFormatted(msg string) bool {
	first, firstN := utf8.DecodeRuneInString(msg)
	last, _ := utf8.DecodeLastRuneInString(msg)
	if last == '.' || last == ':' || last == '!' || last == '\n' {
		return false
	}

	if unicode.IsUpper(first) {
		if len(msg) <= firstN {
			return false
		}

		if second, _ := utf8.DecodeRuneInString(msg[firstN:]); !unicode.IsUpper(second) {
			return false
		}
	}

	return true
}
