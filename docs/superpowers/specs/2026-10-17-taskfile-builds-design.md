# Taskfile build shortcuts — design

## Goal

Provide short, repeatable commands for building the local GUI executable, running tests, and producing Arch/Manjaro snapshot packages. Use the Taskfile v3 format supported by the `task` (go-task) executable. Task is a development tool, not a runtime dependency.

## Tasks and outputs

- `task` defaults to `task build`; both compile `./cmd/monctl` for the current machine with `CGO_ENABLED=0` and the `nucular_shiny` build tag, writing `build/monctl`. Create the build directory as needed. Do not overwrite an existing `./monctl` file.
- `task test` runs `go test ./...` and `CGO_ENABLED=0 go test -tags=nucular_shiny ./...` to check both configurations. Tests must not change hardware settings.
- `task package` invokes the existing GoReleaser v2 configuration with `goreleaser release --snapshot --clean`, producing Linux amd64/arm64 archives and Arch packages in `dist/`. The user explicitly permits clearing and replacing existing contents of `dist/` for this task. The task is opt-in and must not be a dependency of the default task.

Add `build/` to `.gitignore` without changing how `dist/` is ignored. Update README build and package instructions to show the Task commands, their output paths, the `task` prerequisite, and that package builds replace `dist/` artifacts. Keep direct Go and GoReleaser commands available as alternatives. Do not modify application behavior or other worktree files.

## Verification

Check Taskfile syntax with the Task CLI if installed; otherwise report that task execution could not be tested directly and validate the equivalent Go build and tests. Do not run `task package` merely to validate its declaration because it clears existing `dist/` contents. Verify the new build goes to `build/monctl` and leaves the repository-root `monctl` untouched.
