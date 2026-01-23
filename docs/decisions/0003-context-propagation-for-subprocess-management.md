# 0003. Context Propagation for Subprocess Management

Date: 2026-01-22

## Status

Accepted

## Context

OmniTranscripts spawns external subprocesses (yt-dlp, ffmpeg) for media download and audio normalization. These subprocesses were invoked with `context.Background()`, which never cancels. This caused subprocess leaks when:

- HTTP requests were cancelled by clients
- Request timeouts expired
- Processing errors occurred in earlier pipeline stages

The go-ytdlp library internally uses `exec.CommandContext(ctx, ...)`, which sends SIGKILL to child processes when the context is cancelled. By passing `context.Background()`, we bypassed this cleanup mechanism entirely.

## Decision

Add explicit context propagation throughout the transcription pipeline:

1. **API Signature Changes**: Add `ctx context.Context` as the first parameter to all transcription functions:
   - `engine.Transcribe(ctx, url, jobID, opts)`
   - `engine.GetMediaDuration(ctx, url)`
   - `lib.ProcessTranscription(ctx, url, jobID)`
   - `lib.GetVideoDuration(ctx, url)`

2. **Request-Scoped Contexts**: Create contexts from HTTP request lifecycle with appropriate timeouts:
   - Duration check: 30 seconds
   - Sync processing (≤2 min media): 5 minutes
   - Async processing (>2 min media): 30 minutes

3. **Cancellation Propagation**: Pass contexts through to all subprocess invocations, enabling automatic cleanup when contexts are cancelled.

## Consequences

### Positive

- Subprocesses terminate automatically on request cancellation
- Subprocesses terminate automatically on timeout expiry
- System resources are properly released after errors
- No more process accumulation under normal operation
- Predictable resource usage under load

### Negative

- Breaking API change for library consumers (context parameter required)
- Existing code must be updated to pass contexts
- Callers must understand context lifecycle implications

## Alternatives Considered

### 1. Process Group Management

Track and kill process groups manually. Rejected because:
- More complex implementation
- Platform-specific behavior differences
- Context-based approach is idiomatic Go

### 2. Background Process Reaper

Periodic cleanup of orphaned processes. Rejected because:
- Reactive rather than preventive
- May kill processes still in use
- Adds operational complexity

### 3. Process Pool with Limits

Limit concurrent subprocess count. Rejected because:
- Doesn't address the root cause (leak)
- May reject valid requests under load
- Complexity without solving the fundamental issue

## Related

- Planning: `.plan/.done/fix-yt-dlp-subprocess-leak/`
- Investigation: `.plan/.done/fix-yt-dlp-subprocess-leak/investigation.md`
- Implementation: `.plan/.done/fix-yt-dlp-subprocess-leak/implementation.md`
