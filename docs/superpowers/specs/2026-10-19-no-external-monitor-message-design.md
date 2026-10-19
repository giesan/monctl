# Friendly message when no external monitor is connected

## Goal

When the desktop window starts without an external monitor, show `No external monitor connected` instead of a raw `ddcutil` error. Show the message even if the laptop brightness slider is available. Keep the interface in English.

## Design

The DDC detection backend distinguishes the absence of displays from detection failures. A successful empty `ddcutil detect --brief` result and a recognized `no displays found` response both produce an empty display list without an error. Other failures, including missing I2C devices, missing `ddcutil`, permission errors, timeouts, and malformed output, remain errors with their existing diagnostics. Limit this classification to detection; other DDC operations retain their error handling.

The window model maps an empty display list with no detection error to the friendly message, whether or not a laptop backlight is available. The no-monitor status persists after successful laptop brightness writes; a write error may temporarily replace it until a subsequent successful write. A detection error still shows the existing error message. A detected monitor whose features are unsupported retains the existing `No controls available` behavior. No CLI code or GTK layout changes are needed; the existing `list` command handles empty results and the existing status label displays the message.

## Verification

Add backend tests for empty and recognized no-display responses and for unrelated failures. Add window-model tests for no displays with and without laptop controls, and ensure detection errors and unsupported-feature behavior remain distinct. Run the Go tests.
