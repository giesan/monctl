# Local GUI run task — design

## Goal and behavior

Add `task run` to the Taskfile v3 build shortcuts. It invokes the existing `build` task first, then starts `./build/monctl` in the foreground with no arguments to open the desktop window. Rebuilding before launch ensures local tests use the current source. Closing the window ends the task. Build or application failures should return a nonzero task status; do not start an older binary if the build fails.

The default `task` continues to build only. `task test` and `task package` retain their behavior. The new task does not modify installed binaries, the root-level `./monctl`, or `dist/`. The user's current changes to Taskfile and README remain in place.

## Documentation and verification

Mention `task run` next to the local build instructions in README, including the need for a graphical X11/XWayland desktop session. Verify the command order and syntax via the Task CLI if available, otherwise validate the Taskfile YAML and the build command without launching a desktop window in an automated environment. Do not use the task to change hardware settings in automated tests.
