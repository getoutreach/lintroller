// Copyright 2022 Outreach Corporation. Licensed under the Apache License 2.0.

package copyright

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestCopyright runs the copyright analyzer against the testdata packages. analysistest
// matches each diagnostic to the `// want` comment on the line it was reported at, so
// this also guards against the analyzer reporting at token.NoPos, which would leave the
// diagnostic without a file, line, and column for downstream tooling to parse.
func TestCopyright(t *testing.T) {
	const pattern = `^Copyright 20[0-9]{2} Outreach Corporation\. All Rights Reserved\.$`

	analysistest.Run(t, analysistest.TestData(), NewAnalyzerWithOptions("", pattern), "missing", "valid")
}
