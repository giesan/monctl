# Ignore non-controllable displays during DDC detection

## Problem

The supplied screenshot shows `ddcutil detect --brief` reporting only `Invalid
display` entries, preceded by a connector-discovery warning. Detection currently
rejects this output as unexpected, exposing a long diagnostic in the window.

## Design

Update detection in `internal/ddc/ddc.go` to recognize both valid `Display N`
headers and `Invalid display` headers as block boundaries. Skip invalid blocks
and parse valid blocks independently, so fields from an invalid entry cannot
be attributed to a valid display.

A successful detection containing only invalid entries returns no monitors and
no error. This includes the connector warning shown in the screenshot. The
existing window model then displays `No external monitor connected`; laptop
brightness remains available without changes to the UI.

Keep command failures, permission errors, missing I2C devices, malformed valid
entries, and unrecognized output without display headers visible as errors.
Do not generally suppress diagnostics or reinterpret unsuccessful commands as
successful detection.

## Alternatives

Suppressing the diagnostic in the window would hide the parser defect and could
hide real failures. Treating all unexpected output as an empty result has the
same risk. Recognizing invalid display blocks fixes the cause in the shared
backend used by both CLI and window.

## Verification

Add regression tests for the screenshot output, invalid-only output, mixed
valid and invalid blocks, and a malformed valid block followed by an invalid
block. Retain coverage for genuine command failures and unexpected output.
Run the Go tests; no monitor writes or system configuration changes are needed.
