# Add CLI Transcribe Target

**Branch:** `refactor/add-cli-transcribe-target`
**Status:** Planning
**Backlog ID:** 10

## Summary

Add `make transcribe URL="..."` for quick one-off transcriptions without starting the HTTP server.

## Implementation

1. Move `examples/local-files/transcribe.go` → `examples/transcribe/main.go`
2. Remove file-exists check (URLs work via yt-dlp)
3. Add Makefile target with URL validation
4. Update docs

## Files

- `Makefile` - Add ##@ CLI section with transcribe target
- `examples/transcribe/main.go` - Relocated CLI tool
- `CLAUDE.md` - Document new command
- `docs/development.md` - Usage examples
