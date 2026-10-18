# Update GoReleaser configuration and CI version

Replace `nfpms.builds` with `nfpms.ids` in both Arch and Debian package definitions,
preserving their existing build selection and all package metadata. Pin the
GoReleaser version in both GitHub Actions steps (configuration check and release)
to v2.18.3, matching the locally tested version and keeping CI consistent with
the new configuration. Preserve the ongoing Arch package rename and README
edits without committing them as part of this specification.

Verify `goreleaser check` succeeds without deprecation warnings. Build a local
snapshot into an isolated temporary directory, leaving `dist/` intact, and
inspect the Arch and Debian package names, installed binary paths, and
checksums. Do not install packages, push tags, or publish a release. Do not
modify unrelated working-tree files.
