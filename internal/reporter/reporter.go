// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: See package comment for this one file package.

// Package reporter implements an interface that captures the analysis.Pass.Reportf
// function within it so that it can be wrapped. This provides the ability for consumers
// of the linters exposed in lintroller to explicitly ignore linter errors for the
// linters that report on specific lines by way of nolint directives as well as provide
// the framework within lintroller itself to report lint reports as warnings instead of
// errors for specific linters.
package reporter

import (
	"fmt"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// noLintDirective is the string that is looked for to gather information what to skip when using
// Pass.
const noLintDirective = "nolint:"

// Reporter is a convenience interface that allows Pass to be provided to helper functions
// in linters who need to be able to use Reportf.
type Reporter interface {
	Reportf(pos token.Pos, format string, args ...interface{})
}

// noLint is a struct that depicts a filename/line tandem that shouldn't be linted against
// for the current linter.
//
// The reason we can't use token.Pos directly is because token.Pos will be different integers
// even if the tokens are on the same line due to the fact that it bakes in the column
// number, which we do not care about. Furthermore, the reason we don't use token.Position
// directly is because it contains extraneous information we don't care about, which is just
// a waste of memory.
type noLint struct {
	filename string
	line     int
}

// Matches is a convenience function that matches the receiver with a token.Position.
//
// The reason we match on both noLint.line == position.Line and the line after noLint.line
// (noLint.line+1) is to allow users to specify their nolint directives on the exact same
// line that the linter is complaining about, as well as the one before it.
func (n *noLint) Matches(position token.Position) bool {
	return n.filename == position.Filename && (n.line == position.Line || n.line+1 == position.Line)
}

// Pass is a wrapper around *analysis.Pass that accounts for nolint directives as well as any
// other functionality that it is configured with during initialization with the factory function.
// Please never initialize this type directly, only through NewPass. If you initialize this type
// directly, it will essentially be a no-op wrapper around *analysis.Pass.
type Pass struct {
	noLints []noLint
	linter  string

	*analysis.Pass

	// Functional option supplied configuration
	warn bool
}

// NewPass returns a wrapped version of *analysis.Pass to do reporting that takes account for nolint
// directives as well as any functionality provided by the functional options passed to this.
func NewPass(linter string, pass *analysis.Pass, opts ...PassOption) *Pass {
	p := Pass{
		Pass:   pass,
		linter: linter,
	}

	for i := range opts {
		opts[i](&p)
	}

	for _, file := range p.Files {
		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

				// whySlashesIdx finds the next set of slashes if the nolint directive is in the form
				// of:
				//	nolint: why,doculint // Why: reasoning
				// If these slashes exist we use the index to trim them and all text following it off
				// of the string, effectively producing:
				//	nolint: why,doculint
				if whySlashesIdx := strings.Index(text, "//"); whySlashesIdx != -1 {
					text = strings.TrimSpace(text[:whySlashesIdx])
				}

				if strings.HasPrefix(text, noLintDirective) {
					linters := strings.Split(strings.TrimSpace(strings.TrimPrefix(text, noLintDirective)), ",")

					for i := range linters {
						if linters[i] == linter {
							position := pass.Fset.PositionFor(comment.Pos(), false)
							p.noLints = append(p.noLints, noLint{
								filename: position.Filename,
								line:     position.Line,
							})
							break
						}
					}
				}
			}
		}
	}

	return &p
}

// Reportf is a wrapper around *analysis.Pass.Reportf that respects nolint directives and any other
// functionality provided by the functional options when Pass was formed with its factory function.
func (p *Pass) Reportf(pos token.Pos, format string, args ...interface{}) {
	for i := range p.noLints {
		if p.noLints[i].Matches(p.Pass.Fset.PositionFor(pos, false)) {
			return
		}
	}

	if p.warn {
		fmt.Printf("%s: %s (%s) [WARNING]", p.Fset.PositionFor(pos, false).String(), fmt.Sprintf(format, args...), p.linter)
		return
	}
	p.Pass.Reportf(pos, format+" (%s)", append(args, p.linter)...)
}
