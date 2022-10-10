// Copyright 2022 Outreach Corporation. All Rights Reserved.

package todo

import (
	"go/ast"
	"testing"

	"gotest.tools/v3/assert"
)

func TestMatchTodo(t *testing.T) {
	tt := []struct {
		name        string
		commentText string
		expected    bool
	}{
		{
			name:        "Passes one-line comment",
			commentText: "// One-liner",
			expected:    true,
		},
		{
			name:        "Passes multi-line comment",
			commentText: "/* This is \nmultiple\n lines*/",
			expected:    true,
		},
		{
			name:        "Passes TODO with only username",
			commentText: "// TODO(jkinkead): Fix this.",
			expected:    true,
		},
		{
			name:        "Passes TODO with only ticket",
			commentText: "// TODO[JT-101]: Fix this.",
			expected:    true,
		},
		{
			name:        "Passes TODO with both username and ticket",
			commentText: "// TODO(jkinkead)[JT-101]: Fix this.",
			expected:    true,
		},
		{
			name:        "Allows TODO in multi-line comment",
			commentText: "/* Here we're doing stuff badly.\n\tTODO(jkinkead): Fix this.\n*/",
			expected:    true,
		},
		{
			name:        "Forbids TODO with no colon",
			commentText: "// TODO(jkinkead)[JT-101] Fix this.",
			expected:    false,
		},
		{
			name:        "Forbids TODO with only colon",
			commentText: "// TODO: Fix this.",
			expected:    false,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			comment := &ast.Comment{Text: test.commentText}
			assert.Equal(t, matchTodo(comment), test.expected)
		})
	}
}
