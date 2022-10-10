# lintroller
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/getoutreach/lintroller)
[![Generated via Bootstrap](https://img.shields.io/badge/Outreach-Bootstrap-%235951ff)](https://github.com/getoutreach/bootstrap)
[![Coverage Status](https://coveralls.io/repos/github/getoutreach/lintroller/badge.svg?branch=main)](https://coveralls.io/github//getoutreach/lintroller?branch=main)
<!-- <<Stencil::Block(extraBadges)>> -->

<!-- <</Stencil::Block>> -->

Lintroller houses all of the custom linters that Outreach uses for Go projects.

## Contributing

Please read the [CONTRIBUTING.md](CONTRIBUTING.md) document for guidelines on developing and contributing changes.

## High-level Overview

<!-- <<Stencil::Block(overview)>> -->

lintroller is a collection of linting rules for go projects, implemented as
[Analyzers](https://pkg.go.dev/golang.org/x/tools@v0.1.12/go/analysis#Analyzer)
run through
[unitchecker](https://pkg.go.dev/golang.org/x/tools/go/analysis/unitchecker).
This makes lintroller compatible with `go vet`, the recommended way to run lintroller.

### Implemented rules

- `copyright` - Checks that files start with a header that matches a regular expression.
- `doculint` - Checks that packages and various top-level items in the package have well-formed comments.
- `header` - Checks that source code files have structured headers.
- `todo` - Checks that TODO comments:
  - Start the comment line.
  - Have one or more of a github username in parenthesis (`(username)`) or a Jira ticket (`[ticket-123]`), in that order, immediately after the TODO text.
  - Have a colon and space after the username or ticket.
- `why` - Checks that `nolint` comments:
  - Have specific rule(s) they are ignoring (are followed by a colon then one or more comma-separated rules to ignore).
  - Are followed by ` // Why: <explanation>` on the same line.

<!-- <</Stencil::Block>> -->
