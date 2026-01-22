# Docker Usage

Docker-based workflows for OmniTranscripts.

## Quick Start

### Run the Service

```bash
docker run -d \
  --name omnitranscripts \
  -p 3000:3000 \
  -e API_KEY=your-api-key \
  -v $(pwd)/media:/media \
  wilmoore/omnitranscripts:latest
```

This mounts a local `./media` directory into the container for local file access.

### Transcribe a URL

```bash
curl -X POST http://localhost:3000/transcribe \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.youtube.com/shorts/VIDEO_ID"}'
```

### Transcribe a Local File

```bash
# Via file upload
curl -X POST http://localhost:3000/transcribe \
  -H "Authorization: Bearer your-api-key" \
  -F "file=@./media/video.mp4"
```

## Docker Compose

Use `docker-compose.yml` for a complete setup with persistent storage.

```bash
docker-compose up -d
```

## Examples

- `docker-compose.yml` - Full Docker Compose configuration
- `transcribe-url.sh` - Transcribe from URL via Docker
- `transcribe-local.sh` - Transcribe local file via Docker
