# Monitor Control application icon — design

## Goal and visual direction

Add a recognizable icon for the GTK window, Cinnamon taskbar, and installed application launcher. Use the approved symbol: a monitor outline with two horizontal slider tracks and offset knobs. Keep the geometry and contrast legible at 16–24 px; use a light/neutral screen and a restrained blue accent, not a fixed pink background. Retain a vector SVG master and checked-in transparent PNGs at 16, 24, 48, and 128 px. The icon is application identity, not a new control in the window.

## Runtime and installation

Embed a PNG in the GTK build and set the GTK window/application icon at startup, so local `task run` works without any installed asset or an absolute source-tree path. On X11 the window icon should appear in the taskbar and where window decorations display icons. On Wayland the compositor and desktop launcher influence taskbar matching; set a stable application identity where supported and use a matching desktop entry. Do not add a custom widget to the native title bar: Cinnamon may choose not to show title-bar icons even if one is set.

Include a `monctl.desktop` launcher in the existing Arch/Manjaro package, with `Exec=monctl`, `Icon=monctl`, appropriate Settings category, and a matching startup identity. Install appropriately sized PNGs in the standard hicolor icon directories and keep the SVG master in the repository; package layout must not affect the headless CLI or require root for `task run`. Do not package an unrelated screenshot or browser mockup files.

## Verification

Check that committed PNGs can be decoded, have the intended sizes and transparency, and stay legible at taskbar size. Test GTK icon loading without opening real monitor controls and validate the GoReleaser package configuration. Build GUI and headless CLI; do not run a package cleanup simply for validation. Check local `task run` and the installed launcher visually under Cinnamon X11/Wayland where available. Verify taskbar/menu identity and title-bar behavior without claiming the application can override window-manager theme decisions.
