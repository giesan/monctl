# Linux monitor brightness CLI — design

## Goal and scope

Build a Go command-line application for controlling the hardware brightness of external monitors on Linux. The first release lists reachable monitors, reads their brightness, and sets it. It is intended to work in Cinnamon and other desktop environments on either X11 or Wayland. Laptop backlights, contrast, monitor volume, tray integration, and a graphical window are future work.

## Approach

Use the installed `ddcutil` executable as the initial DDC/CI backend. DDC communication does not depend on the window system. Keep the backend behind a small monitor-control interface so later backends, including a laptop-backlight backend, can be added without changing the commands. No daemon or desktop integration is required for the first release.

## CLI and behavior

- `monctl list` lists DDC/CI-reachable external monitors with their model and current I2C bus number. The bus number is the selector used by the other commands; it can change when hardware is reconnected or after a reboot, so users should consult `list` again.
- `monctl get <bus>` reads the selected monitor's DDC/CI brightness (VCP code `0x10`) and prints its current value as a percentage of the reported maximum.
- `monctl set <bus> <percent>` accepts an integer from 0 through 100, reads the monitor's reported maximum, converts the percentage to a device value, and sets VCP code `0x10` on that bus. Round to the nearest supported integer. The command reports success or an actionable error; it does not promise that the physical luminance is linear in the percentage.
- Invalid arguments return a nonzero exit status and usage information. Diagnostic messages go to standard error; command results go to standard output.

## Components and data flow

1. The CLI parses commands, validates the bus selector and percentage, and renders results.
2. A monitor service implements listing, reading, and setting in terms of a small DDC backend interface. It handles normalization between device values and percentages.
3. The initial backend invokes `ddcutil detect`, `ddcutil getvcp 10`, and `ddcutil setvcp 10` with an explicit bus selector. Invocations use argument arrays, not a shell, and have timeouts. Device commands for a selected bus are serialized so concurrent callers cannot overlap transactions.

Each component has one responsibility: UI/arguments, brightness semantics, or hardware access. Future tray and window frontends can call the monitor service; a future laptop backend can implement the same brightness operations.

## Failure handling

If `ddcutil` is unavailable, report how to install it. If no suitable display is found or the selected bus is absent, explain that detection should be rerun. For DDC errors, distinguish where possible between a monitor with DDC/CI disabled, missing access to `/dev/i2c-*`, an unsupported brightness feature, and a timeout; otherwise include the underlying diagnostic. Do not report a successful set when `ddcutil` fails. Display availability and support are determined by probing, not by assuming every external monitor implements brightness control.

## Verification

Test argument validation, percentage conversion including non-100 maxima and boundary values, and backend output/error handling using a fake command runner. On a real Linux machine, verify detection, read, set, and read-back with a DDC/CI-enabled external display, including the error path for missing device access. Automated tests must not change real monitor settings.

## Dependencies and limits

The user needs `ddcutil`, a monitor and connection that expose DDC/CI, and access to the appropriate I2C devices. Some docks, drivers, and monitors do not provide a working DDC channel. X11/Wayland independence applies to these CLI hardware operations; future shortcuts, OSD, and tray behavior may need desktop-specific integration.
