# OmniTranscripts Examples

Real-world usage examples for OmniTranscripts.

## Directory Structure

| Directory | Description |
|-----------|-------------|
| [local-files/](local-files/) | Transcribing local audio/video files |
| [short-form/](short-form/) | YouTube Shorts, TikTok, Instagram Reels |
| [docker/](docker/) | Docker-based workflows |
| [production/](production/) | Async jobs, webhooks, polling patterns |

## Prerequisites

All examples assume:
- OmniTranscripts is running on `http://localhost:3000`
- You have a valid `API_KEY` set

Set your API key:
```bash
export API_KEY="your-api-key-here"
```

## Quick Reference

### Local File Upload
```bash
curl -X POST http://localhost:3000/transcribe \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@./video.mp4"
```

### URL Transcription
```bash
curl -X POST http://localhost:3000/transcribe \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/audio.mp3"}'
```

### Check Job Status
```bash
curl -X GET "http://localhost:3000/transcribe/$JOB_ID" \
  -H "Authorization: Bearer $API_KEY"
```
