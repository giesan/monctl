# Native Cinnamon-style monitor window — design

## Goal and visual design

Replace the current Nucular/Shiny window with a GTK 3 window themed by Cinnamon. Retain the reference image's compact vertical stack of rounded monitor cards, but use the system's light/dark colors, typography, native title bar, and GTK slider behavior instead of a fixed pink palette. Each external monitor card shows its name and up to three equally prominent rows in order: brightness, contrast, and monitor-speaker volume. Each row has a meaningful icon or text fallback, a short accessible label, a GTK scale, and a right-aligned percentage label. Reserve enough width for `100 %` and never wrap the number and percent sign. The laptop card only has brightness. Hide unsupported features rather than rendering disabled sliders. No bottom toolbar or placeholder settings button: the native title bar supplies Close.

Size the window for a compact panel on Cinnamon and use a vertical scroller when several monitors are present. Give the cards modest padding and rounded corners using narrowly scoped GTK CSS; inherit foreground/background/accent colors from the GTK theme instead of hardcoding theme-wide colors. The percentages and controls must remain readable and keyboard accessible under light and dark themes and at larger text scales. An error area reports a failing control without disabling unrelated controls; use a no-controls state when no supported devices exist.

## Architecture and interaction

Keep the existing `internal/window` slider model and monitor/backlight services as the behavioral core; replace only the toolkit-specific presentation with a GTK 3 adapter built via Go GTK 3 bindings and CGO. Create GTK widgets on the GTK main thread. Start with a loading state, perform monitor discovery and initial DDC/backlight reads off that thread, then publish the populated model and subsequent status changes to GTK on its main thread. On close, cancel outstanding work and terminate as the current window does. Programmatic initial values must not trigger device writes.

On user interaction, update the percentage label immediately, keeping the existing one-pending-value-per-slider behavior. DDC writes use the maximum captured at initial read rather than repeating `getvcp`; laptop writes keep using logind. Apply read/write errors without blocking unrelated sliders. CLI arguments and their headless behavior remain unchanged.

## Build and packaging

Use a GUI-specific `gtk` build tag for the GTK presentation; without it, the CLI stays buildable without CGO or GTK development packages and invoking the window reports the build requirement. Remove the old Nucular/Shiny GUI path and dependencies after the GTK path passes tests. `task build` and `task run` build the GTK window with CGO enabled and check for GTK 3 development headers through pkg-config. `task test` covers headless and tagged builds. The GoReleaser Arch package builds for Linux amd64 only, with GTK 3 runtime libraries as a package dependency alongside the existing `ddcutil` and `systemd` dependencies. Cross-compilation to arm64 is not part of this change. Update build and runtime requirements and commands in README; do not change installed CLI syntax.

## Verification and limits

Unit tests cover feature ordering, unsupported controls, no-controls/error states, and nonblocking last-value writes without real hardware. Tagged GTK tests validate widget construction, icon/text fallback, percentage labels including `100 %`, no wrapping, and avoiding writes on initialization where practical without requiring a real display. Run headless and tagged Go tests, build the amd64 GUI binary, and check GoReleaser's configuration without clearing the existing `dist/` merely for verification. Visually check the compact layout, scrolling, theme adaptation, keyboard operation, and readable status under Cinnamon X11 and Wayland. GTK may still require a graphical display and actual DDC/CI hardware writes can take time; the UI should stay responsive while they do.
