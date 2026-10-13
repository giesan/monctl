# monctl

A small Go CLI for controlling external monitor brightness on Linux via DDC/CI.
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

```sh
go run ./cmd/monctl list
go run ./cmd/monctl get 7
go run ./cmd/monctl set 7 50
```

For a standalone executable, run `go build -o monctl ./cmd/monctl` and use
`./monctl list`, `./monctl get <bus>`, and `./monctl set <bus> <percent>`.
Replace `7` with a bus number returned by `list`. Bus numbers can change after
reconnecting displays or rebooting. `set` accepts whole percentages from 0 to
100; the actual monitor brightness may be rounded or limited by the hardware.

This first version supports external monitor brightness only. Laptop backlights,
contrast, volume, and graphical controls are planned for later.
