# Feature Plan: README Local Files Positioning + File Upload API

## Branch
`feat/readme-local-files-positioning`

## Problem Statement

The README mentally anchors users to YouTube and URLs, causing them to believe OmniTranscripts only works with streaming platforms. This is a positioning and affordance problem:

1. **Tagline excludes local files**: "from any URL" doesn't suggest local `.mp4` files
2. **All examples use YouTube/HTTP**: Zero examples show local file paths
3. **HTTP API lacks file upload**: Users can't upload files via the API - only Go library supports local paths
4. **No explicit documentation**: Supported input types are never spelled out

## Solution

Two-part fix:

### Part 1: HTTP API File Upload Support
Add multipart/form-data support to `POST /transcribe` endpoint so users can upload local files directly.

### Part 2: README & Documentation Updates
- Update tagline to include "or local files"
- Add "Local Files" section with examples
- Add cURL file upload example
- Update Features list
- Add FAQ entry in Troubleshooting

## Technical Design

### File Upload Endpoint

**Endpoint:** `POST /transcribe` (extend existing)

**Content Types Supported:**
- `application/json` (existing URL-based flow)
- `multipart/form-data` (new file upload flow)

**Request (multipart):**
```
POST /transcribe
Content-Type: multipart/form-data
Authorization: Bearer YOUR_API_KEY

file: <binary file data>
```

**Supported File Types:**
- Audio: `.mp3`, `.wav`, `.m4a`, `.flac`, `.ogg`, `.aac`
- Video: `.mp4`, `.mkv`, `.webm`, `.avi`, `.mov`

**Processing Flow:**
1. Detect content type
2. If multipart: save uploaded file to `WORK_DIR` with job ID prefix
3. Pass local file path to `engine.Transcribe()` (yt-dlp handles local files)
4. Clean up uploaded file after processing

**File Size Limit:** Configurable via `MAX_UPLOAD_SIZE` env var (default: 500MB)

### Code Changes Required

1. **handlers/transcribe.go**
   - Add content-type detection
   - Add multipart file handling
   - Save uploaded file to work directory
   - Pass file path to processing pipeline

2. **models/transcript.go**
   - Update `ValidateURL()` to `ValidateSource()` that accepts URLs or file paths
   - Add file extension validation helper

3. **config/config.go**
   - Add `MAX_UPLOAD_SIZE` configuration

4. **README.md**
   - Update tagline
   - Add "Local Files" section
   - Update Features list
   - Add file upload cURL example

5. **docs/api.md**
   - Document multipart/form-data support
   - Add file upload examples

6. **docs/troubleshooting.md**
   - Add FAQ: "Does OmniTranscripts only work with YouTube?"

## Implementation Steps

### Phase 1: API Changes (Code)

- [ ] 1.1 Add `MAX_UPLOAD_SIZE` to config/config.go
- [ ] 1.2 Create file validation helper in models/transcript.go
- [ ] 1.3 Update handlers/transcribe.go to detect content type
- [ ] 1.4 Implement multipart file upload handling
- [ ] 1.5 Add file cleanup after processing
- [ ] 1.6 Write tests for file upload endpoint

### Phase 2: Documentation Updates

- [ ] 2.1 Update README.md tagline
- [ ] 2.2 Add "Local Files" section to README.md
- [ ] 2.3 Update Features list in README.md
- [ ] 2.4 Add file upload cURL example to README.md
- [ ] 2.5 Update docs/api.md with multipart documentation
- [ ] 2.6 Add FAQ to docs/troubleshooting.md

### Phase 3: Verification

- [ ] 3.1 Run `make check` (fmt + lint + vet + test)
- [ ] 3.2 Manual test: file upload via cURL
- [ ] 3.3 Manual test: URL-based transcription still works
- [ ] 3.4 Review all documentation for consistency

## Acceptance Criteria

1. `POST /transcribe` accepts multipart/form-data with file uploads
2. Uploaded files are processed through the same pipeline as URLs
3. File size limits are enforced
4. Uploaded files are cleaned up after processing
5. README clearly shows local file support as first-class
6. All existing URL-based functionality unchanged
7. Tests pass for both URL and file upload flows

## Out of Scope

- Drag-and-drop UI (no frontend in this project)
- stdin support (CLI tool, not API)
- Batch file uploads (single file per request)

## Related ADRs

- **ADR-0001**: Establishes dual consumption model (library + API) and universal media scope
- **ADR-0002**: Examples directory for usage patterns (`doc/decisions/0002-examples-directory-for-usage-patterns.md`)

## Notes

- yt-dlp natively supports local file paths, so no changes needed to engine/
- File uploads bypass yt-dlp download stage but still go through normalize + transcribe
- Consider adding `source_type` field to response to indicate if input was URL or file
