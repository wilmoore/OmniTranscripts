# Bug Fix: Local file handling in CLI and engine

## Status: COMPLETE

## Bug Details

**Steps to Reproduce:**
```bash
make transcribe URL='file://Users/wilmooreiii/Downloads/video.mp4'
```

**Expected:** Transcribes the local file directly without downloading

**Actual (Bug 1 - CLI):**
```
Error: File not found: file://Users/wilmooreiii/Downloads/video.mp4
```

**Actual (Bug 2 - Engine):**
```
Transcription failed at stage 'download': failed to download audio
  Cause: yt-dlp failed: exit code 1
ERROR: [generic] '/Users/.../video.mp4' is not a valid URL
```

**Severity:** High (blocks local file transcription entirely)

## Root Causes

### Bug 1: CLI doesn't handle file:// URIs
In `examples/transcribe/main.go:32`, URL detection only checks for `http://` and `https://`.
When `file://` URI is passed, it's treated as a literal filesystem path.

### Bug 2: Engine always uses yt-dlp
In `engine/engine.go`, the `Transcribe()` function always calls `downloadAudio()` which uses yt-dlp.
It doesn't detect whether the input is already a local file.

## Fixes

### Fix 1: CLI file:// URI handling
```go
// Handle file:// URIs by converting to local path
if strings.HasPrefix(input, "file://") {
    input = strings.TrimPrefix(input, "file://")
    if !strings.HasPrefix(input, "/") {
        input = "/" + input
    }
}
```

### Fix 2: Engine local file detection
```go
func isLocalFile(input string) bool {
    if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
        return false
    }
    _, err := os.Stat(input)
    return err == nil
}

// In Transcribe():
if isLocalFile(url) {
    // Local file: skip download, normalize directly
    if err := normalizeAudio(ctx, url, normalizedAudio); err != nil {
        return nil, NewError(StageNormalize, "failed to normalize audio", err)
    }
} else {
    // URL: download via yt-dlp, then normalize
    ...
}
```

## Verification

```bash
# Before fix
$ make transcribe URL='file://Users/.../video.mp4'
ERROR: [generic] '/Users/.../video.mp4' is not a valid URL

# After fix
$ make transcribe URL='file://Users/.../video.mp4'
Transcribing: /Users/.../video.mp4
Type: Local file
Processing local file: /Users/.../video.mp4
--- Transcript ---
...
```

## Files Modified

1. `examples/transcribe/main.go` - Add file:// URI handling
2. `engine/engine.go` - Add isLocalFile() and skip download for local files
