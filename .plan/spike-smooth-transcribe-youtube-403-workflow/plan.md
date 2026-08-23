# Spike: Smooth YouTube Transcribe Failure Workflow

**Branch:** `spike/smooth-transcribe-youtube-403-workflow`
**Status:** Completed — Promote
**Timebox:** 90 minutes
**Started:** 2026-08-23

## Uncertainty to Reduce

What is the smallest reliable change that makes `make transcribe URL=...` feel coherent when the bundled `yt-dlp` is stale or YouTube rejects a media format with HTTP 403?

## Success Criteria

- Identify every source of duplicate CLI status output.
- Verify which `yt-dlp` binary/version the Go wrapper executes.
- Reproduce the supplied URL failure and test current credible download strategies.
- Determine whether automatic updates, controlled fallbacks, preflight checks, or better diagnostics are appropriate under ADR 0004.
- Recommend a bounded implementation with testable acceptance criteria.

## Options

1. Improve messaging only and tell users how to update dependencies.
2. Refresh the bundled dependency and retain strict pinning.
3. Add a controlled system-binary or freshly installed fallback after a recognized stale/403 failure.
4. Add an explicit dependency-management command or preflight that updates/verifies `yt-dlp` before transcription.
5. Combine a quiet Make target, structured failure classification, and one bounded retry path.

## Evidence Needed

- Make target and CLI output ownership.
- ADR 0004 constraints and rationale.
- Wrapper installation/cache behavior and executable path.
- Installed, bundled, and current upstream `yt-dlp` versions.
- Results for the supplied URL across relevant client/format/update variants.
- Operational, security, determinism, and testability tradeoffs.

## Related ADR Review

- ADR 0004 governs `yt-dlp` dependency management and currently forbids system fallback.
- ADR 0005 governs stdout/stderr separation for CLI workflows.

## Investigation Log

- 2026-08-23: Created the mandatory spike branch and confirmed the worktree was clean.
- 2026-08-23: Located duplicate status ownership: Makefile line 94 and `examples/transcribe/main.go` line 39 both emit `Transcribing:`.
- 2026-08-23: Confirmed `engine.ensureYtdlp` uses `DisableSystem: true` in accordance with ADR 0004.
- 2026-08-23: Read ADR 0004 and ADR 0005. ADR 0004 explicitly rejects system fallback and runtime auto-update; ADR 0005 assigns human diagnostics to stderr and transcript data to stdout.
- 2026-08-23: Confirmed `go-ytdlp v1.3.1` pins and checksum-verifies `yt-dlp 2026.02.21` in `~/Library/Caches/go-ytdlp/`; the system binary is even older at `2025.12.08`.
- 2026-08-23: Upstream `yt-dlp` latest stable is `2026.08.19`. The latest available `go-ytdlp`, v1.3.6, pins `yt-dlp 2026.07.04` and requires Go 1.25 or newer; this repository already declares Go 1.25.
- 2026-08-23: Checksum-verified the official `yt-dlp_macos` 2026.08.19 release asset (`0f192b7ec147ab6288885d6351d9ab67367640029b4377576ef46dd79cf7b202`).
- 2026-08-23: `yt-dlp 2026.02.21 --test --no-playlist -f bestaudio` failed for `Gb8pA3xj-lU` after the SABR warning; the same command with `2026.08.19` selected format `251-20` and completed.
- 2026-08-23: A throwaway wrapper probe using `go-ytdlp v1.3.6`, `DisableSystem: true`, and pinned `yt-dlp 2026.07.04` also completed the supplied URL test successfully. No production prototype code was added.

## Findings

1. **The immediate failure is dependency drift, not the transcription layer.** The exact URL fails before FFmpeg or Whisper. A current upstream binary and the latest pinned wrapper binary both download it successfully.
2. **The existing upgrade policy works but is not being exercised frequently enough.** ADR 0004 anticipated precisely this failure mode and says to update `go-ytdlp`; v1.3.6 is available and remains reproducible.
3. **System fallback would not help this machine and would violate ADR 0004.** The installed system `yt-dlp` is `2025.12.08`, older than the bundled binary.
4. **The upstream `yt-dlp -U` advice is misleading here.** OmniTranscripts deliberately executes a versioned cached binary managed by `go-ytdlp`; users should not mutate it independently.
5. **Output duplication is a simple ownership bug.** The Make target and CLI both announce the same operation. ADR 0005 makes the CLI the natural owner of diagnostics, so the Makefile echo should be removed.
6. **Automatic runtime update is the wrong smoothing mechanism.** It breaks reproducibility, adds a network/write dependency at runtime, and contradicts ADR 0004. A recognized error should instead explain that the bundled downloader is stale and point maintainers/users to the supported dependency-refresh path.
7. **A blind retry is low value.** Retrying the same pinned executable and arguments cannot repair extractor drift. A retry should only exist if it changes a controlled, tested input; current evidence does not require one after the dependency refresh.

## Recommendation

**Promote** to follow-up backlog item 14.

Implement one bounded change set:

- Update the pinned dependency from `go-ytdlp v1.3.1` to `v1.3.6`, retaining `DisableSystem: true`.
- Remove the Makefile `Transcribing:` echo; let the CLI own stderr diagnostics.
- Preserve raw downloader details for debugging, but recognize stale-version, SABR, requested-format, and HTTP-403 signatures and lead with one concise explanation: the bundled downloader could not obtain a usable YouTube format, identify its version, and direct maintainers to update the pinned `go-ytdlp` dependency. Do not recommend `yt-dlp -U` for the managed binary.
- Add pure, table-driven tests for failure classification and user-facing guidance. Add a command/output test proving one status announcement and clean stdout.
- Add dependency-update automation or a documented scheduled `go list -m -u` check so the wrapper does not remain stale for months.
- Verify with unit tests, build/vet, stdout/stderr capture, and a separately marked network smoke test for the supplied URL.

Do not add system fallback or automatic runtime self-update unless ADR 0004 is explicitly superseded with deployment and supply-chain evidence.

## Remaining Uncertainty

- How quickly `go-ytdlp` will publish a wrapper release for upstream `2026.08.19`; v1.3.6 currently lags at `2026.07.04` but passes this URL.
- Whether YouTube will later require PO tokens, cookies, or a JavaScript runtime for this or other videos; no single downloader upgrade can guarantee future platform behavior.
- Which dependency automation cadence best balances extractor freshness against reproducible release verification.
