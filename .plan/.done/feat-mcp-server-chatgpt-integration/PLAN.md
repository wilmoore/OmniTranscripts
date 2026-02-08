# Feature: MCP Server + ChatGPT Apps SDK Integration

## Related ADRs

- [004. Async-First MCP Server Integration](../../../doc/decisions/0004-async-first-mcp-server-integration.md)

## Overview

Add an MCP (Model Context Protocol) server to OmniTranscripts to enable ChatGPT integration via the OpenAI Apps SDK, allowing users to submit YouTube URLs and get transcripts for summarization—including long-form content like podcasts and lectures.

## Confirmed Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| MCP Library | `mark3labs/mcp-go` | HTTP transport, high quality, active development |
| Port Configuration | Same port (`:3000/mcp`) | Simpler deployment, single service |
| Processing Model | **Async-first** | Handles any video length without timeout issues |
| Authentication | Same API key as HTTP API | Consistent auth, reuses existing config |

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         ChatGPT                                 │
│  User: "Summarize this 2-hour podcast: youtube.com/watch?v=..." │
└───────────────────────────┬─────────────────────────────────────┘
                            │ MCP Protocol (HTTP)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│              OmniTranscripts Server (:3000)                     │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ /mcp (MCP Server)                                           ││
│  │   ├── Tool: transcribe_url → starts job, returns job_id    ││
│  │   └── Tool: get_transcription → returns status/transcript  ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ /transcribe, /health (Existing HTTP API)                   ││
│  └─────────────────────────────────────────────────────────────┘│
└───────────────────────────┬─────────────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
┌──────────────────────┐    ┌──────────────────────┐
│   jobs.Queue         │    │   engine.Transcribe  │
│   (existing)         │    │   (existing)         │
└──────────────────────┘    └──────────────────────┘
```

## MCP Tools Specification

### Tool 1: `transcribe_url`

Starts transcription of media from a URL. Returns immediately with job info.

**Input Schema:**
```json
{
  "url": {
    "type": "string",
    "description": "URL of media to transcribe (YouTube, Vimeo, podcast, etc.)",
    "required": true
  }
}
```

**Output:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "title": "Conference Talk: Building Scalable Systems",
  "duration_seconds": 3600,
  "message": "Transcription started. For a 60-minute video, expect 5-10 minutes processing time. Use get_transcription to check status."
}
```

**Tool Description (guides ChatGPT behavior):**
> "Starts transcription of a video or audio URL. Returns a job_id immediately. Processing happens in the background. For videos over 2 minutes, wait 1-2 minutes then call get_transcription to check if complete. Longer videos take proportionally longer. Supports YouTube, Vimeo, SoundCloud, direct media URLs, and 1000+ platforms."

### Tool 2: `get_transcription`

Retrieves the status and result of a transcription job.

**Input Schema:**
```json
{
  "job_id": {
    "type": "string",
    "description": "Job ID returned from transcribe_url",
    "required": true
  }
}
```

**Output (processing):**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "progress": "Transcribing audio (stage 3/3)",
  "message": "Still processing. Check again in 1-2 minutes."
}
```

**Output (complete):**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "complete",
  "transcript": "Full transcript text here...",
  "segments": [
    {"start": 0.0, "end": 5.2, "text": "Welcome to the show..."},
    {"start": 5.2, "end": 12.8, "text": "Today we're discussing..."}
  ],
  "word_count": 15420,
  "duration_seconds": 3600
}
```

**Output (error):**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "error",
  "error": "Download failed: Video is private or unavailable"
}
```

**Tool Description:**
> "Check the status of a transcription job. Returns 'processing' if still running, 'complete' with the full transcript, or 'error' with details. Call this after transcribe_url to retrieve results."

## Implementation Plan

### Phase 1: MCP Server Foundation
1. Add `github.com/mark3labs/mcp-go` dependency
2. Create `mcp/` package structure:
   - `mcp/server.go` - MCP server setup and configuration
   - `mcp/tools.go` - Tool definitions and handlers
   - `mcp/types.go` - Request/response types
3. Implement `transcribe_url` tool handler
4. Implement `get_transcription` tool handler
5. Add authentication middleware (reuse API key validation)

### Phase 2: Integration with Main Server
1. Mount MCP server at `/mcp` endpoint in `main.go`
2. Share job queue between HTTP API and MCP server
3. Add MCP-specific config options to `config/config.go`:
   - `MCP_ENABLED` (default: true)
   - `MCP_ENDPOINT` (default: "/mcp")
4. Update CORS configuration for ChatGPT domains

### Phase 3: Enhanced Job Status
1. Add progress tracking to job processing:
   - Stage indicators: "downloading", "normalizing", "transcribing"
   - Percentage or step count where possible
2. Store media metadata (title, duration) in job record
3. Add word count to completed transcripts

### Phase 4: Testing & Documentation
1. Add unit tests for MCP tool handlers
2. Add integration test with mock MCP client
3. Update `docs/api.md` with MCP server documentation
4. Create `docs/chatgpt-integration.md` setup guide
5. Add example conversation flow to README

### Phase 5: ChatGPT App Registration
1. Create app manifest for ChatGPT developer portal
2. Configure OAuth (if required) or API key auth
3. Test end-to-end with real ChatGPT integration
4. Document app submission process

## File Changes Summary

| File | Change |
|------|--------|
| `go.mod` | Add `github.com/mark3labs/mcp-go` |
| `mcp/server.go` | NEW - MCP server setup |
| `mcp/tools.go` | NEW - Tool definitions and handlers |
| `mcp/types.go` | NEW - MCP-specific types |
| `main.go` | Mount MCP server at `/mcp` |
| `config/config.go` | Add MCP config options |
| `jobs/job.go` | Add progress field, metadata |
| `lib/transcription.go` | Emit progress updates |
| `docs/api.md` | Document MCP tools |
| `docs/chatgpt-integration.md` | NEW - Setup guide |

## Dependencies

- `github.com/mark3labs/mcp-go` - MCP protocol implementation
- Existing `engine/` package - transcription functionality
- Existing `jobs/` package - job queue

## Success Criteria

- [ ] MCP server starts and responds to tool discovery at `/mcp`
- [ ] `transcribe_url` creates job and returns job_id
- [ ] `get_transcription` returns status and transcript
- [ ] Authentication validates API key
- [ ] Long videos (1+ hour) transcribe successfully
- [ ] ChatGPT can invoke tools and retrieve transcripts
- [ ] Progress updates visible during processing
- [ ] Documentation complete with setup instructions
- [ ] Tests verify MCP tool handlers

## Example ChatGPT Conversation Flow

```
User: Can you summarize this podcast for me?
      https://youtube.com/watch?v=abc123

ChatGPT: I'll transcribe that podcast for you. Let me start the process...

[Calls transcribe_url]

ChatGPT: I've started transcribing "Tech Talk Episode 42" (1 hour 23 minutes).
         This will take about 8-12 minutes. Let me check the progress...

[Waits ~2 minutes, calls get_transcription]

ChatGPT: Still processing - currently transcribing the audio.
         I'll check again shortly...

[Waits ~2 minutes, calls get_transcription]

ChatGPT: The transcription is complete! Here's a summary of the podcast:

         **Key Topics Discussed:**
         1. ...
         2. ...

         **Notable Quotes:**
         - "..."

         Would you like me to focus on any particular section?
```

## Open Items / Future Enhancements

- [ ] Streaming transcript chunks as they complete (for very long content)
- [ ] Cancel job tool (if user abandons request)
- [ ] Webhook notifications (if MCP supports callbacks)
- [ ] Rate limiting per ChatGPT user session
