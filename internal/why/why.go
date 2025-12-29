// Copyright 2022 Outreach Corporation. Licensed under the Apache License 2.0.

// Description: See package comment for this one file package.

// Package why contains the necessary logic for the why linter. The why linter ensures that
// all nolint directives contain a followup // Why: ... statement after the nolint as well
// as no naked nolint directives exist (//nolint as opposed to //nolint:specificLinter).
package why

import (
	"regexp"
	"strings"

	"github.com/getoutreach/lintroller/internal/common"
	"github.com/getoutreach/lintroller/internal/reporter"
	"golang.org/x/tools/go/analysis"
)

// name defines the name for the why linter.
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

// NewAnalyzerWithOptions returns the Analyzer package-level variable, with the options
// that would have been defined via flags if this was ran as a vet tool. This is so the
// analyzers can be ran outside of the context of a vet tool and config can be gathered
// from elsewhere.
func NewAnalyzerWithOptions(_warn bool) *analysis.Analyzer {
	warn = _warn
	return &Analyzer
}

// Variable block to keep track of flags whose values are collected at runtime. See the
// init function that immediately proceeds this block to see more.
var (
	// warn denotes whether or not lint reports from this linter will result in warnings or
	// errors.
	warn bool
)

func init() { //nolint:gochecknoinits // Why: This is necessary to grab flags.
	Analyzer.Flags.BoolVar(&warn,
		"warn", false, "controls whether or not reports from this linter will result in errors or warnings")
}

// whyPattern is a regular expression fragment that matches just a "Why"
// comment.
const whyPattern = `//\s?Why:.+`

// reNoLintWhy is the regular expression that every nolint comment must match, which
// ensures it contains a // Why: <reason> immediately proceeding the nolint directive.
var reNoLintWhy = regexp.MustCompile(`^nolint(?::\s?[\w\-,]+)?\s+` + whyPattern + `$`)

// reNoLintNaked is the regular expression that every nolint comment is checked against
// to ensure no naked nolint directives exist. This matches `nolint` comments
// without a directive and with an optional Why comment.
var reNoLintNaked = regexp.MustCompile(`^nolint\s*(?:` + whyPattern + `)?$`)

// why is the function that gets passed to the Analyzer which runs the actual analysis
// for the why linter on a set of files.
func why(_pass *analysis.Pass) (interface{}, error) {
	// Ignore test packages.
	if common.IsTestPackage(_pass) {
		return nil, nil
	}

	var opts []reporter.PassOption
	if warn {
		opts = append(opts, reporter.Warn())
	}

	// Wrap _pass with reporter.Pass to take nolint directives into account and potentially
	// warn instead of error.
	pass := reporter.NewPass(name, _pass, opts...)

	for _, file := range pass.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(file) || common.IsTestFile(pass.Pass, file) {
			continue
		}

		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

				if strings.HasPrefix(text, "nolint") {
					if reNoLintNaked.MatchString(text) {
						pass.Reportf(comment.Pos(), "nolint directive must contain the specific linters it is nolinting against")
					}

					if !reNoLintWhy.MatchString(text) {
						pass.Reportf(comment.Pos(), "nolint comment must immediately be followed by // Why: <reason> on the same line.")
					}
				}
			}
		}
	}

	return nil, nil
}
