# Hide monitor bus numbers in the GTK window — design

## Goal

Show only the monitor model as each external monitor card heading, without `(Bus N)`. Retain the `Laptop display` heading. The CLI output and diagnostic errors keep bus numbers so users can distinguish or troubleshoot displays.

## Behavior and implementation

The GTK card grouping must use monitor identity (the existing bus number), not the rendered heading. This preserves separate cards for two monitors with the same model name and keeps each slider routed to its original bus and feature. The heading is presentation text only. No changes to device discovery, CLI syntax, DDC writes, or backlight behavior.

## Verification

Test ordinary two-monitor grouping and two adjacent monitors with identical model names but distinct buses: both produce separate cards with the same bus-free heading and correctly assigned slider indices. Confirm no bus number appears in a card heading while the existing error and CLI tests keep their bus context. Run headless and GTK-tagged tests and build; no automated test changes hardware settings.
