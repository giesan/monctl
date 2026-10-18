# Content-sized GTK monitor window — design

## Problem and desired behavior

The GTK window currently calls `Resize` with an estimate based on the number of sliders and monitor cards. With one external monitor and a laptop card this overestimates the needed height and leaves a large empty region below the cards. After loading, size the client area to the actual GTK-preferred height of visible cards and any status message. Keep a sensible compact loading state. Do not leave unnecessary space or clip cards, including with a different Cinnamon theme or larger text.

## Sizing and scrolling

Measure GTK widget preferred sizes after the cards are added and shown on the GTK thread; do not calculate height from a guessed per-row constant. Obtain the current display's monitor work area and reserve space for window decorations and margins. Cap the window's client height to that available space; only then use the scrolled card area to reach remaining controls. Recompute height when a status message appears or disappears; hide its unused space when no message is shown. Preserve the current width unless the GTK minimum width needed for controls is larger. Do not interfere with user-initiated resizing after initial layout unless a status message requires more space within the same screen cap.

Keep the GTK presentation separate from monitor and backlight services. Device requests, percentages, and CLI behavior do not change. The final result should show all rows from the supplied screenshot without the large blank bottom region and remain usable on small screens or with many monitors.

## Verification

Test a small pure sizing decision with content smaller and larger than the available work area, plus visible/hidden status height; validate GTK preferred size and scrolling behavior in a display-backed test where possible. Run both the CGO-free CLI tests and tagged GTK tests/build. Visually inspect the one-monitor-plus-laptop case shown by the user and a multi-monitor/small-screen case; no automated test should change hardware settings.
