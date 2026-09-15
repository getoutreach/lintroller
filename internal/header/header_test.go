// Copyright 2022 Outreach Corporation. Licensed under the Apache License 2.0.

package header

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestHeader runs the header analyzer against the testdata packages. analysistest
// matches each diagnostic to the `// want` comment on the line it was reported at, so
// this also guards against the analyzer reporting at token.NoPos, which would leave the
// diagnostic without a file, line, and column for downstream tooling to parse.
func TestHeader(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzerWithOptions("Description"), "missing", "valid")
}
