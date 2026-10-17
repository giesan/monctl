# External monitor controls in the brightness window — design

## Goal and scope

Add contrast and monitor-speaker volume sliders to the existing `monctl` window alongside brightness. These use the already implemented DDC/CI VCP controls; laptop system audio and laptop contrast are out of scope. CLI behavior does not change.

## Behavior and layout

Group controls by external monitor, identified by model and bus. Within each monitor group show labeled percentage sliders in the order brightness, contrast, monitor volume. Display a slider only if its feature can be read successfully; specifically, hide unsupported features rather than showing unusable controls. Show a laptop brightness slider separately when available. Keep the window usable when a single monitor or feature fails; show a contextual error for unexpected read or write failures. Continue coalescing rapid changes per slider before writing, so the last selected value is applied without flooding DDC/CI. A successful write clears the error message as today.

## Components and data flow

Expose monitor detection and feature-aware reads/writes through the existing monitor service, so the window enumerates external displays once and checks each of the three features on each bus. Keep the UI model responsible for slider state and a small per-slider feature/bus selector; do not duplicate DDC/CI conversions or command parsing in the window. The laptop remains on its existing backlight controller. The Nucular window renders the labels and adjusts its initial height for the number of available controls, bounded to screen-friendly dimensions as it is today.

## Errors and tests

Unsupported DDC features and absent laptop backlight do not create sliders. Unexpected errors remain visible in the window with the monitor and control name; one failing feature must not prevent other available controls from appearing. Retain the existing no-controls message when nothing is available. Test mixed feature support across multiple monitors, correct routing of writes to feature and bus, laptop behavior, and debouncing with fake backends; no automated test changes real device settings. Build the window with the `nucular_shiny` tag and run the full Go test suite.
