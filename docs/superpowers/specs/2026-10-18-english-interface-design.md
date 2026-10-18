# English application interface — design

## Scope

Use English for all current application-owned UI strings, including GTK loading/status messages, monitor control names, laptop card and error context, empty state, and percentage formatting. Keep CLI command syntax and underlying hardware behavior unchanged; CLI help and other active user documentation are already in English. Update affected test expectations and active documentation if it mentions the exact UI wording. Existing historical design specs remain as records and are not translated. Diagnostics supplied verbatim by the OS, logind, or `ddcutil` may still follow the system locale.

## Wording and presentation

Use “Brightness”, “Contrast”, and “Monitor volume” for external controls; “Laptop display” for the laptop card and write-error context, with “Brightness” as its row label. Use “External monitors” for detection errors, “Loading monitors…” while scanning, “Window” for GTK construction errors, and “No controls available” when nothing can be adjusted. Format percentages as `0%` through `100%` without a space, preserving the reserved GTK label width, right alignment, and single-line display. Keep product name “Monitor Control” and existing technical bus numbers unchanged.

## Implementation and verification

Change only application-owned text in the current model, layout, GTK adapter, and related tests. Do not add gettext, language selection, or another localization layer for a single supported language. Test the exact English labels, errors, and percentage boundaries (including `100%` without wrapping). Run CGO-free CLI tests, tagged GTK tests, and a GTK build. Automated tests do not change hardware settings. A manual `task run` check can confirm that the window has no remaining German application text.
