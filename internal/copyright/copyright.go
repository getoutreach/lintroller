// Package copyright contains the necessary logic for the copyright linter.
package copyright

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/analysis"
)

// Here is an example regular expression that can be used to test this linter:
// ^Copyright 20[2-9][0-9] Outreach Corporation\. All Rights Reserved\.$

// doc defines the help text for the copyright linter.
const doc = `Ensures each .go file has a comment at the top of the file containing the 
copyright string requested via flags.`

// Analyzer exports the copyright analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: "copyright",
	Doc:  doc,
	Run:  copyright,
}

// Variable block to keep track of flags whose values are collected at runtime. See the
// init function that immediately proceeds this block to see more.
var (
	// copyrightString is a variable that gets collected via flags. This variable contains
	// the copyright string required at the top of each .go file.
	copyrightString string

	// regularExpression is a variable that gets collected via flags. This variable denotes
	// whether or not the copyrightString passed via the copyright flag is a regular
	// expression or just a normal string.
	regularExpression bool
)

func init() { //nolint:gochecknoinits
	Analyzer.Flags.StringVar(&copyrightString, "copyright", "", "the copyright string required at the top of each .go file. if empty this linter is a no-op")
	Analyzer.Flags.BoolVar(&regularExpression, "regex", false, "denotes whether or not the copyright string is a regular expression")
}

// comparer is a convience type used to conditionally compare using either a string or a
// compiled regular expression based off of the existence of the regular expression.
type comparer struct {
	copyrightString string
	copyrightRegex  *regexp.Regexp
}

// compare uses the struct fields stored in the receiver to conditionally compare a given
// string value.
func (c *comparer) compare(value string) bool {
	// If the copyright regular expression is non-nil, use that to
	// compare.
	if c.copyrightRegex != nil {
		return c.copyrightRegex.MatchString(value)
	}

	// Else, use the copyright string to compare.
	if c.copyrightString == value {
		return true
	}
	return false
}

// copyright is the function that gets passed to the Analyzer which runs the actual
// analysis for the copyright linter on a set of files.
func copyright(pass *analysis.Pass) (interface{}, error) { //nolint:funlen
	if copyrightString == "" {
		return nil, nil
	}

	c := comparer{
		copyrightString: copyrightString,
	}

	if regularExpression {
		reCopyright, err := regexp.Compile(c.copyrightString)
		if err != nil {
			return nil, errors.Wrap(err, "compile copyright regular expression")
		}
		c.copyrightRegex = reCopyright
	}

	for _, file := range pass.Files {
		fp := pass.Fset.PositionFor(file.Package, false).Filename

		// Variable to keep track of whether or not the copyright string was found at the
		// top of the current file.
		var foundCopyright bool

		for _, commentGroup := range file.Comments {
			if pass.Fset.PositionFor(commentGroup.Pos(), false).Line != 1 {
				// The copyright comment needs to be on line 1. Ignore all other comments.
				continue
			}

			// Set the value of the foundCopyright to the comparison of this comment's text
			// to the stored copyrightString value or the regular expression compiled from
			// it.
			foundCopyright = c.compare(strings.TrimSpace(commentGroup.Text()))

			// We can safely break here because if we got here, regardless on the outcome of
			// the previous statement, we know this is the only comment that matters because
			// it is on line one. This can be verified by the conditional at the top of the
			// current loop.
			break
		}

		if !foundCopyright {
			matchType := "string"
			if regularExpression {
				matchType = "regular expression"
			}

			pass.Reportf(0, "file \"%s\" does not contain the required copyright %s [%s] (sans-brackets) as a comment on line 1", fp, matchType, copyrightString)
		}
	}

	return nil, nil
}
