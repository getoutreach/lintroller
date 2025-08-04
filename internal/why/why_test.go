// Copyright 2022 Outreach Corporation. Licensed under the Apache License 2.0.

package why

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMatchNoLintWhy(t *testing.T) {
	tt := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "Matches a well-formed nolint",
			text:     "nolint: errcheck // Why: We're unit testing.",
			expected: true,
		},
		{
			name:     "Matches a well-formed nolint with weird spacing",
			text:     "nolint:errcheck //Why: We're unit testing.",
			expected: true,
		},
		{
			name:     "Does not match when there is no space before the Why comment",
			text:     "nolint:errcheck// Why: This is weird.",
			expected: false,
		},
		{
			name:     "Does not match empty why",
			text:     "nolint:errcheck // Why:",
			expected: false,
		},
		{
			name:     "Does not match when there is no Why",
			text:     "nolint:errcheck",
			expected: false,
		},
		{
			name:     "Does not match simple comment",
			text:     "nolint:errcheck // Don't like this.",
			expected: false,
		},
		{
			name:     "Matches when there is a Why on a naked nolint",
			text:     "nolint // Why: No one should do this.",
			expected: true,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, reNoLintWhy.MatchString(test.text), test.expected)
		})
	}
}

func TestMatchNoLintNaked(t *testing.T) {
	tt := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "Matches a simple nolint",
			text:     "nolint",
			expected: true,
		},
		{
			name:     "Matches a simple nolint with a Why",
			text:     "nolint // Why: We're unit testing.",
			expected: true,
		},
		{
			name:     "Does not match a well-formed nolint",
			text:     "nolint: errcheck // Why: We're unit testing.",
			expected: false,
		},
		{
			name:     "Does not match a well-formed nolint with weird spacing",
			text:     "nolint:errcheck //Why: We're unit testing.",
			expected: false,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, reNoLintNaked.MatchString(test.text), test.expected)
		})
	}
}
