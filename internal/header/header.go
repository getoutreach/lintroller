// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: See package comment for this one file package.

// Package header defines the logic for the header linter. The header linter ensures
// that certain key value pairs are defined at the top of the file in the form of
// comments.
package header

import (
	"fmt"
	"strings"

	"github.com/getoutreach/lintroller/internal/common"
	"golang.org/x/tools/go/analysis"
)

// name defines the name of the header linter.
const name = "header"

// doc defines the help text for the header linter.
const doc = `The header linter ensures each .go file has a header comment section
defined before the package keyword that matches the fields defined via the -fields
flag.

As an example, say the following flag was passed to the linter:

	-fields=Author(s),Description,Gotchas

This would mean the linter would pass on files that resemble the following:

	// Authors(s): <value>
	// Description: <multi-line
	// value>
	// Gotchas: <value>

	package foo

The <value> portion of each field can extend to the next line, as shown in the
value for the description field above; however, there can't be a break in the
comment group for the header, as follows:

	// Authors(s): <value>

	// Description: <value>
	// Gotchas: <value>

	package foo`

// Analyzer exports the doculint analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: name,
	Doc:  doc,
	Run:  header,
}

// NewAnalyzerWithOptions returns the Analyzer package-level variable, with the options
// that would have been defined via flags if this was ran as a vet tool. This is so the
// analyzers can be ran outside of the context of a vet tool and config can be gathered
// from elsewhere.
func NewAnalyzerWithOptions(_rawFields string) *analysis.Analyzer {
	rawFields = _rawFields
	return &Analyzer
}

// Variable block to keep track of flags whose values are collected at runtime. See the
// init function that immediately proceeds this block to see more.
var (
	// rawFields is a variable that gets collected via flags. This variable contains a
	// comma-separated list of fields required to be filled out within the header of a
	// file.
	rawFields string
)

func init() { //nolint:gochecknoinits // Why: This is necessary to grab flags.
	Analyzer.Flags.StringVar(&rawFields, "fields", "Description", "comma-separated list of fields required to be filled out in the header")
}

// header is the function that gets passed to the Analyzer which runs the actual
// analysis for the header linter on a set of files.
func header(pass *analysis.Pass) (interface{}, error) { //nolint:funlen // Why: Doesn't make sense to break this up.
	// Ignore test packages.
	if common.IsTestPackage(pass) {
		return nil, nil
	}

	fields := strings.Split(rawFields, ",")
	validFields := make(map[string]bool, len(fields))

	for _, file := range pass.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(file) || common.IsTestFile(pass, file) {
			continue
		}

		if pass.Pkg.Name() == common.PackageMain {
			// Ignore the main package, there should really one ever be one file in the
			// main package and it should contain func main, leaving implementation to
			// exist in the calling functions that exist in their own packages.

			continue
		}

		// Reset the validFields variable and assume all fields are valid by default
		// for each file in this pass (package).
		for i := range fields {
			validFields[fields[i]] = false
		}

		// Note the package keyword line. All of these header comments must exist before
		// this line number.
		packageKeywordLine := pass.Fset.PositionFor(file.Package, false).Line

		for _, commentGroup := range file.Comments {
			line := pass.Fset.PositionFor(commentGroup.Pos(), false).Line
			if line >= packageKeywordLine {
				// Ignore comments past the line that the package keyword is on. These header
				// fields are required to exist before that.
				continue
			}

			var numFound int

			// Look to see if all of the fields are found in the format we expect them to be
			// in (sans-quotes):
			// "<field>: "
			// by a simple strings.Contains check in the entire text of the comment group. If
			// we end up finding all fields we will do further validation.
			for i := range fields {
				if strings.Contains(commentGroup.Text(), fmt.Sprintf("%s: ", fields[i])) {
					numFound++
				}
			}

			// All fields are found, do further validation.
			if len(fields) == numFound {
				for _, comment := range commentGroup.List {
					for i := range fields {
						cleanComment := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
						prefix := fmt.Sprintf("%s: ", fields[i])

						if strings.HasPrefix(cleanComment, prefix) {
							if len(strings.TrimPrefix(cleanComment, prefix)) > 0 {
								// If the current comment line has a field prefix we're looking for and
								// data proceeding the colon and space after the colon, we will mark the
								// field as valid.
								validFields[fields[i]] = true
							}
						}
					}
				}

				// We found a comment block containing all fields, we don't need to search any further.
				break
			}
		}

		// Get current filepath for potential reporting.
		fp := pass.Fset.PositionFor(file.Package, false).Filename

		for field, valid := range validFields {
			if !valid {
				// Required field not found, report it.
				pass.Reportf(
					0,
					"file \"%s\" does not contain the required header key \"%s\" and corresponding value existing before the package keyword",
					fp,
					field)
			}
		}
	}

	return nil, nil
}
