# User-facing README and application screenshot — design

## Goal

Make the English README a guide to using the installed application rather than developing it. Both the GTK window and CLI are installed by the packages, so document both. Show an actual screenshot of the application window. Keep the in-progress Debian/Ubuntu packaging changes intact.

## Documentation split

`README.md` describes the application's purpose, installation of an already-built Arch/Manjaro or Debian/Ubuntu package, runtime prerequisites and initial `monctl setup`, launching the window, supported sliders, and practical CLI examples using the installed `monctl` command. Include brief troubleshooting for DDC/CI and I2C permissions. Link to `DEVELOPMENT.md` rather than including source-tree commands, Go/Task/GTK development dependencies, build tags, tests, GoReleaser invocations, or package build and compatibility notes in the README.

`DEVELOPMENT.md` contains the displaced source build/run/test instructions, development dependencies, snapshot packaging commands and compatibility caveat for packages built on the host. Avoid dropping useful instructions from the current README, including the uncommitted package documentation. No application or packaging behavior changes are needed.

## Screenshot

Capture the real GTK application window on the available desktop, crop to the application window only, inspect it for private information, save a suitable image in `assets/`, and reference it from `README.md` with useful alt text. Do not show other desktop windows. If capturing a genuine clean application screenshot is not possible, report the limitation instead of committing a fabricated or unreviewed image.

## Verification

Review both documents for user/developer separation, correct package filenames or discoverable filename instructions, working relative links and image path, and accurate commands. Inspect the screenshot visually and ensure no private or unrelated desktop content appears. Check `git diff --check`. Do not install packages on the host or change monitor settings for documentation verification.
