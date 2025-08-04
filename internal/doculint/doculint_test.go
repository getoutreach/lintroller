// Copyright 2022 Outreach Corporation. Licensed under the Apache License 2.0.

package doculint

import (
	"go/ast"
	"go/token"
	"testing"

	"gotest.tools/v3/assert"
)

type MockReporter struct {
	lastFormat string
}

func (r *MockReporter) Reportf(pos token.Pos, format string, args ...interface{}) {
	r.lastFormat = format
}

func TestValidateFunDecl(t *testing.T) {
	tt := []struct {
		name     string
		funcName string
		funcDoc  string
		// The expected format string for the Reportf function argument.
		expectedFormat string
	}{
		{
			name:           "Ignores init function",
			funcName:       "init",
			funcDoc:        "",
			expectedFormat: "",
		},
		{
			name:           "Produces an error when no doc exists",
			funcName:       "foo",
			funcDoc:        "",
			expectedFormat: "function \"%s\" has no comment associated with it",
		},
		{
			name:           "Produces an error when the doc doesn't start with the function name",
			funcName:       "foo",
			funcDoc:        "This function is foo.",
			expectedFormat: "comment for function \"%s\" should be a sentence that starts with \"%s \"",
		},
		{
			name:           "Produces an error when the doc is malformed",
			funcName:       "foo",
			funcDoc:        "foo: Does a foo thing.",
			expectedFormat: "comment for function \"%s\" should be a sentence that starts with \"%s \"",
		},
		{
			name:           "Produces an error when the doc has a bad function name",
			funcName:       "foo",
			funcDoc:        "fooBar: Does a foobar thing.",
			expectedFormat: "comment for function \"%s\" should be a sentence that starts with \"%s \"",
		},
		{
			name:           "Allows good comments",
			funcName:       "foo",
			funcDoc:        "foo sure is a function.",
			expectedFormat: "",
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			reporter := &MockReporter{}
			funcDecl := &ast.FuncDecl{
				Name: &ast.Ident{Name: test.funcName},
				Type: &ast.FuncType{},
			}
			if test.funcDoc != "" {
				funcDecl.Doc = &ast.CommentGroup{List: []*ast.Comment{{Text: test.funcDoc}}}
			}
			validateFuncDecl(reporter, funcDecl)
			assert.Equal(t, reporter.lastFormat, test.expectedFormat)
		})
	}
}
