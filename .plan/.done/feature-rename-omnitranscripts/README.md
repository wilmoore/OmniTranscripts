# Feature: Rename VideoTranscript to OmniTranscripts

## Overview
Rename the project from VideoTranscript to OmniTranscripts and formally expand scope from library-only to combined library + HTTP API.

**Branch**: `feature/rename-omnitranscripts`
**ADR**: `docs/decisions/0001-rename-to-omnitranscripts.md`
**Status**: Planning

---

## Implementation Plan

### Phase 1: Create Engine Package (Core Library)

**Goal**: Establish the public `engine/` package that external consumers can import.

#### Step 1.1: Create engine package structure
```
engine/
├── engine.go      # Package doc, core types, TranscriptionError
├── transcribe.go  # ProcessTranscription function (migrated from lib/)
├── options.go     # TranscriptionOptions type
└── doc.go         # Package documentation
```

#### Step 1.2: Define TranscriptionError
```go
type Stage string

const (
    StageDownload   Stage = "download"
    StageNormalize  Stage = "normalize"
    StageTranscribe Stage = "transcribe"
)

type TranscriptionError struct {
    Stage   Stage
    Message string
    Err     error
}

func (e *TranscriptionError) Error() string
func (e *TranscriptionError) Unwrap() error
```

#### Step 1.3: Migrate ProcessTranscription
- Move core transcription logic from `lib/transcription.go` to `engine/transcribe.go`
- Update to return `*TranscriptionError` for stage-specific errors
- Keep internal helpers (downloadAudio, normalizeAudio, etc.) as unexported

#### Step 1.4: Create TranscriptionOptions
```go
type TranscriptionOptions struct {
    WorkDir          string
    WhisperModelPath string
    AssemblyAIKey    string
    WhisperServerURL string
}
```

### Phase 2: Update Go Module

#### Step 2.1: Update go.mod
- Change `module videotranscript-app` to `module omnitranscripts`

#### Step 2.2: Update all import paths
Files to update:
- `main.go`
- `handlers/transcribe.go`
- `handlers/transcribe_test.go`
- `jobs/job.go`
- `jobs/queue.go`
- `lib/transcription.go`
- `lib/auth.go`
- `lib/subtitles.go`
- `lib/webhooks.go`
- `lib/audio_loader.go`
- `lib/whisper_native.go`
- `models/transcript.go`
- `models/transcript_test.go`
- `config/config.go`
- `transcribe/service.go`
- `transcribe/db.go`
- `web-dashboard.go`

### Phase 3: Update HTTP Layer

#### Step 3.1: Update health endpoint message
- Change "VideoTranscript.app API is running" to "OmniTranscripts API is running"

#### Step 3.2: Update handlers to use engine package
- Import `omnitranscripts/engine` instead of `omnitranscripts/lib`
- Adapt to new TranscriptionError type

### Phase 4: Update Documentation

#### Step 4.1: Update README.md
- Title: `# OmniTranscripts`
- Description: Universal media transcription
- Update all references to VideoTranscript
- Update Docker image names
- Update GitHub URLs (document pending rename)

#### Step 4.2: Update CLAUDE.md
- Project name and description
- All VideoTranscript references

#### Step 4.3: Update docs/*.md
- api.md
- architecture.md
- deployment.md
- development.md
- troubleshooting.md
- contributing.md
- changelog.md

#### Step 4.4: Update swagger.yaml
- API title and description

### Phase 5: Update Docker & Deployment

#### Step 5.1: Dockerfile updates
- Update any VideoTranscript references
- Update binary/image names if applicable

#### Step 5.2: Document GitHub repo rename steps
Create `docs/repo-rename.md`:
```markdown
# GitHub Repository Rename

When ready to rename the repository:
1. Go to Settings > General > Repository name
2. Change from `VideoTranscript.app` to `omnitranscripts`
3. Update go.mod to canonical path (optional)
4. Update README badge URLs
5. Update redirect references
```

### Phase 6: Validation

#### Step 6.1: Build validation
```bash
go build ./...
```

#### Step 6.2: Test validation
```bash
go test -short ./...
```

#### Step 6.3: Import validation
- Verify engine package is importable
- Verify no circular dependencies

---

## Files Changed Summary

### New Files
- `engine/engine.go` - Core types and TranscriptionError
- `engine/transcribe.go` - ProcessTranscription function
- `engine/options.go` - TranscriptionOptions
- `engine/doc.go` - Package documentation
- `docs/decisions/0001-rename-to-omnitranscripts.md` - ADR
- `docs/repo-rename.md` - Repo rename instructions

### Modified Files
- `go.mod` - Module name
- `main.go` - Imports, health message
- `handlers/*.go` - Imports
- `lib/*.go` - Imports, delegation to engine
- `models/*.go` - Imports
- `jobs/*.go` - Imports
- `config/*.go` - Imports
- `transcribe/*.go` - Imports
- `README.md` - Full update
- `CLAUDE.md` - Full update
- `docs/*.md` - Name references
- `Dockerfile` - Name references

---

## Rollback Plan

If issues arise:
1. `git checkout main`
2. `git stash pop` to restore whisper.cpp work
3. Branch preserved for future reference

---

## Definition of Done

- [ ] Engine package created with public API
- [ ] TranscriptionError type implemented
- [ ] All imports updated to omnitranscripts
- [ ] Health endpoint returns new name
- [ ] All tests pass
- [ ] All documentation updated
- [ ] ADR documented
- [ ] Repo rename steps documented
- [ ] Build succeeds
- [ ] No linting errors
