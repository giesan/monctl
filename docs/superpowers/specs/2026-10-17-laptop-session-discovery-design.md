# Automatic laptop backlight session discovery — design

## Goal

After installation, changing laptop brightness from the window or `monctl laptop set` should work when the program is started from a graphical terminal managed by the systemd user manager, even if `XDG_SESSION_ID` is absent and logind cannot associate the process PID with a session. Do not require the user to export variables, run as root, or install another brightness utility.

## Approach and boundaries

Keep brightness writes through logind `Session.SetBrightness`. In `internal/backlight`, first try the existing `GetSession(XDG_SESSION_ID)` or `GetSessionByPID(os.Getpid())`. If this lookup fails, ask logind for its sessions and select exactly one whose UID matches the caller, whose session is active, local (not remote), and graphical (`Type` is `x11` or `wayland`). Use that session's object path for the existing write. If there is no match or more than one match, return an actionable error rather than choosing an arbitrary session. Do not fall back from a failed brightness write or bypass logind authorization.

The backlight controller remains the sole owner of session discovery and of the `busctl` calls; both the GUI slider and CLI already use this controller. Preserve the existing `Get` behavior, which only reads sysfs and does not need a session. No installer, window, or command syntax changes are needed.

## Error handling and verification

Keep the underlying logind diagnostics when the direct lookup or enumeration fails. Distinguish no matching graphical session from ambiguous matches and from a failed `SetBrightness` authorization. Validate paths returned by logind before using them. Tests with a fake command runner should cover the existing ID/PID routes, fallback after an unresolvable PID, fallback after an invalid or stale session ID, filtering out another UID, inactive/non-graphical/remote sessions, ambiguous and absent matches, malformed replies, and a write rejected by logind. No automated test should change hardware brightness. On a real Cinnamon desktop, verify both GUI and CLI from a Ghostty terminal without `XDG_SESSION_ID`.
