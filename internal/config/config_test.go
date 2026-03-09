// Copyright 2022 Outreach Corporation. Licensed under the Apache License 2.0.

package config

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCompiledExclusionPaths(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for no exclusions", func(t *testing.T) {
		t.Parallel()

		cfg := &Lintroller{}
		compiled, err := cfg.CompiledExclusionPaths()
		assert.NilError(t, err)
		assert.Equal(t, len(compiled), 0)
	})

	t.Run("compiles valid regex patterns", func(t *testing.T) {
		t.Parallel()

		cfg := &Lintroller{Exclusions: Exclusions{Paths: []string{"(^|/)node_modules(/|$)", "third_party$"}}}
		compiled, err := cfg.CompiledExclusionPaths()
		assert.NilError(t, err)
		assert.Equal(t, len(compiled), 2)
		assert.Assert(t, compiled[0].MatchString("foo/node_modules/bar"))
		assert.Assert(t, compiled[1].MatchString("project/third_party"))
	})

	t.Run("returns error for invalid regex patterns", func(t *testing.T) {
		t.Parallel()

		cfg := &Lintroller{Exclusions: Exclusions{Paths: []string{"("}}}
		_, err := cfg.CompiledExclusionPaths()
		assert.ErrorContains(t, err, "invalid exclusions.paths[0] pattern")
	})
}
