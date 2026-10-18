# English repository text and AI disclosure — design

## Goal

Make all versioned, project-authored text in the repository English, including historical design documents, and tell users in the README that the entire project was implemented with AI. Keep behavior and technical claims unchanged.

## Changes

Translate the remaining German document, `docs/superpowers/specs/2026-10-15-single-window-entrypoint-design.md`, into English without rewriting it to match the present-day GTK implementation: it records an earlier design. Add the exact, concise statement “This project was implemented entirely with AI.” to `README.md`, visible to readers of the user guide. Inspect other versioned text files (documentation, code strings and comments, tests, launchers, build/package configuration and metadata) for German text and translate any project-owned findings. Do not translate proper names, code symbols, external library strings, third-party dependencies, binary assets, or OS/tool diagnostics supplied at runtime. Do not introduce a translation framework or alter program behavior.

## Verification

Review the tracked-file list and search text files for German words and accented characters, then inspect likely false negatives manually. Check that the translated historical spec preserves its original meaning and that the README statement is unambiguous. Run `git diff --check`; run tests if executable source or tests need edits.
