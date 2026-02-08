# TODO: MCP Server + ChatGPT Integration

## Completed

- [x] Add `github.com/mark3labs/mcp-go` dependency
- [x] Create `mcp/` package structure
  - [x] `mcp/types.go` - Request/response types
  - [x] `mcp/tools.go` - Tool definitions and handlers
  - [x] `mcp/server.go` - MCP server setup
- [x] Implement `transcribe_url` tool handler
- [x] Implement `get_transcription` tool handler
- [x] Add authentication middleware (reuses API key)
- [x] Mount MCP server at `/mcp` endpoint in `main.go`
- [x] Share job queue between HTTP API and MCP server
- [x] Add MCP config options to `config/config.go`
- [x] Add unit tests for MCP tool handlers
- [x] Create `docs/chatgpt-integration.md` setup guide
- [x] Update `docs/api.md` with MCP server reference
- [x] Update `CLAUDE.md` with MCP package info

## Remaining / Future

- [ ] Add progress tracking to job processing (Phase 3)
  - Stage indicators: "downloading", "normalizing", "transcribing"
- [ ] Store media metadata (title, duration) in job record
- [ ] Add word count to completed transcripts in job record
- [ ] Test end-to-end with real ChatGPT integration (Phase 5)
- [ ] Create app manifest for ChatGPT developer portal
- [ ] Document app submission process

## Notes

- MCP server is enabled by default (`MCP_ENABLED=true`)
- Uses async-first approach to handle videos of any length
- Authentication reuses the existing `API_KEY` environment variable
- CORS headers configured for ChatGPT integration
