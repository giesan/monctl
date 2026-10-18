# Developing Monitor Control

The [README](README.md) covers installation and everyday use of the window and
CLI. This file covers builds, tests, and snapshot packages from the source tree.

## Requirements

- Go 1.27 or newer
- [Task](https://taskfile.dev/) (go-task), optional build shortcuts
- GTK 3 development files and `pkg-config` to build the window: on Debian/Ubuntu,
  install `libgtk-3-dev` and `pkg-config`; on Arch/Manjaro, install `gtk3` and
  `pkgconf` (for example, `sudo pacman -S gtk3 pkgconf`)
- `ddcutil` for running against external monitors (on Manjaro:
  `sudo pacman -S ddcutil`)

The hardware controls work independently of Cinnamon, GNOME, KDE, X11, and
Wayland; the GTK window uses the desktop's theme. On X11/XWayland, `task run`
rebuilds and opens the window, then exits when the window closes.

## Build and run

```sh
task build      # Build the GTK application into build/monctl
task run        # Rebuild and open the window
task test       # Run CLI and GTK test suites
```

Without Task, build and test directly:

```sh
mkdir -p build
CGO_ENABLED=1 go build -tags=gtk -o build/monctl ./cmd/monctl
CGO_ENABLED=0 go test ./...
CGO_ENABLED=1 go test -tags=gtk ./...
```

Run `./build/monctl` to open the window, or invoke its CLI commands such as
`./build/monctl list`. For a quick source-tree run, use
`CGO_ENABLED=1 go run -tags=gtk ./cmd/monctl` for the window or
`go run ./cmd/monctl list` for the CLI. The CLI can also be built without GTK:

```sh
CGO_ENABLED=0 go build ./cmd/monctl
```

Without the `gtk` build tag, calling `monctl` without arguments reports how
to build the window. Window slider changes are sent immediately; while a device
command runs, only the latest pending value is retained to avoid overwhelming
DDC/CI monitors.

Local builds show `dev` in `monctl --version` and the window title. GoReleaser
sets the shared build version through linker flags from the release version;
snapshot packages display a snapshot version instead of `dev`.

## Build packages

With GoReleaser v2 installed, run `task package`, or without Task:

```sh
goreleaser release --snapshot --clean
```

GoReleaser needs a Git remote even for a snapshot once GitHub publishing is
configured. Set `origin` to the hosted repository before running this command;
the snapshot itself never uploads artifacts.

This **replaces the contents of `dist/`** with an Arch/Manjaro `.pkg.tar.zst`,
a Debian/Ubuntu `.deb`, and a `.tar.gz` for x86-64 Linux. The packages install
`monctl` in `/usr/bin` along with an application-menu launcher and icons.
Snapshot packages are for local testing; no publishing or Git tag is needed.

The GTK-enabled binary links to libraries on the build host. A successful
package build does not prove runtime compatibility with every Debian/Ubuntu
release: test installation and launch on the target distribution before
distributing a package broadly. For details, see the
[Debian/Ubuntu package design](docs/superpowers/specs/2026-10-18-debian-ubuntu-package-design.md).

## GitHub checks and releases

GitHub Actions runs formatting checks, `go vet`, headless and GTK-tagged Go
tests, and `goreleaser check` on pull requests and pushes to `main` or a
version tag. GTK development files are installed on the Ubuntu 22.04 runner;
display-dependent tests may skip when no desktop is available.

After the version feature and workflow are merged, create and push a release
tag pointing to the tested commit:

```sh
git tag v1.2.3
git push origin v1.2.3
```

A successful `vMAJOR.MINOR.PATCH` tag build automatically publishes a GitHub
Release with the amd64 `.deb`, Arch/Manjaro package, tarball and checksums.
The binary reports the tag with `monctl --version` and in the window title.
The release job alone has `contents: write` permission; the checks job is
read-only. Protect release tags in GitHub repository settings so only trusted
maintainers can trigger publishing. GoReleaser snapshots remain local and do
not publish anything. Validate end-to-end publishing only after the repository
is hosted on GitHub; a local snapshot cannot test the release upload.
