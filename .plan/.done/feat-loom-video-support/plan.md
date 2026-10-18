# Loom video support

## Requirements

- Explicitly support public `https://www.loom.com/share/<id>` URLs in the standalone CLI and Make convenience workflow.
- Keep transcript-only default output and opt-in verbose diagnostics unchanged.
- Document Loom alongside the other named supported platforms.
- Add deterministic regression coverage for Loom URL routing.
- Verify the supplied public Loom URL end to end with the pinned downloader and native Whisper.

## Decisions

- Continue using the bundled, pinned yt-dlp extractor under ADR 0004; do not add a Loom-specific scraper or system downloader fallback.
- Pass Loom share URLs unchanged to the engine because yt-dlp already resolves the underlying media.
- Treat support as public-share support. Authenticated, workspace-restricted, and password-protected Loom videos are outside this scope.
- Do not add new output or platform-specific flags.

## Non-goals

- Loom authentication, cookies, passwords, or private workspace access.
- Importing Loom captions, comments, chapters, or metadata.
- Bypassing Loom access controls.
- Changing downloader pinning or cache-key behavior.

## Implementation

1. Add a CLI regression test proving a Loom share URL reaches the engine unchanged as a URL.
2. Add Loom examples to CLI help, README, API, development, and troubleshooting documentation where platform support is described.
3. Verify quiet output and direct CLI/Make examples remain accurate.

## Acceptance criteria

- The supplied public Loom share URL downloads, normalizes, and transcribes successfully.
- Loom appears in CLI help and primary usage documentation.
- Tests prevent Loom share URLs from being mistaken for local paths or rewritten incorrectly.
- Existing output-stream and downloader-management ADRs remain satisfied.

## Verification

- Run targeted CLI and engine tests.
- Build the CGO standalone CLI.
- Run all non-handler repository tests and full vet, accounting for the existing handler startup-hang backlog item.
- Run the supplied Loom URL through native transcription in verbose and quiet modes.

## Risks

- Loom may change its public share page or media APIs; ongoing compatibility follows the pinned yt-dlp update policy in ADR 0004.
- Public links can later become private or be deleted, so the live fixture is smoke evidence rather than a deterministic unit-test dependency.

## Deferred issues

- Private or authenticated Loom access remains deferred until a concrete authorized use case is provided.

## Delivered result

- Public Loom share URLs are explicitly supported and documented across the standalone CLI, Make workflow, API guide, OpenAPI examples, development guide, troubleshooting guide, ChatGPT integration, and marketplace copy.
- Loom URLs pass unchanged through the existing URL pipeline backed by the pinned bundled yt-dlp extractor.
- CLI help now includes a public Loom example, with deterministic regression coverage for routing and help discovery.
- No new scraper, authentication mechanism, output mode, or architectural decision was introduced.

## Verification evidence

- Supplied Loom fixture completed end to end in verbose mode: exit 0, 4,135 transcript bytes, 63 segments, approximately 52 seconds.
- The cached quiet-mode run completed with the same 4,135 transcript bytes and zero stderr bytes.
- `go test ./examples/transcribe ./engine` passes.
- Every repository package except the already-backlogged hanging `handlers` package passes.
- `go vet ./...` passes.
- `make build-transcribe` succeeds with CGO/native Whisper.
- `docs/swagger.yaml` parses successfully as YAML.
- `git diff --check` passes.

## Related ADRs

- [ADR 0004: yt-dlp Dependency Management Strategy](../../../docs/decisions/0004-ytdlp-dependency-management.md)
