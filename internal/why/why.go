// Package why contains the necessary logic for the why linter.
package why

import (
	"regexp"
	"strings"

	"github.com/getoutreach/lintroller/internal/common"
	"github.com/getoutreach/lintroller/internal/nolint"
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
var reNoLintWhy = regexp.MustCompile(`^nolint:\s?[\w\-,]+\s?\/\/\s?Why:\s?.+$`) //nolint:gocritic // Why: It is suggesting bad syntax by not escaping each of the forward slashes.

// reNoLintNaked is the regular expression that every nolint comment is checked against
// to ensure no naked nolint directives exist.
//
// For examples, see https://regex101.com/r/XJY8md/1
var reNoLintNaked = regexp.MustCompile(`^nolint(\s*\/\/\s*Why:.*)*$`) //nolint:gocritic // Why: It is suggesting bad syntax by not escaping each of the forward slashes.

// why is the function that gets passed to the Analyzer which runs the actual analysis
// for the why linter on a set of files.
func why(pass *analysis.Pass) (interface{}, error) {
	// Ignore test packages.
	if common.IsTestPackage(pass) {
		return nil, nil
	}

	// Wrap pass with nolint.Pass to take nolint directives into account.
	passWithNoLint := nolint.PassWithNoLint(name, pass)

	for _, file := range passWithNoLint.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(file) || common.IsTestFile(passWithNoLint.Pass, file) {
			continue
		}

		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

				if strings.HasPrefix(text, "nolint") {
					if reNoLintNaked.MatchString(text) {
						passWithNoLint.Reportf(comment.Pos(), "nolint directive must contain the specific linters it is nolinting against")

						// Continue, if it doesn't contain the why it'll catch it after the error is resolved here.
						// We'd have to modify the regex to be more confusing than it already is to account for it.
						continue
					}

					if !reNoLintWhy.MatchString(text) {
						passWithNoLint.Reportf(comment.Pos(), "nolint comment must immediately be followed by // Why: <reason> on the same line.")
					}
				}
			}
		}
	}

	return nil, nil
}
