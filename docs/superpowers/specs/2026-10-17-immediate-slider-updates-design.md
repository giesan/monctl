# Immediate slider updates — design

## Goal

In the `monctl` window, begin applying brightness, contrast, and monitor-speaker volume changes as soon as a slider moves. Laptop brightness follows the same rule. Remove the current 220 ms quiet-period debounce so changes are not delayed until dragging stops.

## Behavior and implementation

Keep one worker and a one-slot pending-value channel per slider. The worker starts the first device write immediately after receiving a value. While that write is running, new changes replace the pending value; when the write finishes, the worker sends the latest pending value without a timer. Thus a slow DDC/CI transaction cannot build up a queue of stale intermediate values, but the final slider position is eventually sent. No writes run concurrently for the same slider; backend-level DDC serialization remains unchanged. Requests remain nonblocking for the UI. Apply this uniformly to external-monitor features and the laptop panel, with the existing error display unchanged.

The change is limited to slider request processing and its tests/documentation; no CLI or hardware backend changes are needed.

## Verification

Use a controllable fake backend to assert that the first write starts without a fixed waiting period, that multiple changes during a blocked write yield only the latest follow-up value, and that cancellation stops further writes. Verify both monitor and laptop paths, run the Go test suite and the tagged GUI build, and do not change real device settings in automated tests. Hardware calls may themselves take time; "immediate" means no intentional application delay, not instantaneous physical response.
