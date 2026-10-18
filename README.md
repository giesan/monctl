# monctl

A small Go CLI for controlling external monitor brightness, contrast, and
monitor speaker volume via DDC/CI, plus laptop panel brightness on Linux.
The hardware controls work independently of Cinnamon, GNOME, KDE, X11, and Wayland.

## Requirements

- Go 1.27 or newer to build
- [Task](https://taskfile.dev/) (go-task) for the optional `task` build shortcuts
- GTK 3 development files and `pkg-config` to build the graphical window (on Manjaro: `sudo pacman -S gtk3 pkgconf`); the window also needs GTK 3 at runtime
- `ddcutil` installed (on Manjaro: `sudo pacman -S ddcutil`)
- A DDC/CI-capable external monitor with DDC/CI enabled in its on-screen menu
- Permission to access its `/dev/i2c-*` device; consult the [ddcutil setup guide](https://www.ddcutil.com/config_steps/) if detection fails

If `list` reports that no `/dev/i2c-*` devices exist, load the kernel module
with `sudo modprobe i2c-dev` and try again. If devices exist but access is
denied, follow the ddcutil guide to configure persistent device permissions.

## Use

Run `monctl setup` once after installation (or `go run ./cmd/monctl setup`
from the source tree). It checks `i2c-dev`, asks for your sudo password only
if it needs to load the module or configure it to load at boot, then checks
for DDC/CI monitors. It does not change device permissions; if access is
denied, follow the ddcutil guide linked above. Repeating setup is safe.

```sh
CGO_ENABLED=1 go run -tags=gtk ./cmd/monctl
go run ./cmd/monctl setup
go run ./cmd/monctl list
go run ./cmd/monctl get
go run ./cmd/monctl set 50
go run ./cmd/monctl get 7
go run ./cmd/monctl set 7 50
go run ./cmd/monctl contrast get
go run ./cmd/monctl contrast set 45
go run ./cmd/monctl volume get
go run ./cmd/monctl volume set 30
go run ./cmd/monctl laptop get
go run ./cmd/monctl laptop set 50
```

For a standalone executable with the window, run `task` (or `task build`).
The result is `build/monctl`; run `./build/monctl` to open the window, or use
`./build/monctl list`, `./build/monctl get [bus]`, and `./build/monctl set [bus] <percent>`.
Use `task run` on an X11/XWayland desktop to rebuild and open the window in
one step. The task exits when the window closes.
Without Task, use `mkdir -p build` and
`CGO_ENABLED=1 go build -tags=gtk -o build/monctl ./cmd/monctl`.
Run `task test` to test both CLI and GUI configurations.
The CLI alone can still be built without GTK using `CGO_ENABLED=0 go build ./cmd/monctl`.
With exactly one brightness-capable external monitor, the bus number is optional.
With multiple monitors, specify a bus number returned by `list`. Bus numbers can change after
reconnecting displays or rebooting. `set` accepts whole percentages from 0 to
100; the actual monitor brightness may be rounded or limited by the hardware.
Use `contrast get [bus]`, `contrast set [bus] <percent>`, `volume get [bus]`,
and `volume set [bus] <percent>` for the other controls. The optional bus
selection works the same way, but only monitors supporting the chosen control
are considered. `volume` adjusts the monitor's DDC/CI speaker volume, **not**
the Linux system volume. Some monitors do not support these controls.

`laptop get` and `laptop set <percent>` control the built-in panel using
`/sys/class/backlight` and systemd-logind (via `busctl`). They do not use DDC/CI
or `ddcutil`. If multiple laptop backlight devices are found, the command
reports the ambiguity instead of picking one.

## Brightness window

Run `monctl` without arguments for a desktop window with brightness, contrast,
and monitor-speaker volume sliders for external monitors, plus a brightness
slider for the laptop panel. Unsupported monitor controls are hidden. Monitor
volume does not affect the Linux system volume. Close the window to exit. The
window uses native GTK 3 widgets with the desktop's GTK theme on X11 or Wayland.
The window carries its own application icon, including when launched locally
with `task run`; the desktop theme determines whether it also appears in the title bar.
Without the `gtk` build tag, only CLI commands are available; the
no-argument invocation reports how to build the window. Slider changes start
immediately; while a device command is running, only the latest pending value
is kept to avoid overwhelming DDC/CI monitors.

## Build an Arch/Manjaro package

With GoReleaser v2 installed, run `task package` (or
`goreleaser release --snapshot --clean`). This replaces existing contents of
`dist/`.
The resulting Arch Linux package for x86-64 is placed in `dist/`.
The package installs `monctl` to `/usr/bin`, an application-menu launcher and
icons for the Cinnamon panel. It depends on `ddcutil` and
`systemd` (`busctl`) plus GTK 3. Building it requires GTK 3 development files
and `pkg-config` on an x86-64 Linux machine. After installing the package, run `monctl setup` as your
normal user to configure the I2C module and check monitor access. Snapshot
packages are for local testing; no publishing or Git tag is required.
