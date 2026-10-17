# Faster external-monitor slider writes — design

## Problem and goal

The window no longer waits for a 220 ms debounce, but an external-monitor `SetFeature` still performs a `ddcutil getvcp` before every `setvcp` to learn the feature maximum. A read-only `getvcp 10` on the local monitor took about 0.4 seconds. Remove this redundant per-write read for window sliders without changing CLI behavior or promising faster DDC hardware than the monitor provides.

## Data flow and components

Extend the monitor service with a feature read that returns both normalized percent and the device maximum. Window initialization already reads each available feature; retain its maximum in that slider's state. Add a service write method that accepts that validated maximum, converts percent with the existing rounding and bounds rules, then calls the backend `Write` directly. Share conversion/validation code with the existing `SetFeature` method to avoid diverging CLI and window semantics. Keep CLI `SetFeature` behavior unchanged: it continues reading the current maximum for each command.

The window worker uses the saved maximum for external-monitor writes, so each change entails only a `setvcp` call. The laptop path remains unchanged. When reopening the window, feature maxima are read again; a device that disconnects or rejects a stale value reports the existing write error, and the user can reopen the window to refresh. Do not add persistent caches, extra background polling, new dependencies, or an intentional delay. Continue keeping only the latest pending slider value while a write is running.

## Verification

Test percent conversion with non-100 maxima and boundary/invalid values, check that a window drag uses the captured maximum and does not issue another read, and verify that CLI `SetFeature` still reads. Preserve tests for unsupported controls and write errors. Run the normal and `nucular_shiny` test suites and race tests; do not change physical monitor settings in automated tests. A manual check with `task run` is needed to judge perceived monitor response because DDC/CI writes themselves take hardware-dependent time.
