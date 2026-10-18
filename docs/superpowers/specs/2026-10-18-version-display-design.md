# Version display in CLI and GTK — design

## Goal

Show the same version in `monctl --version` and the GTK window title without changing monitor controls. Local development builds display `dev`; tagged release builds display the tag, such as `v1.2.3`. This is the first of two tasks; GitHub Actions and publishing are designed separately after the version feature is in place.

## Version source and presentation

Keep one Go version value with default `dev` in a small shared package importable by the CLI and GTK window. GoReleaser sets that value from the release Git tag using Go linker flags. Headless builds must not depend on GTK; neither build should require Git at runtime. `monctl --version` prints `monctl <version>` to stdout and exits successfully without opening GTK, detecting monitors, or running setup. The window title becomes `Monitor Control <version>`; the sliders, window body, and application-menu name stay unchanged. Unknown CLI arguments keep their existing usage and exit behavior. Existing local `go build`, `go run`, and `task run` builds show `dev` unless explicitly supplied a version at build time.

## Documentation and verification

Document `monctl --version` briefly in the user README and the linker-injected version in the developer guide. Test the CLI's default and injected output and the shared title formatting without requiring a display or hardware. Run the headless and GTK-tagged test suites, build both variants, and check GoReleaser's configuration. A snapshot can use its snapshot version, but must not claim to be a tagged GitHub release.
