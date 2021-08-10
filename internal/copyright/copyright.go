// Package copyright contains the necessary logic for the copyright linter.
package copyright

import (
	"regexp"
	"strings"
	"sync"

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

// comparer is a convience type used to conditionally compare using either a string or a
// compiled regular expression based off of the existence of the regular expression.
type comparer struct {
	copyrightString string
	copyrightRegex  *regexp.Regexp

	uniqueCopyrightsInternal map[string]struct{}

	once sync.Once
}

// init gets passed to c.once to initialize the comparer using values pulled from flags.
func (c *comparer) init() {
	// Grab the value passed via flag.
	c.copyrightString = copyrightString

	// Initialize an empty uniqueCopyrightsInternal map.
	c.uniqueCopyrightsInternal = make(map[string]struct{})

	// If the value passed via flag was denoted to be a regular expression, compile it.
	if regularExpression {
		c.copyrightRegex = regexp.MustCompile(c.copyrightString)
	}
}

// compare uses the struct fields stored in the receiver to conditionally compare a given
// string value.
func (c *comparer) compare(value string) bool {
	c.once.Do(c.init)

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

// stringMatchType returns the match type the comparer is using as a string, used for
// reporting.
func (c *comparer) stringMatchType() string {
	c.once.Do(c.init)

	if c.copyrightRegex != nil {
		return "regular expression"
	}
	return "string"
}

// trackUniqueness takes a copyright string and checks to see if we've already encountered
// it. If it we have, this is a no-op, if we haven't, we mark it as seen for reporting
// purposes at the end of the run.
func (c *comparer) trackUniqueness(copyrightString string) {
	if _, exists := c.uniqueCopyrightsInternal[copyrightString]; !exists {
		c.uniqueCopyrightsInternal[copyrightString] = struct{}{}
	}
}

// uniqueCopyrights returns a slice of all of the unqiue copyright strings found.
func (c *comparer) uniqueCopyrights() []string {
	var unique []string
	for copyrightString, _ := range c.uniqueCopyrightsInternal {
		unique = append(unique, copyrightString)
	}
	return unique
}

func init() { //nolint:gochecknoinits
	// Setup flags.
	Analyzer.Flags.StringVar(&copyrightString, "string", "", "the copyright string required at the top of each .go file. if empty this linter is a no-op")
	Analyzer.Flags.BoolVar(&regularExpression, "regex", false, "denotes whether or not the copyright string was given as a regular expression")
}

// copyright is the function that gets passed to the Analyzer which runs the actual
// analysis for the copyright linter on a set of files.
func copyright(pass *analysis.Pass) (interface{}, error) { //nolint:funlen
	if copyrightString == "" {
		return nil, nil
	}

	// comparer to use on this pass.
	var c comparer

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

			lineOneText := strings.TrimSpace(commentGroup.Text())

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
			pass.Reportf(0, "file \"%s\" does not contain the required copyright %s [%s] (sans-brackets) as a comment on line 1", fp, c.stringMatchType(), copyrightString)
		}
	}

	if uniqueCopyrights := c.uniqueCopyrights(); len(uniqueCopyrights) > 1 {
		pass.Reportf(0, "found multiple unique versions of copyright strings, consider consolidating to one version: %+v", uniqueCopyrights)
	}

	return nil, nil
}
