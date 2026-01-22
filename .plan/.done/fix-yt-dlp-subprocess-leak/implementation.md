# Implementation Summary: yt-dlp Subprocess Leak Fix

## Changes Made

### Core Fix: Context Propagation

The root cause was that all yt-dlp and ffmpeg subprocess invocations used `context.Background()`, which never cancels. This prevented subprocess termination on request cancellation, timeout, or error conditions.

### Files Modified

| File | Changes |
|------|---------|
| `engine/engine.go` | Added `ctx context.Context` param to `Transcribe()`, `GetMediaDuration()`, `downloadAudio()`, `normalizeAudio()` |
| `lib/transcription.go` | Added `ctx context.Context` param to `ProcessTranscription()`, `GetVideoDuration()`, `downloadAudio()`, `normalizeAudio()` |
| `handlers/transcribe.go` | Created request-scoped contexts with timeouts, propagated to all processing functions |
| `transcribe/service.go` | Updated Encore service to use request context for cancellation |
| `examples/local-files/transcribe.go` | Updated example to use context |
| `engine/doc.go` | Updated documentation with new API signature |
| `lib/transcription_bench_test.go` | Updated benchmark to use context |

### Timeout Strategy

| Operation | Timeout | Rationale |
|-----------|---------|-----------|
| Duration check | 30 seconds | Metadata-only operation |
| Sync processing (<=2 min media) | 5 minutes | Short media should complete quickly |
| Async processing (>2 min media) | 30 minutes | Allow for longer downloads/transcriptions |
| File uploads | 30 minutes | Unknown duration, use conservative timeout |

### API Changes

```go
// Before
func Transcribe(url string, jobID string, opts Options) (*Result, error)
func ProcessTranscription(url, jobID string) (string, []models.Segment, error)
func GetVideoDuration(url string) (int, error)
func GetMediaDuration(url string) (int, error)

// After
func Transcribe(ctx context.Context, url string, jobID string, opts Options) (*Result, error)
func ProcessTranscription(ctx context.Context, url, jobID string) (string, []models.Segment, error)
func GetVideoDuration(ctx context.Context, url string) (int, error)
func GetMediaDuration(ctx context.Context, url string) (int, error)
```

### How Context Cancellation Works

1. `go-ytdlp` library internally uses `exec.CommandContext(ctx, ...)`
2. `ffmpeg-go` library uses `stream.Context` field
3. When context is cancelled, Go sends SIGKILL to child processes
4. This ensures all subprocesses terminate when:
   - HTTP request is cancelled
   - Timeout expires
   - Parent goroutine exits

## Testing

- All existing tests pass
- Build compiles without errors
- API contracts preserved (context added as first parameter)

## Verification Steps

1. Start OmniTranscripts in development mode
2. Submit multiple transcription requests
3. Monitor `yt-dlp` process count via Activity Monitor
4. Verify processes terminate after request completion
5. Verify processes terminate on timeout/cancellation

## Future Improvements (Optional)

1. Add observability metrics for active subprocess count
2. Consider process group management for cleanup guarantees
3. Add configurable timeouts via environment variables
