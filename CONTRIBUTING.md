# Contributing to monctl

Thanks for your interest in contributing.

## Before you start

Check existing issues before opening a new one. For a substantial change,
open an issue first to discuss the approach. For a bug report, include your
Linux distribution, the steps to reproduce the problem, and what you expected
to happen. Do not post credentials or other sensitive information in issues.

For a suspected security vulnerability, follow [SECURITY.md](SECURITY.md)
instead of opening a public issue or pull request.

## Make a change

See [DEVELOPMENT.md](DEVELOPMENT.md) for build requirements, local commands,
and packaging details. Keep changes focused, format Go code with `gofmt`,
and add or update tests when changing behavior. Before opening a pull request,
run the checks used by CI when you can:

```sh
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go test ./...
CGO_ENABLED=1 go vet -tags=gtk ./...
CGO_ENABLED=1 go test -tags=gtk ./...
```

GTK-tagged checks require the GTK 3 development files described in
`DEVELOPMENT.md`. Hardware-dependent behavior may need manual testing;
describe what you tested and any limitations in your pull request.

## Open a pull request

Create a branch and open a pull request against `main`; direct pushes to
`main` are not permitted. Explain the change, link relevant issues, and list
the checks you ran. Include a screenshot if you change the desktop window.

GitHub Actions must pass before a pull request can be merged. Workflows from
external contributors' forks require maintainer approval before they run.
Changes from other contributors require the repository owner's review.
The maintainer handles version tags and releases; please do not create a
release tag for a pull request.
