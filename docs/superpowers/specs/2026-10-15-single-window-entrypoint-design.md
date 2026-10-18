# One window instead of a tray — design

## Goal and behavior

`monctl` is the only distributed program. With no arguments, it opens the existing brightness window with sliders for detected external monitors and, if present, the laptop display. Closing the window exits the program. The existing CLI commands (`setup`, `list`, `get`, `set`, `contrast`, `volume`, `laptop`) retain their arguments and return values and work without a graphical session. Unknown arguments continue to show usage information and exit with code 2.

The window continues to use Nucular's CGO-free Shiny backend and requires X11 or XWayland. If `DISPLAY` is missing, `monctl` reports the error and exits with code 1. Wayland without XWayland is outside the scope of this step.

## Structure

The CLI entry point distinguishes only between invocation without arguments and a CLI command. The window path calls `internal/window.Run` with the existing monitor and laptop services; the window model and its deferred writes remain in place. The GUI integration in the CLI is tied to the `nucular_shiny` build tag. Without that tag, the CLI remains buildable; invocation without arguments prints a clear build instruction. The distributed `monctl` is built with that tag and `CGO_ENABLED=0`. A window error is written to stderr and results in exit code 1.

The tray entry point, tray package, and separate `monctl-window` entry point are removed. The tray-specific implementation and its tests are removed; `internal/window` remains a separate component. Go dependencies no longer needed afterward are cleaned up. The package configuration distributes only `monctl`. The README describes launching the window, running from source with the build tag, and the still-available CLI commands.

## Verification

Tests cover the no-argument branch, including window errors, and unchanged CLI argument handling without requiring a display or real monitor hardware. `go test ./...` and `go test -tags=nucular_shiny ./...` check both build variants; a package build checks that only one binary is distributed. A visual check on X11/XWayland confirms that the window opens, sliders appear, and closing the window exits the program.
