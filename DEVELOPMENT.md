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

## Build packages

With GoReleaser v2 installed, run `task package`, or without Task:

```sh
goreleaser release --snapshot --clean
```

This **replaces the contents of `dist/`** with an Arch/Manjaro `.pkg.tar.zst`,
a Debian/Ubuntu `.deb`, and a `.tar.gz` for x86-64 Linux. The packages install
`monctl` in `/usr/bin` along with an application-menu launcher and icons.
Snapshot packages are for local testing; no publishing or Git tag is needed.

The GTK-enabled binary links to libraries on the build host. A successful
package build does not prove runtime compatibility with every Debian/Ubuntu
release: test installation and launch on the target distribution before
distributing a package broadly. For details, see the
[Debian/Ubuntu package design](docs/superpowers/specs/2026-10-18-debian-ubuntu-package-design.md).
