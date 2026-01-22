# Implementation TODO

## Phase 1: API Changes - COMPLETED

### 1.1 Add MAX_UPLOAD_SIZE to config
- [x] Add `MaxUploadSize int64` field to Config struct
- [x] Add env var parsing with default 500MB (524288000)
- [x] Update .env.example with new variable

### 1.2 Create file validation helper
- [x] Add `ValidateFileExtension(filename string) bool` to models/transcript.go
- [x] Supported extensions: mp3, wav, m4a, flac, ogg, aac, mp4, mkv, webm, avi, mov
- [x] Add tests for file validation

### 1.3 Update handlers/transcribe.go
- [x] Detect content type (application/json vs multipart/form-data)
- [x] For multipart: extract file from form
- [x] Validate file extension
- [x] Check file size against MAX_UPLOAD_SIZE
- [x] Save file to WorkDir with job ID prefix
- [x] Pass local file path to processing pipeline
- [x] Add defer cleanup for uploaded file

### 1.4 Add tests for file upload
- [x] Test multipart file upload with valid audio file
- [x] Test file size limit enforcement
- [x] Test invalid file extension rejection
- [x] Test that JSON URL flow still works

## Phase 2: Documentation Updates - COMPLETED

### 2.1 README.md changes
- [x] Update tagline: "from any URL" → "from URLs or local files"
- [x] Add "Transcribing Local Files" section after Quick Start
- [x] Update Features list to mention local files first
- [x] Add file upload cURL example in Usage Examples
- [x] Add Go library local file example

### 2.2 docs/api.md updates
- [x] Document multipart/form-data content type
- [x] Add file upload request/response examples
- [x] Document supported file types
- [x] Document MAX_UPLOAD_SIZE limit

### 2.3 docs/troubleshooting.md
- [x] Add FAQ: "Does OmniTranscripts only work with YouTube?"

## Phase 3: Verification - COMPLETED

- [x] Run `go fmt ./... && go vet ./... && go test ./...`
- [x] Build compiles successfully
- [x] All tests pass
