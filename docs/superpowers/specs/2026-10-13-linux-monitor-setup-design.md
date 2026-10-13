# Linux monitor setup command — design

## Scope

Add `monctl setup` to the existing Go CLI. It checks that the Linux `i2c-dev` module is loaded and configured to load at boot, performs those two actions when needed, then checks whether DDC/CI monitors can be detected. It does not install `ddcutil` or change device permissions. Package creation with GoReleaser is future work.

## Behavior

- Use `/sys/module/i2c_dev` to check the live module and `/etc/modules-load.d/i2c-dev.conf` (contents `i2c-dev` and a newline) for persistent loading.
- If both are present, skip privilege escalation. Otherwise explain what will change, ask `sudo` to run a narrowly scoped `monctl setup --apply` helper, and report errors without claiming success.
- The helper requires effective UID 0. It loads `i2c-dev` with `modprobe` when needed and creates the configuration file only if absent. It accepts a correctly configured file; it does not overwrite a file with unexpected contents. Running setup again is safe.
- Finally, perform a read-only DDC detection and brightness probe as the original user. If access fails, report the existing DDC error, including the device-permission hint when applicable. No monitor is a diagnostic, not a reason to change permissions.
- No privileged command adjusts brightness or gives persistent access to I2C devices.

## Components and testing

Keep filesystem and command operations inside a setup component; the CLI only selects the command and prints the result. Test idempotence, conflicting files, failed commands, absence of privilege, and CLI argument handling with temporary files and injected command runners. Do not run privileged commands in tests. Manually verify on a machine where `i2c-dev` is already configured; testing a fresh install requires another machine or a deliberate teardown.
