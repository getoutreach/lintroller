// Copyright 2022 Outreach Corporation. Licensed under the Apache License 2.0.

package main

import (
	"regexp"
	"testing"

	"gotest.tools/v3/assert"
)

func TestMatchesAnyExclusion(t *testing.T) {
	t.Parallel()

	t.Run("matches normalized windows paths", func(t *testing.T) {
		t.Parallel()

		exclusions := []*regexp.Regexp{regexp.MustCompile(`(^|/)node_modules(/|$)`)}
		assert.Assert(t, matchesAnyExclusion(`foo\\node_modules\\bar\\file.go`, exclusions))
	})

	t.Run("matches unix style paths", func(t *testing.T) {
		t.Parallel()

		exclusions := []*regexp.Regexp{regexp.MustCompile(`(^|/)vendor/some-third-party(/|$)`)}
		assert.Assert(t, matchesAnyExclusion("project/vendor/some-third-party/file.go", exclusions))
	})

	t.Run("does not match unrelated paths", func(t *testing.T) {
		t.Parallel()

		exclusions := []*regexp.Regexp{regexp.MustCompile(`(^|/)node_modules(/|$)`)}
		assert.Assert(t, !matchesAnyExclusion("project/internal/config/config.go", exclusions))
	})
}
