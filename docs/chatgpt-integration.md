# ChatGPT Integration Guide

OmniTranscripts includes an MCP (Model Context Protocol) server that enables integration with ChatGPT via the OpenAI Apps SDK. This allows users to transcribe videos and audio directly from ChatGPT conversations.

## Overview

The MCP server exposes two tools:

| Tool | Description |
|------|-------------|
| `transcribe_url` | Start transcription of a video/audio URL |
| `get_transcription` | Check status and retrieve transcript |

## Quick Start

### 1. Enable the MCP Server

The MCP server is enabled by default. Configure via environment variables:

```bash
# .env
MCP_ENABLED=true        # Enable/disable MCP server (default: true)
MCP_ENDPOINT=/mcp       # Endpoint path (default: /mcp)
API_KEY=your-api-key    # Required for authentication
```

### 2. Start the Server

```bash
make dev
# Output includes: MCP server enabled at /mcp
```

### 3. Expose to the Internet

For ChatGPT to reach your server, it needs a public URL. Options:

**Development (ngrok):**
```bash
ngrok http 3000
# Use the https URL provided
```

**Production:**
Deploy to a cloud provider with a public domain and TLS.

### 4. Register with ChatGPT

1. Go to [ChatGPT Developer Settings](https://chat.openai.com/settings/developer)
2. Create a new App
3. Configure the MCP server URL: `https://your-domain.com/mcp`
4. Add the API key in the authentication settings

## MCP Tools Reference

### transcribe_url

Starts transcription of media from a URL.

**Input:**
```json
{
  "url": "https://youtube.com/watch?v=abc123"
}
```

**Output:**
```
Job ID: 550e8400-e29b-41d4-a716-446655440000
Status: processing
Duration: 3600 seconds (60 minutes)

Transcription started. For a 60-minute video, expect 10-15 minutes processing time. Use get_transcription with job_id to check status.
```

### get_transcription

Retrieves the status and result of a transcription job.

**Input:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Output (processing):**
```
Job ID: 550e8400-e29b-41d4-a716-446655440000
Status: running
Progress: Processing (downloading, normalizing, or transcribing)

Still processing. Check again in 1-2 minutes.
```

**Output (complete):**
```
Job ID: 550e8400-e29b-41d4-a716-446655440000
Status: complete

Word Count: 15420

--- TRANSCRIPT ---
[Full transcript text here...]
--- END TRANSCRIPT ---

Segments: 342 timestamped segments available

Transcription complete.
```

## Example Conversation

Here's how a typical interaction looks in ChatGPT:

```
User: Can you summarize this podcast for me?
      https://youtube.com/watch?v=abc123

ChatGPT: I'll transcribe that podcast for you. Let me start the process...

[Calls transcribe_url]

ChatGPT: I've started transcribing a 1 hour 23 minute video. This will take
         about 10-15 minutes. Let me check the progress...

[Waits ~2 minutes, calls get_transcription]

ChatGPT: Still processing - currently downloading and transcribing the audio.
         I'll check again shortly...

[Waits ~2 minutes, calls get_transcription]

ChatGPT: The transcription is complete! Here's a summary of the podcast:

         **Key Topics Discussed:**
         1. Introduction to AI agents
         2. Model Context Protocol explained
         3. Practical implementation examples

         **Notable Quotes:**
         - "MCP enables seamless integration between AI systems..."

         Would you like me to focus on any particular section?
```

## Authentication

The MCP server uses API key authentication. Include the key in requests:

**Header options:**
```
Authorization: Bearer your-api-key
```
or
```
X-API-Key: your-api-key
```

## Processing Times

Estimated processing times based on media duration:

| Duration | Processing Time |
|----------|-----------------|
| ≤2 min   | 1-2 minutes |
| 2-10 min | 2-5 minutes |
| 10-30 min | 5-10 minutes |
| 30-60 min | 10-15 minutes |
| >60 min | ~1 min per 5 min of content |

## Supported Platforms

OmniTranscripts supports 1000+ platforms via yt-dlp, including:

- YouTube
- Vimeo
- SoundCloud
- Spotify (podcasts)
- Twitter/X
- TikTok
- Direct audio/video URLs
- And many more...

## Troubleshooting

### "Invalid or missing API key"

Ensure your API key is correctly configured in both:
1. Your `.env` file (`API_KEY=...`)
2. ChatGPT app settings

### Job stuck in "processing"

Long videos may take time. If a job seems stuck:
1. Check server logs for errors
2. Verify yt-dlp and ffmpeg are working
3. Check if the video is accessible (not private/deleted)

### "Download failed"

Common causes:
- Video is private or age-restricted
- URL is invalid or expired
- Platform not supported
- Rate limiting by the source platform

## Security Considerations

1. **API Key**: Keep your API key secure. Don't share it publicly.
2. **Rate Limiting**: Consider implementing rate limits for production.
3. **Content**: The server processes whatever URLs are submitted. Be aware of content policies.

## Architecture

```
ChatGPT → MCP Protocol (HTTP) → OmniTranscripts (:3000/mcp)
                                       │
                                       ├── transcribe_url
                                       │     └── Creates job, starts processing
                                       │
                                       └── get_transcription
                                             └── Returns job status/transcript
                                       │
                                       ▼
                              jobs.Queue (shared with HTTP API)
                                       │
                                       ▼
                              lib.ProcessTranscription
                                       │
                              ┌────────┴────────┐
                              ▼                 ▼
                           yt-dlp           ffmpeg
                          (download)     (normalize)
                                       │
                                       ▼
                              whisper.cpp / AssemblyAI
                                 (transcribe)
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_ENABLED` | `true` | Enable/disable MCP server |
| `MCP_ENDPOINT` | `/mcp` | MCP server endpoint path |
| `API_KEY` | `your-api-key-here` | API key for authentication |
| `PORT` | `3000` | Server port |

## Related Documentation

- [API Documentation](api.md) - Full HTTP API reference
- [Architecture](architecture.md) - System design details
- [Deployment](deployment.md) - Production deployment guide
