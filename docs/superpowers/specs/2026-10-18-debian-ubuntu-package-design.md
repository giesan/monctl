# Debian and Ubuntu package — design

## Goal

Extend the existing GoReleaser snapshot workflow with an amd64 `.deb` for current Debian and Ubuntu releases while preserving the existing Arch/Manjaro package and tarball. `task package` should produce all artifacts in `dist/` in one run; it continues to replace prior `dist/` contents. This is packaging only: do not change the application or CLI behavior.

## Package configuration

Add a Debian-format nfpm entry using the existing Linux amd64 GTK-enabled `monctl` build. Install `/usr/bin/monctl`, `monctl.desktop` under `/usr/share/applications`, and the same 16/24/48/128 px hicolor icons used by the Arch package. Include dependencies on `ddcutil`, `systemd` (for logind/busctl), and a Debian alternative for GTK 3's package names (`libgtk-3-0 | libgtk-3-0t64`). Keep package name, version and description aligned with the Arch package. Document the new `.deb` path, local installation via `apt install ./...deb`, and GTK 3 development requirements for building on Linux.

## Compatibility and verification

Validate GoReleaser's configuration, run the existing headless/GTK test suites, produce a snapshot, and inspect `.deb` metadata and installed paths with `dpkg-deb`; verify all checksums. Check the linked binary's glibc requirement against the oldest intended target (Debian 12). The same binary is built on the current Linux host, so a successful package build does not prove that GTK/GLib symbols and other shared libraries are compatible with every Debian 12/13 and Ubuntu 22.04/24.04/26.04 installation. Test an actual install and launch on target systems or suitable containers before claiming full compatibility; if that fails, move the Debian build to an older compatible environment rather than misrepresenting package support. Do not install the package on the user's host as part of automated validation.
