# Bug Fix: YouTube Shorts 403 Forbidden Download Error

## Bug Details

**Severity:** Critical (blocks work)
**Branch:** `fix/yt-dlp-youtube-shorts-403`
**Backlog ID:** 11

### Steps to Reproduce
```bash
make transcribe URL='https://www.youtube.com/shorts/sMQRVvc4-9c'
```

### Expected Behavior
Audio should download from YouTube Shorts and be transcribed successfully.

### Actual Behavior
```
ERROR: unable to download video data: HTTP Error 403: Forbidden
WARNING: [youtube] sMQRVvc4-9c: Some web client https formats have been skipped as they
are missing a url. YouTube is forcing SABR streaming for this client.
See https://github.com/yt-dlp/yt-dlp/issues/12482 for more details
```

### Environment
- go-ytdlp: v1.2.7 (bundles yt-dlp ~2025.09.05)
- System yt-dlp: 2025.12.08
- Platform: darwin (macOS)

---

## Root Cause Analysis

### Investigation Summary

1. **GitHub Issue #12482**: YouTube rolled out changes to their web client that removed traditional HTTPS format options for `adaptiveFormats`, replacing them entirely with SABR (Segmented Adaptive Bitrate) streaming protocol.

2. **SABR Protocol**: A custom YouTube streaming protocol that yt-dlp doesn't fully support yet. The issue is still open, though progress is being made on a SABR downloader.

3. **Bundled yt-dlp Version**: go-ytdlp v1.2.7 bundles an older yt-dlp version (~2025.09.05) that predates YouTube's SABR changes.

4. **Available Fix**: go-ytdlp v1.3.1 (released 2026-02-22) bundles yt-dlp 2026.02.21, which has improved YouTube client handling that works around SABR limitations.

### Related ADR

**ADR-0004: yt-dlp Dependency Management Strategy**
- Decision: Use bundled yt-dlp from go-ytdlp and keep the library updated
- Guidance: "When platform downloads fail, check if a go-ytdlp update is available before investigating other causes"
- This fix follows the established ADR guidance

---

## Implementation Plan

### Step 1: Update go-ytdlp dependency
```bash
go get github.com/lrstanley/go-ytdlp@v1.3.1
go mod tidy
```

### Step 2: Verify the fix
```bash
make transcribe URL='https://www.youtube.com/shorts/sMQRVvc4-9c'
```

### Step 3: Test additional YouTube URLs
- Regular YouTube video
- YouTube Shorts (the failing case)
- YouTube playlist entry

### Step 4: Run existing tests
```bash
make test-short
```

---

## Implementation Details

### Changes Made

1. **Updated go-ytdlp dependency** (`go.mod`)
   - v1.2.7 → v1.3.1
   - Bundles yt-dlp 2026.02.21 (was ~2025.09.05)

2. **Added bundled yt-dlp initialization** (`engine/engine.go:16-32`)
   - Added `ensureYtdlp()` function with `sync.Once` for lazy initialization
   - Calls `ytdlp.Install(ctx, &ytdlp.InstallOptions{DisableSystem: true})`
   - **No system fallback** - only uses bundled binary (per ADR-0004)
   - Returns error if bundled install fails

3. **Called ensureYtdlp before yt-dlp usage with error handling**
   - Added call in `GetMediaDuration()` - returns error if install fails
   - Added call in `downloadAudio()` - returns error if install fails

### Root Cause
The previous code used `ytdlp.New()` without calling `Install()`, which caused go-ytdlp
to fall back to the system PATH yt-dlp (version 2025.12.08). This older version doesn't
handle YouTube's new SABR streaming protocol, resulting in 403 errors on YouTube Shorts.

### Why It Works Now
Calling `ytdlp.Install()` with `DisableSystem: true` downloads the bundled yt-dlp binary
(2026.02.21) to `~/Library/Caches/go-ytdlp/` and **exclusively** uses it. This ensures:
- Consistent behavior across all environments
- No dependency on system-installed yt-dlp
- Access to latest YouTube extractor fixes

---

## Verification Checklist

- [x] YouTube Shorts URL downloads successfully
- [x] Regular YouTube URLs still work
- [x] `make test-short` passes
- [x] No regressions in other platform downloads

---

## Related Issues for Backlog

None identified - this is a straightforward dependency update per ADR-0004.
