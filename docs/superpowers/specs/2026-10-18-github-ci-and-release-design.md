# GitHub CI and tagged releases — design

## Goal

Run checks on pull requests and pushes; automatically publish the existing Linux amd64 `.deb`, Arch/Manjaro package, tarball, and checksums as a GitHub Release when a `vMAJOR.MINOR.PATCH` tag passes checks. This follows the version-display feature: released binaries must report the same tag in `monctl --version` and the GTK title. Do not push a tag or release from the local development environment.

## GitHub Actions

Add one workflow with PR, main-branch push, and version-tag push triggers. A read-only checks job on Ubuntu 22.04 installs Go from `go.mod` and GTK 3 development dependencies, verifies `gofmt`, runs `go vet` and headless plus GTK-tagged Go tests, and validates the GoReleaser configuration. A release job runs only for semver-like `vX.Y.Z` tags, depends on successful checks, and receives `contents: write` only in that job. Fetch complete Git history/tags for GoReleaser; use a non-interactive `GITHUB_TOKEN` supplied by Actions. Pin action/tool versions to known, compatible versions; prevent competing release runs for the same ref. No CLI or GTK tests require connected monitors; GTK tests may skip display-only checks on a headless runner.

## Release and compatibility

Enable GitHub publishing in `.goreleaser.yaml` while retaining the local `goreleaser release --snapshot --clean` workflow, which does not publish. Keep the current GTK-enabled linux/amd64 build and nfpm package contents/dependencies; GoReleaser attaches the packages, archive, and checksum file to the GitHub Release and injects the release version. Use an Ubuntu 22.04 runner to build against an older GTK/glibc baseline rather than `ubuntu-latest`. The first release still requires testing on intended Debian/Ubuntu/Arch target systems; glibc compatibility alone is insufficient. Do not claim broader architecture support.

## Documentation and validation

Explain the CI checks, tag/release procedure, permissions, and artifacts in `DEVELOPMENT.md`; show users how to obtain packages from GitHub Releases in `README.md` once publication is configured. Validate workflow syntax and GoReleaser configuration locally, run tests and vet/format checks, and verify snapshot packaging remains nonpublishing if safe to run. Without a configured GitHub remote, report that end-to-end publishing cannot be tested locally. Never create or push a tag or publish a release as part of validation.
