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
```
<!--- EndBlock(custom) -->

## Dependencies and Setup

### Dependencies

Make sure you've followed the [Launch Plan](https://outreach-io.atlassian.net/wiki/spaces/EN/pages/695698940/Launch+Plan).

<!--- Block(dependencies) -->
<!--- EndBlock(dependencies) -->
