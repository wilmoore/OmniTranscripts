# Production Patterns

Examples for production usage: async job handling, webhooks, and polling patterns.

## Async Processing

Media longer than 2 minutes is processed asynchronously. The API returns a job ID immediately, and you poll for results.

### Flow

1. Submit transcription request → receive `job_id`
2. Poll `GET /transcribe/{job_id}` until `status` is `complete` or `error`
3. Retrieve transcript from response

### Job States

| Status | Description |
|--------|-------------|
| `pending` | Job queued, waiting to start |
| `running` | Transcription in progress |
| `complete` | Done, transcript available |
| `error` | Failed, check `error` field |

## Webhooks

Configure a webhook URL to receive notifications when jobs complete.

```json
{
  "url": "https://example.com/long-video.mp4",
  "webhook_url": "https://your-server.com/webhook"
}
```

## Examples

- `async-job-polling.sh` - Poll for async job completion
- `webhook-server.go` - Simple webhook receiver
- `webhook-payload.json` - Example webhook payload
