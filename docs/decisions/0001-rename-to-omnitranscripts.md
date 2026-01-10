# ADR-0001: Rename VideoTranscript to OmniTranscripts

## Status
Accepted

## Date
2026-01-09

## Context

The project was originally named "VideoTranscript" based on its initial use case: transcribing YouTube videos. However, the system's actual capabilities are broader:

1. **Multi-format support**: The system uses yt-dlp which supports 1000+ websites and platforms, not just YouTube
2. **Audio-first capability**: Audio-only files (podcasts, voice memos, audio recordings) are first-class citizens
3. **URL agnosticism**: Any valid HTTP/HTTPS URL is accepted; failures occur during processing, not validation
4. **Platform independence**: The transcription engine is transport-agnostic and works with any audio source

The name "VideoTranscript" is inaccurate because:
- It suggests video-only when audio-only is fully supported
- It implies platform specificity when the system is platform-agnostic
- It undersells the actual capability (universal media transcription)

Additionally, the project needs to formally support two consumption modes:
1. **Go library**: Direct programmatic use via `import`
2. **HTTP API**: Network-accessible service for external clients

## Decision

1. **Rename the project** from `VideoTranscript` to `OmniTranscripts`
2. **Use simple module name**: `omnitranscripts` (not canonical path)
3. **Create `engine/` package**: Expose core transcription as public Go package
4. **Add `TranscriptionError` type**: Structured errors with stage identification
5. **HTTP API as adapter**: The API remains a thin layer over the engine, not a separate product

### Module Name Rationale

Chose `omnitranscripts` over `github.com/wilmoore/omnitranscripts` because:
- Works immediately for local development
- No dependency on GitHub URL being valid
- Shorter import paths
- Can migrate to canonical path when publishing publicly

### Package Structure After Change

```
omnitranscripts/
├── engine/           # Public library package
│   ├── engine.go     # Core types and TranscriptionError
│   ├── transcribe.go # ProcessTranscription function
│   └── options.go    # Configuration options
├── handlers/         # HTTP handlers (internal)
├── lib/              # Internal utilities (deprecated, migrate to engine/)
├── jobs/             # Job queue (internal)
├── config/           # Configuration loading
└── main.go           # HTTP server entry point
```

## Consequences

### Positive
- Name accurately reflects capabilities
- Clear separation between library and HTTP API concerns
- External Go consumers can import `omnitranscripts/engine`
- Structured errors improve debugging and error handling
- Positioned as infrastructure rather than one-off utility

### Negative
- Breaking change for any existing import paths
- Documentation updates required across all files
- GitHub repo rename recommended (but not required immediately)

### Neutral
- No behavior changes to transcription pipeline
- No new dependencies
- Stashed work on whisper.cpp integration preserved separately

## References
- Change request: Rename and Scope Update (2026-01-09)
- Related: whisper.cpp native integration (stashed, separate work)
