# 004. Async-First MCP Server Integration

Date: 2026-02-04

## Status

Accepted

## Context

We needed to add ChatGPT integration to OmniTranscripts via the Model Context Protocol (MCP). The primary use case is allowing users to submit YouTube URLs (and other media) to get transcripts for summarization. The challenge is that long-form content like podcasts and lectures—the core use case—can take significant time to process, and synchronous approaches would timeout.

Key considerations:
- Videos can range from 2 minutes to 2+ hours
- Processing includes download, audio normalization, and transcription
- ChatGPT has request timeout limits
- Users expect to summarize content "too long to watch quickly"

## Decision

We implemented an **async-first** MCP server with two tools:

1. **`transcribe_url`** - Accepts a URL, creates a background job, and immediately returns a `job_id`
2. **`get_transcription`** - Polls job status and returns results when complete

This approach allows ChatGPT to:
1. Start transcription and immediately acknowledge the request
2. Periodically poll for completion with natural conversation updates
3. Handle videos of any length without timeout issues

Additional decisions:
- **MCP Library**: Selected `mark3labs/mcp-go` for HTTP transport support and active maintenance
- **Port**: Mounted at `:3000/mcp` alongside existing REST API (simpler deployment)
- **Authentication**: Reuses existing `API_KEY` environment variable
- **Encore Removal**: Removed Encore dependency in favor of standard Go tooling (`CGO_ENABLED=0`)

## Consequences

### Positive
- Videos of any length can be processed without timeout issues
- ChatGPT can provide natural progress updates during long transcriptions
- Simple two-tool interface is easy for ChatGPT to understand and use
- Shared job queue means MCP and REST API jobs are unified
- No vendor lock-in after removing Encore

### Negative
- Requires polling (ChatGPT must call `get_transcription` periodically)
- No real-time progress streaming (future enhancement)
- Users may need to wait several minutes for long content

### Neutral
- Processing times scale linearly with content duration (~1 min per 5 min of video)

## Alternatives Considered

### Sync-first with async fallback
Would handle short videos quickly but still timeout on long content—the primary use case. Rejected because it doesn't solve the core problem.

### Streaming/webhook approach
Would provide real-time updates but MCP protocol doesn't support server-initiated callbacks. Would require additional complexity. Deferred for future consideration.

### Encore for infrastructure
Evaluated Encore for rate limiting, observability, and secrets management. Rejected because:
- Rate limiting still requires custom implementation
- Vendor lock-in to Encore Cloud
- Standard Go tooling provides sufficient capability for current needs

## Related

- Planning: `.plan/.done/feat-mcp-server-chatgpt-integration/`
- Documentation: `docs/chatgpt-integration.md`
