// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: See package comment for this one file package.

// Package todo contains the necessary logic for the todo linter. This linter ensures that
// all TODO comments contain a Jira ticket.
package todo

import (
	"go/ast"
	"regexp"
	"strings"

	"github.com/getoutreach/lintroller/internal/common"
	"github.com/getoutreach/lintroller/internal/reporter"
	"golang.org/x/tools/go/analysis"
)

// name defines the name of the todo linter.
const name = "todo"

// doc defines the help text for the todo linter.
const doc = "Ensures that each TODO comment defined in the codebase conforms to the " +
	"format `TODO(<gh-user>)[<jira-ticket>]: <summary>`, with one of `(gh-user)` or `[jira-ticket]` being required."

// Analyzer exports the todo analyzer (linter).
var Analyzer = analysis.Analyzer{
	Name: name,
	Doc:  doc,
	Run:  todo,
}

// reTodo is the regular expression that matches the required TODO format by this
// linter. This is a TODO followed by one or more of a username in parens and a
// Jira ticket ID in brackets.
var reTodo = regexp.MustCompile(`^TODO(\([\w-]+\))?(\[[a-zA-Z\d-]+\])?: .+$`)

// todo is the function that gets passed to the Analyzer which runs the actual
// analysis for the todo linter on a set of files.
func todo(_pass *analysis.Pass) (interface{}, error) {
	// Ignore test packages.
	if common.IsTestPackage(_pass) {
		return nil, nil
	}

	// Wrap _pass with reporter.Pass to take nolint directives into account.
	pass := reporter.NewPass(name, _pass)

	for _, file := range pass.Files {
		// Ignore generated files and test files.
		if common.IsGenerated(file) || common.IsTestFile(pass.Pass, file) {
			continue
		}

		for _, commentGroup := range file.Comments {
			for _, comment := range commentGroup.List {
				if !matchTodo(comment) {
					pass.Reportf(comment.Pos(),
						"TODO comment must start the line, have a github username and / or a Jira ticket, and be followed by a colon and space: "+
							"`TODO(<gh-user>)[<jira-ticket>]: `")
				}
			}
		}
	}

	return nil, nil
}

// matchTodo returns true if the given comment does not have a TODO, or if it
// matches the required TODO format.
func matchTodo(comment *ast.Comment) bool {
	text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

	if strings.HasPrefix(text, "TODO") {
		matches := reTodo.FindStringSubmatch(text)
		// Verify that we matched & saw at least one of the username or jira ticket.
		return len(matches) >= 3 && (matches[1] != "" || matches[2] != "")
	}

	return true
}
