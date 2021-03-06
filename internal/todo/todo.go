// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: See package comment for this one file package.

// Package todo contains the necessary logic for the todo linter. This linter ensures that
// all TODO comments contain a Jira ticket.
package todo

import (
	"regexp"
	"strings"

	"github.com/getoutreach/lintroller/internal/common"
	"github.com/getoutreach/lintroller/internal/nolint"
	"golang.org/x/tools/go/analysis"
)

// name defines the name of the todo linter.
const name = "todo"

// doc defines the help text for the todo linter.
const doc = `Ensures that each TODO comment defined in the codebase conforms to one of the
following formats: TODO(<gh-user>)[<jira-ticket>]: <summary> or TODO[<jira-ticket>]: <summary>`

// Analyzer exports the todo analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: name,
	Doc:  doc,
	Run:  todo,
}

// reTodo is the regular expression that matches the required TODO format by this
// linter. If ever necessary, the subexpression indexes ghUser and jiraTicket can
// be used by calling reTodo.SubexpIndex(...).
//
// For examples, see: https://regex101.com/r/vsbdEm/1
var reTodo = regexp.MustCompile(`^TODO(\((?P<ghUser>[\w-]+)\))?\[(?P<jiraTicket>[A-Z]+-\d+)\]: .+$`)

// todo is the function that gets passed to the Analyzer which runs the actual
// analysis for the todo linter on a set of files.
func todo(pass *analysis.Pass) (interface{}, error) {
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

				if strings.HasPrefix(text, "TODO") {
					if !reTodo.MatchString(text) {
						passWithNoLint.Reportf(comment.Pos(),
							"TODO comment must match one of the required formats: TODO(<gh-user>)[<jira-ticket>]: <summary> or TODO[<jira-ticket>]: <summary>")
					}
				}
			}
		}
	}

	return nil, nil
}
