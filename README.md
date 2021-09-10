# lintroller

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://engdocs.outreach.cloud/github.com/getoutreach/lintroller)
[![CircleCI](https://circleci.com/gh/getoutreach/lintroller.svg?style=shield&circle-token=<YOUR_STATUS_API_TOKEN:READ:https://circleci.com/docs/2.0/status-badges/>)](https://circleci.com/gh/getoutreach/lintroller)
[![Generated via Bootstrap](https://img.shields.io/badge/Outreach-Bootstrap-%235951ff)](https://github.com/getoutreach/bootstrap)

<!--- Block(description) -->
lintroller is contains all of the custom linters that outreach runs against Go code.
<!--- EndBlock(description) -->

----

[Developing and contributing changes](CONTRIBUTING.md) |
[Generated Documentation](https://engdocs.outreach.cloud/github.com/getoutreach/lintroller/)

<!--- Block(custom) -->
## Running Locally

```shell
# This will put the local lintroller binary into your Go bin folder.
go install ./cmd/lintroller

# Navigate to a repository you'd like to run the lintroller against.
go vet -vettool $(which lintroller) ./...

# Example of passing flags:
go vet -vettool $(which lintroller) -header.fields description,gotchas ./...
```

An alternative way to run the tool which is easier for rapid development is to build and then pass the absolute path of
the binary to `go vet -vettool`:

```shell
# In the root of this repository:
make build

# In the root of a repository you're testing against:
go vet -vettool ~/go/src/github.com/getoutreach/lintroller/bin/lintroller ./...
```

The reason this is easier for rapid development is because these two steps can be done in separate terminal windows/panes/
tabs and it removes the annoyance of dealing with cached binaries that come along with `go install`.

## Singular Linters and Flags

To get information regarding singular linters and their flags you can run the following command(s) (after building):

```shell
# Provides a list of linters defined within lintroller.
./bin/lintroller help

# Shows the flags, descriptions, and defaults for an individual linter.
./bin/lintroller help <linter>
```

## Running as a Standalone Tool (not vettool)

Running as a standalone tool can be useful if you want to pass configuration to the linter via a yaml config file and not
be restricted to the rules that vettool implicit applies to a binary:

```shell
# The path to the binary may not necessarily be the same on your machine
~/go/src/github.com/getoutreach/lintroller/bin/lintroller -config lintroller.yaml ./...
```

The structure for the config yaml can be found in `internal/config/config.go`.

## Trimming Absolute Path into Relative Path on Reporting

By default the `*analysis.Pass.Reportf` function will report the absolute path of any linter errors that fire during the
linting process. This is kind of cumbersome, unnecessary, and results in errors that are harder to read due to extraneous
information being present. There is no way to trim these programmatically at the current time, but you can do some bash-fu
to make the output look a little better:

```shell
# The path to the binary may not necessarily be the same on your machine
~/go/src/github.com/getoutreach/lintroller/bin/lintroller -config lintroller.yaml ./... 2>&1 | sed "s#^$(pwd)/##"
```
<!--- EndBlock(custom) -->

## Dependencies and Setup

### Dependencies

Make sure you've followed the [Launch Plan](https://outreach-io.atlassian.net/wiki/spaces/EN/pages/695698940/Launch+Plan).

<!--- Block(dependencies) -->
<!--- EndBlock(dependencies) -->
