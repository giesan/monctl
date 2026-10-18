# Distinguish the Arch package and document uninstalling

## Goal

Prevent yay from treating the unrelated AUR `monctl` package as an update to
this project's locally installed Arch package. Document removal of installed
packages. Do not add migration instructions to the public README: the only
existing installation affected by the old name is the maintainer's machine.

## Changes

Name only the Arch/Manjaro package `monctl-giesan` in GoReleaser. Preserve the
Debian package name `monctl`, the executable `/usr/bin/monctl`, launcher, icons,
repository identity, and application title. Do not use package `provides`,
`replaces`, or `conflicts` to claim the unrelated AUR package's identity. Update
the Arch release filename pattern and install command in the README to match
the new package name. Add concise Arch and Debian uninstall instructions using
`pacman -R monctl-giesan` and `apt remove monctl`, respectively. No automatic
modification of existing installations and no host package-manager commands.

## Verification

Run GoReleaser configuration validation and existing tests. Build a local
snapshot into an alternate temporary output directory, preserving existing
`dist/` files. Inspect the generated Arch package name and installed binary
path, as well as the unchanged Debian package metadata. Check that the README
commands match the produced artifact names. Do not push tags, publish a GitHub
release, install packages on the host, or alter the user's untracked files.
