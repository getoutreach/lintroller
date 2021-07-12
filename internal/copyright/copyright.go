package copyright

import (
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer exports the copyright analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: "copyright",
	Doc:  "ensures each .go file has a comment at the top of the file containing the copyright string requested via flags",
	Run:  copyright,
}

// Variable block to keep track of flags whose values are collected at runtime. See the
// init function that immediately proceeds this block to see more.
var (
	// copyrightString is a variable that gets collected via flags. This variable contains
	// the copyright string required at the top of each .go file.
	copyrightString string
)

func init() { //nolint:gochecknoinits
	Analyzer.Flags.StringVar(&copyrightString, "copyright", "", "the copyright string required at the top of each .go file, if empty this linter is a no-op")
}

// copyright is the function that gets passed to the Analyzer which runs the actual
// analysis for the copyright linter on a set of files.
func copyright(pass *analysis.Pass) (interface{}, error) { //nolint:funlen
	if copyrightString == "" {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Variable to keep track of whether or not the copyright string was found at the
		// top of the current file.
		var foundCopyright bool

		for _, commentGroup := range file.Comments {
			if pass.Fset.PositionFor(commentGroup.Pos(), false).Line != 1 {
				// The copyright comment needs to be on line 1. Ignore all other comments.
				continue
			}

			if copyrightString == strings.TrimSpace(commentGroup.Text()) {
				foundCopyright = true
			}

			// We can safely break here because if we got here, regardless on the outcome of
			// the previous conditional, we know this is the only comment that matters because
			// it is on line one. This can be verified by the conditional at the top of the
			// current loop.
			break
		}

		if !foundCopyright {
			fp := pass.Fset.PositionFor(file.Package, false).Filename
			pass.Reportf(0, "file \"%s\" does not contain the required copyright string \"%s\" as a comment on line 1", fp, copyrightString)
		}
	}

	return nil, nil
}
