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
			pkgName, isCleanString := lintMessagerStrings(expr)
			if !isCleanString {
				passWithNoLint.Reportf(expr.Pos(), "%s message should be lowercase", pkgName)
			}
		}
	}

	return nil, nil
}

// lintMessagerStrings examines error/trace/log strings for capitalization and valid ending
func lintMessagerStrings(f ast.Expr) (string, bool) {
	ce, ok := f.(*ast.CallExpr)
	if !ok {
		return "", true
	}
	if !isPkgDot(ce.Fun, "errors", "New") && !isPkgDot(ce.Fun, "fmt", "Errorf") &&
		!isPkgDot(ce.Fun, "errors", "Wrap") && !isPkgDot(ce.Fun, "errors", "Wrapf") &&
		!isPkgDot(ce.Fun, "log", "Info") && !isPkgDot(ce.Fun, "log", "Error") &&
		!isPkgDot(ce.Fun, "log", "Warn") && !isPkgDot(ce.Fun, "trace", "StartCall") &&
		!isPkgDot(ce.Fun, "trace", "StartSpan") {
		return "", true
	}
	if len(ce.Args) < 1 {
		return "", true
	}
	msgIndex := 1
	if isPkgDot(ce.Fun, "errors", "New") || isPkgDot(ce.Fun, "fmt", "Errorf") {
		msgIndex = 0
	}
	str, ok := ce.Args[msgIndex].(*ast.BasicLit)
	if !ok || str.Kind != token.STRING {
		return "", true
	}
	s, err := strconv.Unquote(str.Value)
	if err != nil {
		return "", false
	}
	if s == "" {
		return "", true
	}
	clean := isStringFormatted(s)
	return getPkgName(ce.Fun), clean
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
	for _, pkg := range []string{"errors", "fmt", "log", "trace"} {
		if ok && isIdent(sel.X, pkg) {
			return pkg
		}
	}
	return ""
}

// isStringFormatted examines error/trace/log strings for incorrect ending and capitalization
func isStringFormatted(s string) bool {
	first, firstN := utf8.DecodeRuneInString(s)
	last, _ := utf8.DecodeLastRuneInString(s)
	if last == '.' || last == ':' || last == '!' || last == '\n' {
		return false
	}
	if unicode.IsUpper(first) {
		if len(s) <= firstN {
			return false
		}
		if second, _ := utf8.DecodeRuneInString(s[firstN:]); !unicode.IsUpper(second) {
			return false
		}
	}
	return true
}
