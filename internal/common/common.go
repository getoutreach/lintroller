// Package common contains constants, functions, and types that are used in more than
// one linter.
package common

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// PackageMain is a constant denoting the name of the "main" package in a Go program.
const PackageMain = "main"

// DocFilenameWithoutPath is the name of a file that can potentially hold a package
// document as opposed to a file of the same name as the package.
const DocFilenameWithoutPath = "doc"

// FuncMain is a constant denoting the name of the "main" function that exists in
// packageMain of a Go program.
const FuncMain = "main"

// FuncInit is a constant denoting the name of the "init" function that can exist in
// any package of a Go program.
const FuncInit = "init"

// IsGenerated determines if the given file is a generated file.
//
// Developer note: Periodically check whether or not this functionality is exposed in
// a stdlib package by looking at the status of this accepted proposal:
//	https://github.com/golang/go/issues/28089
// Whenever it's added in a stdlib package, use that instead of this because it's
// likely going to have been implemented in a manner that is a lot smarter than
// this implementation.
func IsGenerated(file *ast.File) bool {
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.Contains(strings.ToLower(comment.Text), "code generated") {
				return true
			}
		}
	}

	return false
}

// IsTestPackage determines whether or not the package for the current pass is a test
// package. The analysis.Analyzer is already smart enough to ignore "*_test.go" files,
// but sometimes there are explicit packages only meant to be used in tests. These are,
// or at least should be, suffixed with "test" (usually "_test").
func IsTestPackage(pass *analysis.Pass) bool {
	return strings.HasSuffix(pass.Pkg.Name(), "test")
}
