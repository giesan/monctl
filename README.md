# monctl

A small Go CLI for controlling external monitor brightness, contrast, and
monitor speaker volume via DDC/CI, plus laptop panel brightness on Linux.
The hardware controls work independently of Cinnamon, GNOME, KDE, X11, and Wayland.

## Requirements

- Go 1.22 or newer to build
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

For a standalone executable, run `go build -o monctl ./cmd/monctl` and use
`./monctl list`, `./monctl get [bus]`, and `./monctl set [bus] <percent>`.
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
reports the ambiguity instead of picking one. A graphical interface is planned
for later.

## Build an Arch/Manjaro package

With GoReleaser v2 installed, run `goreleaser release --snapshot --clean`.
The resulting Arch Linux packages for x86-64 and ARM64 are placed in `dist/`.
The package installs `monctl` to `/usr/bin` and depends on `ddcutil` and
`systemd` (`busctl`). After installing the package, run `monctl setup` as your
normal user to configure the I2C module and check monitor access. Snapshot
packages are for local testing; no publishing or Git tag is required.
