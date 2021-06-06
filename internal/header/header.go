package header

import (
	"github.com/getoutreach/lintroller/internal/common"
	"golang.org/x/tools/go/analysis"
)

// Analyzer exports the doculint analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: "header",
	Doc:  "ensures each .go file has a header that explains the purpose of the file and any notable design decisions",
	Run:  header,
}

// header is the function that gets passed to the Analyzer which runs the actual
// analysis for the header linter on a set of files.
func header(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() == common.PackageMain {
			// Ignore the main package, there should really one ever be one file in the
			// main package and it should contain func main, leaving implementation to
			// exist in the calling functions that exist in their own packages.

			continue
		}

		var foundHeaderComment bool
		for _, comment := range file.Comments {
			line := pass.Fset.PositionFor(comment.Pos(), false).Line

			// Check to see if the comment starts at the first line of the file.
			if line == 1 {
				foundHeaderComment = true
				break
			}
		}

		if !foundHeaderComment {
			// Get current filepath for reporting.
			fp := pass.Fset.PositionFor(file.Package, false).Filename

			pass.Reportf(0, "file \"%s\" does not contain a header which should explain the purpose of the file and any notable design decisions", fp)
		}
	}

	return nil, nil
}
