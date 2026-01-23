# Bug Investigation: yt-dlp Subprocess Leak

## Summary
OmniTranscripts leaks yt-dlp subprocesses because it passes `context.Background()` to all subprocess invocations, preventing proper process termination on cancellation, timeout, or error conditions.

## Root Cause Analysis

### Primary Issue: Non-cancellable Contexts
All yt-dlp invocations use `context.Background()`:

| File | Line | Usage |
|------|------|-------|
| `engine/engine.go` | 65 | `dl.Run(context.Background(), url, "--get-duration", "--no-warnings")` |
| `engine/engine.go` | 84 | `dl.Run(context.Background(), url)` |
| `lib/transcription.go` | 60 | `dl.Run(context.Background(), url)` |
| `lib/transcription.go` | 256 | `dl.Run(context.Background(), url, "--get-duration", "--no-warnings")` |

### Why This Causes Leaks
1. **go-ytdlp library** uses `exec.CommandContext(ctx, ...)` internally (command.go:240)
2. When context is cancelled, Go sends SIGKILL to the child process
3. `context.Background()` **never cancels**, so processes run indefinitely
4. HTTP request cancellation, timeouts, and errors don't propagate to subprocesses

### Secondary Issues
1. **No request context propagation**: `processTranscription()` doesn't accept or use a context
2. **Goroutines detached from request lifecycle**: `go processTranscriptionSync(job)` has no way to cancel
3. **No timeout on subprocess operations**: Individual yt-dlp calls can hang indefinitely

### Process Lifecycle Problem
```
HTTP Request -> Handler -> goroutine -> ProcessTranscription -> downloadAudio -> yt-dlp
                  |
                  v (request cancelled)
              Returns to client
                  |
                  X (no signal to yt-dlp subprocess)
```

## Affected Code Paths

### 1. Duration Check (`GetVideoDuration` / `GetMediaDuration`)
- Called synchronously in `handleURLTranscribe`
- If this hangs, the HTTP handler blocks
- Subprocess leak on client disconnect

### 2. Audio Download (`downloadAudio`)
- Called in background goroutine via `processTranscription`
- No mechanism to cancel on job error/timeout
- Most likely source of leaked processes (longest-running operation)

### 3. FFmpeg Normalization
- Also uses context (via ffmpeg-go library)
- Same issue but less severe (shorter runtime)

## Fix Strategy

### Phase 1: Context Propagation
1. Add context parameter to all transcription functions
2. Pass cancellable context from HTTP handlers
3. Use timeouts for individual operations

### Phase 2: Request Lifecycle Binding
1. Create context from Fiber request context
2. Cancel context when request terminates
3. Propagate cancellation to all child operations

### Phase 3: Observability (Optional)
1. Track active subprocess count
2. Log subprocess lifecycle events
3. Add metrics for monitoring

## Implementation Plan

### Files to Modify
1. `engine/engine.go` - Add context to `Transcribe()`, `GetMediaDuration()`, `downloadAudio()`
2. `lib/transcription.go` - Add context to `ProcessTranscription()`, `GetVideoDuration()`, `downloadAudio()`
3. `handlers/transcribe.go` - Create and propagate request-scoped contexts

### API Changes
```go
// Before
func Transcribe(url string, jobID string, opts Options) (*Result, error)
func ProcessTranscription(url, jobID string) (string, []models.Segment, error)

// After
func Transcribe(ctx context.Context, url string, jobID string, opts Options) (*Result, error)
func ProcessTranscription(ctx context.Context, url, jobID string) (string, []models.Segment, error)
```

### Timeout Strategy
- Duration check: 30 seconds (metadata only)
- Audio download: 10 minutes (configurable)
- Normalization: 5 minutes
- Transcription: 10 minutes

## Acceptance Criteria
- [ ] All yt-dlp subprocesses terminate on request cancellation
- [ ] All yt-dlp subprocesses terminate on timeout
- [ ] All yt-dlp subprocesses terminate on error conditions
- [ ] Process count remains stable under repeated usage
- [ ] No breaking changes to existing API contracts
