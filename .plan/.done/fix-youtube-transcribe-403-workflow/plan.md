# Fix YouTube Transcribe 403 Workflow

**Backlog:** 14, 15
**Branch:** `fix/youtube-transcribe-403-workflow`
**Status:** Completed
**Started:** 2026-08-26T13:07:16Z

## Reproduction

```bash
make transcribe URL='https://www.youtube.com/watch?v=hYRdmo7M-ng'
```

Expected: one progress announcement followed by a completed transcription.

Actual: two identical progress announcements, followed by a download-stage failure. Bundled `yt-dlp 2026.02.21` reports stale-version and YouTube SABR warnings, then an unavailable-format or HTTP 403 error.

Environment: macOS arm64, Go module declares Go 1.25, `go-ytdlp v1.3.1`, bundled `yt-dlp 2026.02.21`.

Frequency: reproducible for both reported URLs, `hYRdmo7M-ng` and `Gb8pA3xj-lU`.

Severity: high; URL transcription is blocked before normalization or transcription.

## Hypotheses

1. YouTube extractor drift has made the pinned downloader unable to select a usable audio format.
2. The Make target and CLI both own the same progress message.
3. The wrapper's raw failure advice is misleading because ADR 0004 prohibits self-updating or system fallback.

## Evidence

- `yt-dlp 2026.02.21 --test --no-playlist -f bestaudio` fails for `hYRdmo7M-ng` after the SABR warning with no requested format available.
- The same probe with the latest wrapper-pinned `yt-dlp 2026.07.04` succeeds, selects audio format 251, solves the JavaScript challenge through Deno, and completes its test download.
- The spike separately established the same old-fails/new-succeeds result for `Gb8pA3xj-lU`.
- The Makefile and `examples/transcribe/main.go` both print `Transcribing:`.
- ADR 0004 requires `DisableSystem: true` and pinned wrapper updates; ADR 0005 assigns progress and errors to stderr.

## Root Cause

`go-ytdlp v1.3.1` pinned `yt-dlp 2026.02.21`, whose YouTube clients could no longer obtain a usable full audio transfer for the reported videos. The initially proposed tagged wrapper v1.3.6 pins July 2026: its partial `--test` probe succeeded, but the exact full extraction still returned HTTP 403. The wrapper's current commit pins `yt-dlp 2026.08.19`, which completes both downloads. Separately, the Make target and CLI both printed the same progress line.

## Planned Fix

- Pin the exact `go-ytdlp` commit containing `yt-dlp 2026.08.19` while retaining the bundled executable and `DisableSystem: true`.
- Return structured downloader failures from the engine, including the pinned version, exit status, normalized diagnostic details, and compatibility classification.
- Suppress the managed binary's irrelevant self-update warning and give callers supported remediation.
- Make the CLI the sole progress owner and render compatibility failures concisely with a diagnostic details section.
- Add weekly Go module dependency monitoring.

## Regression Tests

- Downloader error classification and normalization.
- Error unwrapping and public fields.
- CLI output ownership, stdout/stderr separation, compatibility guidance, and generic error behavior.
- Network smoke tests for both reported YouTube URLs.

## Verification

- [x] `go test ./engine ./examples/transcribe`
- [x] `go test -short -timeout=90s ./...`
- [x] `go vet ./...`
- [x] `go build ./...`
- [x] `go build -o /tmp/omnitranscripts-transcribe examples/transcribe/main.go`
- [x] `go mod verify`
- [x] Backlog JSON validation and `git diff --check`
- [x] Full `yt-dlp 2026.02.21` reproduction fails; full `2026.08.19` extraction succeeds.
- [x] `make transcribe URL='https://www.youtube.com/watch?v=hYRdmo7M-ng'` completes download, normalization, and native Whisper transcription with one progress line.
- [x] `make transcribe URL='https://www.youtube.com/watch?v=Gb8pA3xj-lU'` completes download, normalization, and native Whisper transcription with one progress line.
