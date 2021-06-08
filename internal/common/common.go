// Package common contains constants, functions, and types that are used in more than
// one linter.
package common

// PackageMain is a constant denoting the name of the "main" package in a Go program.
const PackageMain = "main"

// FuncMain is a constant denoting the name of the "main" function that exists in
// packageMain of a Go program.
const FuncMain = "main"

// FuncInit is a constant denoting the name of the "init" function that can exist in
// any package of a Go program.
const FuncInit = "init"
