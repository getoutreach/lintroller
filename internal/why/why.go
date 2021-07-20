// Package why contains the necessary logic for the why linter.
package why

import (
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// doc defines the name for the why linter.
const name = "why"

// doc defines the help text for the why linter.
const doc = `Ensures that each nolint comment also has a // Why: <reason> immediately following it.

A valid example is the following:

	func foo() { //nolint:doculint // Why: This comment doesn't need a function for some reason.`

// Analyzer exports the why analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: name,
	Doc:  doc,
	Run:  why,
}

// reNoLintWhy is the regular expression that every nolint comment must match, which
// ensures it contains a // Why: <reason> immediately proceeding the nolint directive.
//
// For examples, see https://regex101.com/r/I41sfC/1
var reNoLintWhy = regexp.MustCompile(`^nolint:\s?[\w\-,]+\s?\/\/\s?Why:\s?.+$`) //nolint:regexpSimplify,gocritic // Why: It is suggesting bad syntax by not escaping each of the forward slashes.

// why is the function that gets passed to the Analyzer which runs the actual analysis
// for the why linter on a set of files.
func why(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

				if strings.HasPrefix(text, "nolint") {
					if !reNoLintWhy.MatchString(text) {
						pass.Reportf(comment.Pos(), "nolint comment must immediately be followed by // Why: <reason> on the same line.")
					}
				}
			}
		}
	}

	return nil, nil
}
