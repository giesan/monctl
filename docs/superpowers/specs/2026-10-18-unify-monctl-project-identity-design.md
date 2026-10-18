# Unify the technical project identity as monctl — design

## Goal

Use `monctl` consistently as the project's technical identifier and future GitHub repository name. The target GitHub and Go module path is `github.com/giesan/monctl`. The local project directory is renamed by the user later, not in this change.

## Scope

Update `go.mod`, internal Go imports, GoReleaser's GitHub owner/name and linker symbol path, and user-facing URLs pointing to the GitHub repository or its releases. Preserve the CLI command, existing package name, archive and executable names (already `monctl`). Keep “Monitor Control” as the visible product name in the GTK window, desktop launcher, SVG title, README text, and the versioned window title; this is an intentional display name, not a competing repository identifier. Preserve ordinary prose such as “monitor controls” and historical design records. Do not configure or rename the local Git remote, move the working directory, push code, or publish packages.

## Verification

Search all tracked text files for the old `github.com/giesan/monitorcontrol` module/repository path and any technical `monitorcontrol` identifiers, while distinguishing product-name and ordinary prose. Run headless and GTK Go tests and vet, build both variants, validate GoReleaser and workflow syntax, and check `git diff --check`. Verify GoReleaser snapshot version injection still uses the renamed linker path without publishing if safe. The GitHub repository must exist at the new path before remote publishing can succeed.
