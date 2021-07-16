// Package todo contains the necessary logic for the todo linter.
package todo

import (
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// doc defines the help text for the todo linter.
const doc = `Ensures that each TODO comment defined in the codebase conforms to the
followig format: TODO(<gh-user>)[<jira-ticket>]: <summary>`

// Analyzer exports the todo analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: "todo",
	Doc:  doc,
	Run:  todo,
}

var (
	// reTodo is the regular expression that matches the required TODO format by this
	// linter.
	//
	// For examples, see: https://regex101.com/r/yihH4b/1
	reTodo = regexp.MustCompile(`^TODO\((?P<ghUser>[\w\d-]+)\)\[(?P<jiraTicket>[A-Z]+-\d+)\]: .+$`)

	// These subexpression indexes are placeholders just incase they're ever needed to
	// be used for whatever reason in the future.
	_ = reTodo.SubexpIndex("ghUser")
	_ = reTodo.SubexpIndex("jiraTicket")
)

// todo is the function that gets passed to the Analyzer which runs the actual
// analysis for the todo linter on a set of files.
func todo(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

				if strings.HasPrefix(text, "TODO") {
					if !reTodo.MatchString(text) {
						pass.Reportf(comment.Pos(), "TODO comment must match the required format: TODO(<gh-user>)[<jira-ticket>]: <summary>")
					}
				}
			}
		}
	}

	return nil, nil
}
