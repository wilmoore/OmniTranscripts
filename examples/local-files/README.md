# Local File Transcription

Examples for transcribing local audio and video files.

## Supported Formats

OmniTranscripts supports any format that FFmpeg can process:
- Video: `.mp4`, `.mkv`, `.avi`, `.mov`, `.webm`
- Audio: `.mp3`, `.wav`, `.m4a`, `.flac`, `.ogg`

## File Upload via HTTP API

```bash
# Upload and transcribe a local file
curl -X POST http://localhost:3000/transcribe \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@./recording.mp4"
```

## Go Library Usage

See `transcribe.go` for a complete example of using the Go library to transcribe local files.

```go
import "omnitranscripts/engine"

result, err := engine.Transcribe(
    "/path/to/recording.mp4",
    "job-001",
    engine.DefaultOptions(),
)
```

Local files bypass the download stage and go directly through FFmpeg to Whisper.

## Examples

- `transcribe.go` - Go library example for local file transcription
- `upload.sh` - cURL file upload example
