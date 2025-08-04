// Copyright 2022 Outreach Corporation. Licensed under the Apache License 2.0.

// Description: See package description for this one file package.

// Package copyright contains the necessary logic for the copyright linter. The copyright
// linter ensures that a copyright comments exists at the top of each .go file.
package copyright

import (
	"regexp"
	"strings"
	"sync"

	"github.com/getoutreach/lintroller/internal/common"
	"golang.org/x/tools/go/analysis"
)

// Here is an example regular expression that can be used to test this linter:
// ^Copyright 20[2-9][0-9] Outreach Corporation\. All Rights Reserved\.$

// name defines the name of the copyright linter.
const name = "copyright"

// doc defines the help text for the copyright linter.
const doc = `Ensures each .go file has a comment at the top of the file containing the 
copyright string requested via flags.`

// Analyzer exports the copyright analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: name,
	Doc:  doc,
	Run:  copyright,
}

// NewAnalyzerWithOptions returns the Analyzer package-level variable, with the options
// that would have been defined via flags if this was ran as a vet tool. This is so the
// analyzers can be ran outside of the context of a vet tool and config can be gathered
// from elsewhere.
func NewAnalyzerWithOptions(_text, _pattern string) *analysis.Analyzer {
	text = strings.TrimSpace(_text)
	pattern = strings.TrimSpace(_pattern)
	return &Analyzer
}

// Variable block to keep track of flags whose values are collected at runtime. See the
// init function that immediately proceeds this block to see more.
var (
	// text is a variable that gets collected via flags. This variable contains the copyright
	// string as plaintext that is required to be at the top of each .go file.
	text string

	// pattern is a variable that gets collected via flags. This variable contains the copyright
	// string as a regular expression pattern that is required to be at the top of each .go file.
	pattern string
)

// comparer is a convience type used to conditionally compare using either a string or a
// compiled regular expression based off of the existence of the regular expression.
type comparer struct {
	text    string
	pattern *regexp.Regexp

	uniqueCopyrightsInternal map[string]struct{}

	once sync.Once
}

// init gets passed to c.once to initialize the comparer using values pulled from flags.
func (c *comparer) init() {
	// Grab the value passed via flag. Use pattern if it exists, if not, default to text.
	if pattern != "" {
		c.pattern = regexp.MustCompile(pattern)
	} else {
		c.text = text
	}

	// Initialize an empty uniqueCopyrightsInternal map.
	c.uniqueCopyrightsInternal = make(map[string]struct{})
}

// compare uses the struct fields stored in the receiver to conditionally compare a given
// string value.
func (c *comparer) compare(value string) bool {
	c.once.Do(c.init)

	// If the copyright regular expression is non-nil, use that to compare.
	if c.pattern != nil {
		return c.pattern.MatchString(value)
	}

	// Else, use the copyright string to compare.
	if c.text == value {
		return true
	}
	return false
}

// stringMatchType returns the match type the comparer is using as a string (text or pattern),
// for reporting purposes.
func (c *comparer) stringMatchType() string {
	c.once.Do(c.init)

	if c.pattern != nil {
		return "regular expression"
	}
	return "string"
}

// stringMatchType returns the literal string (either text or pattern) it is using to compare,
// for reporting purposes.
func (c *comparer) stringMatchLiteral() string {
	if c.pattern != nil {
		return c.pattern.String()
	}
	return c.text
}

// trackUniqueness takes a copyright string and checks to see if we've already encountered
// it. If it we have, this is a no-op, if we haven't, we mark it as seen for reporting
// purposes at the end of the run.
func (c *comparer) trackUniqueness(copyrightString string) {
	if _, exists := c.uniqueCopyrightsInternal[copyrightString]; !exists {
		c.uniqueCopyrightsInternal[copyrightString] = struct{}{}
	}
}

func init() { //nolint:gochecknoinits // Why: This is necessary to grab flags.
	// Setup flags.
	//nolint:lll // Why: usage long
	Analyzer.Flags.StringVar(&text, "text", "", "the copyright string required at the top of each .go file. if this and pattern are empty the linter is a no-op")
	//nolint:lll // Why: usage long
	Analyzer.Flags.StringVar(&pattern, "pattern", "", "the copyright pattern (as a regular expression) required at the top of each .go file. if this and pattern are empty the linter is a no-op. pattern takes precedence over text if both are supplied")

	// Trim space around the passed in variables just in case.
	text = strings.TrimSpace(text)
	pattern = strings.TrimSpace(pattern)
}

// copyright is the function that gets passed to the Analyzer which runs the actual
// analysis for the copyright linter on a set of files.
func copyright(pass *analysis.Pass) (interface{}, error) { //nolint:funlen // Why: Doesn't make sense to break this function up anymore.
	// Ignore test packages.
	if common.IsTestPackage(pass) {
		return nil, nil
	}

	if text == "" && pattern == "" {
		return nil, nil
	}

	// comparer to use on this pass.
	var c comparer

	for _, file := range pass.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(file) || common.IsTestFile(pass, file) {
			continue
		}

		fp := pass.Fset.PositionFor(file.Package, false).Filename

		// Variable to keep track of whether or not the copyright string was found at the
		// top of the current file.
		var foundCopyright bool

		for _, commentGroup := range file.Comments {
			if pass.Fset.PositionFor(commentGroup.Pos(), false).Line != 1 {
				// The copyright comment needs to be on line 1. Ignore all other comments.
				continue
			}

			// Get the text out of the first line, trimming the // prefix and space before and after
			// that may or may not exist.
			lineOneText := strings.TrimSpace(strings.TrimPrefix(commentGroup.List[0].Text, "//"))

			// Set the value of the foundCopyright to the comparison of this comment's text
			// to the stored copyrightString value or the regular expression compiled from
			// it.
			foundCopyright = c.compare(lineOneText)

			if foundCopyright {
				c.trackUniqueness(lineOneText)
			}

			// We can safely break here because if we got here, regardless on the outcome of
			// the previous statement, we know this is the only comment that matters because
			// it is on line one. This can be verified by the conditional at the top of the
			// current loop.
			break
		}

		if !foundCopyright {
			pass.Reportf(0,
				"file \"%s\" does not contain the required copyright %s [%s] (sans-brackets) as a comment on line 1",
				fp, c.stringMatchType(), c.stringMatchLiteral())
		}
	}

	return nil, nil
}
