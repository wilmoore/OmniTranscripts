# Clean transcribe output

## Requirements

- Print only transcript text for a successful default `make transcribe` run.
- Keep concise, actionable errors visible on failed default runs.
- Preserve progress, completion statistics, and low-level downloader, ffmpeg, and native Whisper diagnostics behind an opt-in verbose mode.
- Keep stdout as the transcript data channel so output can be piped directly to `pbcopy`, a file, or another command.

## Decisions

- Extend accepted ADR 0005: stdout remains transcript-only, while routine stderr progress becomes opt-in instead of default.
- Expose `--verbose` on the CLI and `VERBOSE=1` on the Make target.
- Do not add a clipboard-specific target; document the composable `make transcribe ... | pbcopy` workflow.
- Keep default failure messages concise, but include raw nested diagnostics in verbose mode.
- Make engine diagnostics explicit through `Options.Verbose`; default library behavior is quiet.

## Non-goals

- Changing transcript text, segmentation, caching, backend selection, or downloader behavior.
- Adding persistent log files or a logging framework.
- Implementing clipboard support separately for each operating system.

## Implementation

1. Parse `--verbose` without changing the positional input contract.
2. Gate CLI progress and summaries on verbose mode and pass that mode into the engine.
3. Silence engine, ffmpeg, and native Whisper diagnostics by default; restore them in verbose mode.
4. Keep concise errors visible and gate raw downloader/cause details on verbose mode.
5. Update Makefile help and developer documentation with quiet, verbose, and clipboard examples.
6. Add regression tests for default, verbose, usage, and error output contracts.

## Acceptance criteria

- Default successful execution writes exactly the transcript to stdout and nothing to stderr.
- Verbose successful execution still keeps stdout clean and writes progress plus summary to stderr.
- Native Whisper and ffmpeg do not emit diagnostics by default and remain available in verbose mode.
- Default failures are actionable without dumping raw tool output; verbose failures include diagnostic details.
- `make transcribe URL='...' | pbcopy` works without stderr redirection.

## Verification

- Run targeted CLI and engine tests.
- Run the full Go test suite with the repository's default non-CGO configuration.
- Build the CGO transcribe binary to validate native logging integration.
- Exercise a cached/local transcription in quiet and verbose modes when local native assets are available.

## Risks

- Whisper logging is configured through a process-wide C callback; configuration must be synchronized to avoid cross-call leakage.
- ffmpeg-go exposes command logging as a package global; access must be synchronized when toggled per call.

## Deferred issues

- None currently.

## Delivered result

- Default CLI and Make runs emit only transcript text on success.
- `--verbose` and `VERBOSE=1` restore progress, summary, ffmpeg, Whisper, and detailed failure diagnostics on stderr.
- Concise failures remain visible in quiet mode.
- Clipboard usage remains the direct `| pbcopy` pipeline.
- `make build-transcribe` produces a first-class standalone CLI with GNU-style `--verbose`/`-v` flags; `make transcribe` is documented as the local convenience wrapper.
- Direct CLI execution discovers the repository model relative to the current directory or executable, so it does not require a manual `WHISPER_MODEL_PATH` export.

## Verification evidence

- Targeted CLI and engine tests pass.
- Full non-CGO Go test suite passes.
- CGO transcribe binary builds successfully.
- Cached native transcription smoke test produced 593 identical transcript bytes in both modes, 0 stderr bytes in quiet mode, and 5,091 diagnostic bytes in verbose mode.
- Standalone CLI smoke test from `/tmp` discovered its executable-relative model and produced 593 transcript bytes, 0 stderr bytes, and exit code 0.

## Related ADRs

- [ADR 0006: Make CLI Diagnostics Opt-in](../../../doc/decisions/0006-quiet-cli-output-by-default.md)
