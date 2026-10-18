# 0006. Make CLI Diagnostics Opt-in

Date: 2026-10-07

## Status

Accepted

## Context

ADR 0005 separated transcript data on standard output from diagnostics on standard error. That made pipelines safe, but routine progress, ffmpeg command logging, and native Whisper initialization still overwhelmed interactive terminal output. Redirecting standard error solved clipboard pipelines but made the common command harder to remember and obscured failures as well as routine noise.

## Decision

Successful transcribe commands are quiet by default: standard output contains only transcript text and standard error is empty. Concise failure messages remain on standard error.

Users can pass `--verbose`, or `VERBOSE=1` through the Make target, to enable progress, completion statistics, raw nested failure details, ffmpeg command logging, and native Whisper diagnostics on standard error. Standard output remains transcript-only in every mode.

Clipboard support remains a shell composition such as `make transcribe URL="..." | pbcopy`; the project does not add platform-specific clipboard handling.

## Consequences

The common terminal and clipboard workflows produce only the requested transcript. Troubleshooting data remains available through an explicit, discoverable option. Native Whisper and ffmpeg use process-wide logging controls, so their per-run logging configuration must be synchronized.

This decision refines ADR 0005 by changing routine diagnostics from visible-by-default to opt-in. Its stdout/stderr separation remains unchanged.

## Alternatives Considered

- Keep diagnostics visible by default and require `2>/dev/null` for clean terminal or clipboard workflows.
- Add a platform-specific clipboard command instead of preserving stdout as a composable data interface.
- Suppress diagnostics permanently, which would make downloader, ffmpeg, and native Whisper failures harder to troubleshoot.

## Related

- ADR 0005: Separate CLI Data and Diagnostic Streams
- Planning: `.plan/.done/feat-clean-transcribe-output/`
