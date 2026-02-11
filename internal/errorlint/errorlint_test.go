// Copyright 2023 Outreach Corporation. All Rights Reserved.

package errorlint

import (
	"go/ast"
	"go/token"
	"testing"

	"gotest.tools/v3/assert"
)

func TestValidateErrorMessage(t *testing.T) {
	msgTest := []struct {
		name string
		msg  string
		// The expected format string for the Reportf function argument.
		expectedError []string
	}{
		{
			name:          "empty message string",
			msg:           "",
			expectedError: []string{"message should not be empty"},
		},
		{
			name:          "valid message string",
			msg:           "something good happened",
			expectedError: nil,
		},
		{
			name:          "uppercase message string",
			msg:           "Something bad happened",
			expectedError: []string{"message should not be capitalized"},
		},
		{
			name:          "punctuated message string",
			msg:           "something bad happened.",
			expectedError: []string{"message should not end with punctuation"},
		},
		{
			name:          "uppercase and punctuated message string",
			msg:           "Something bad happened.",
			expectedError: []string{"message should not be capitalized", "message should not end with punctuation"},
		},
	}

	for _, test := range msgTest {
		t.Run(test.name, func(t *testing.T) {
			assert.DeepEqual(t, getErrorMessages(test.msg), test.expectedError)
		})
	}
}

type MockReporter struct {
	lastFormat string
}

func (r *MockReporter) Reportf(pos token.Pos, format string, args ...interface{}) {
	r.lastFormat = format
}

func TestImportDecl(t *testing.T) {
	msgTest := []struct {
		name string
		path string
		// The expected format string for the Reportf function argument.
		expectedOriginalPackage string
	}{
		{
			name:                    "_errors",
			path:                    "github.com/pkg/errors",
			expectedOriginalPackage: "errors",
		},
		{
			name:                    "log",
			path:                    "github.com/getoutreach/gobox/pkg/log",
			expectedOriginalPackage: "log",
		},
	}

	for _, test := range msgTest {
		t.Run(test.name, func(t *testing.T) {
			testFile := &ast.File{
				Imports: []*ast.ImportSpec{{
					Name: &ast.Ident{Name: test.name},
					Path: &ast.BasicLit{Value: test.path},
				},
				},
			}
			file := &file{f: testFile}
			file.ImportDecls()
			assert.Equal(t, file.importSpecs[file.f][file.f.Imports[0].Name], test.expectedOriginalPackage)
		})
	}
}
