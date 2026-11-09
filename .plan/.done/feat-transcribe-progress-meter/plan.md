# Transcribe pipeline progress meter

## Related ADRs

- [ADR 0006: Make CLI Diagnostics Opt-in](../../doc/decisions/0006-quiet-cli-output-by-default.md)

## Requirements

- Add an opt-in `--progress` / `-p` mode to the standalone CLI.
- Render live pipeline stage progress on stderr without contaminating transcript stdout.
- Report download, normalization, and transcription stages from the engine.
- Keep `--verbose` as the detailed troubleshooting mode; progress is concise and compatible with `| pbcopy`.
- Document direct CLI, Make convenience, and clipboard composition.

## Decisions

- Use stage-aware elapsed progress, not a byte percentage: yt-dlp and Whisper currently buffer their work and do not expose a reliable shared byte/percent scale.
- Add an optional `engine.ProgressFunc` callback to `engine.Options`; nil preserves silent library behavior.
- Render a single carriage-return-updated line on stderr and finish with a newline so redirected logs remain readable.
- Keep clipboard behavior composable through shell pipes; do not add platform-specific clipboard code.

## Non-goals

- Exact download percentage or ETA when the underlying tool does not expose one.
- Replacing yt-dlp, ffmpeg, or Whisper diagnostics.
- A built-in clipboard flag.

## Implementation

1. Add pipeline stage callback support to engine options and emit stage transitions.
2. Add a synchronized CLI progress meter with elapsed time and stage position.
3. Parse `--progress` / `-p`, preserve `--verbose`, and update CLI help/Make usage.
4. Add regression tests for callback wiring, progress output, clean stdout, and help.
5. Update README and development/API documentation with progress and clipboard examples.

## Acceptance criteria

- Default behavior is unchanged and silent.
- `build/transcribe --progress URL` shows a live stage meter on stderr and transcript only on stdout.
- `build/transcribe --progress URL | pbcopy` copies only transcript text.
- `--progress --verbose` remains usable for detailed troubleshooting.
- Engine/library callers can receive stage updates without output unless they provide a callback.

## Verification

- Run targeted CLI and engine tests.
- Run all non-handler repository tests and `go vet ./...`, accounting for the existing handler startup-hang backlog item.
- Build the CGO CLI.
- Run a cached native fixture with progress and verify stdout/stderr separation.

## Risks

- Progress output is stage-level rather than percentage-based because subprocess progress streams are not currently captured.
- Carriage-return rendering may appear as repeated lines in log files; the final newline preserves readable completion output.

## Deferred issues

- Native percentage/ETA reporting can be added later if downloader or transcription backends expose reliable progress callbacks.

## Delivered result

- Added `--progress` / `-p` to the standalone CLI and `PROGRESS=1` to the Make convenience wrapper.
- Added optional `engine.ProgressFunc` stage callbacks without changing silent library defaults.
- Added a carriage-return-updated stderr meter with spinner, stage position, and elapsed time.
- Kept `--verbose` for detailed backend diagnostics and stdout transcript-only for clipboard pipelines.

## Verification evidence

- `go test ./examples/transcribe ./engine` passes.
- Cached native Loom fixture with `--progress` completed exit 0 with 4,135 stdout bytes and progress only on stderr.
- `make build-transcribe` succeeds after switching the target to build the whole CLI package.
- CLI help exits 0 and documents `--progress` / `--verbose`.
- `git diff --check` passes.
