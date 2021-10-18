// Package nolint implements the ability for consumers of the linters exposed in lintroller
// to explicitly ignore linter errors for the linters that report on specific lines.
package nolint

import (
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// directive is the string that is looked for to gather information what to skip when using
// Pass.
const directive = "nolint:"

// Reporter is a convienance interface that allows Pass to be provided to helper functions
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

// Matches is a convienance function that matches the receiver with a token.Position.
//
// The reason we match on both noLint.line == position.Line and the line after noLint.line
// (noLint.line+1) is to allow users to specify their nolint directives on the exact same
// line that the linter is complaining about, as well as the one before it.
func (n *noLint) Matches(position token.Position) bool {
	return n.filename == position.Filename && (n.line == position.Line || n.line+1 == position.Line)
}

// Pass is a wrapper around *analysis.Pass that accounts for nolint directives. Please never
// initialize this type directly, only through PassWithNoLint. If you initialize this type
// directly, it will essentially be a no-op wrapper around *analysis.Pass.
type Pass struct {
	noLints []noLint
	linter  string

	*analysis.Pass
}

// PassWithNoLint returns a wrapped version of *analysis.Pass as Pass with nolint directives
// identified for the provided linter name.
func PassWithNoLint(linter string, pass *analysis.Pass) *Pass {
	p := Pass{
		Pass:   pass,
		linter: linter,
	}

	for _, file := range p.Files {
		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

				// whySlashesIdx finds the next set of slashes if the nolint directive is in the form
				// of:
				//	nolint: linter1,linter2 // Why: reasoning
				// If these slashes exist we use the index to trim them and all text following it off
				// of the string, effectively producing:
				//	nolint: linter1,linter2
				if whySlashesIdx := strings.Index(text, "//"); whySlashesIdx != -1 {
					text = strings.TrimSpace(text[:whySlashesIdx])
				}

				if strings.HasPrefix(text, directive) {
					linters := strings.Split(strings.TrimSpace(strings.TrimPrefix(text, directive)), ",")

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

// Reportf is a wrapper around *analysis.Pass.Reportf that respects nolint directives stored
// in *Pass.noLint, which are gathered in PassWithNoLint.
func (p *Pass) Reportf(pos token.Pos, format string, args ...interface{}) {
	for i := range p.noLints {
		if p.noLints[i].Matches(p.Pass.Fset.PositionFor(pos, false)) {
			return
		}
	}

	p.Pass.Reportf(pos, format+" (%s)", append(args, p.linter)...)
}
