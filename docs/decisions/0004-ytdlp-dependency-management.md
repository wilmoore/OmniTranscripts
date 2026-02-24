# 0004. yt-dlp Dependency Management Strategy

Date: 2026-02-10

## Status

Accepted

## Context

OmniTranscripts uses the `go-ytdlp` library to download media from 1000+ platforms. This library bundles a specific version of yt-dlp, which is critical for platform support. When the bundled version becomes outdated:

- Newer platforms (e.g., updated Loom, Instagram API changes) may fail to download
- Security fixes in yt-dlp are not applied
- Platform-specific extractors may break as sites evolve

We discovered this when Loom downloads failed with the bundled yt-dlp (2025.09.05) but worked with the system-installed version (2025.12.08).

## Decision

**Use the bundled yt-dlp from go-ytdlp and keep the library updated.**

1. **Do not fall back to system yt-dlp**: Server deployments should not require system-wide yt-dlp installation. The bundled version ensures consistent behavior across environments.

2. **Regular go-ytdlp updates**: Include go-ytdlp version bumps in regular dependency updates. The library's bundled yt-dlp tracks upstream releases.

3. **Version pinning**: Pin go-ytdlp to specific versions in go.mod (not `latest`) for reproducible builds.

4. **Monitoring**: When platform downloads fail, check if a go-ytdlp update is available before investigating other causes.

## Consequences

### Positive

- No system dependencies required for deployment
- Consistent yt-dlp version across all environments
- Reproducible builds with pinned versions
- Single source of truth for yt-dlp version

### Negative

- Requires proactive dependency updates
- May lag behind upstream yt-dlp releases by days/weeks
- Platform breakages may require waiting for library update

## Alternatives Considered

### 1. System yt-dlp Fallback

Use system yt-dlp if available, fall back to bundled. Rejected because:
- Inconsistent behavior between dev and production
- Deployment complexity (requires yt-dlp installation)
- Version mismatches cause subtle bugs
- Violates "no system dependencies" deployment goal

### 2. Dynamic yt-dlp Installation

Use go-ytdlp's `Install()` to download latest yt-dlp at startup. Rejected because:
- Network dependency at startup
- Unpredictable behavior (different versions over time)
- Security concerns with auto-updating binaries
- Breaks reproducible deployments

### 3. Vendor yt-dlp Binary

Include yt-dlp binary directly in the repository. Rejected because:
- Large binary in git history
- Manual update process
- go-ytdlp already handles this well

## Implementation

### Dependency Updates

```bash
# Check for go-ytdlp updates
go list -m -versions github.com/lrstanley/go-ytdlp

# Update to latest version
go get github.com/lrstanley/go-ytdlp@latest
go mod tidy
```

### Bundled Binary Installation (Required)

The engine must call `ytdlp.Install()` with `DisableSystem: true` to ensure only the
bundled binary is used. This is implemented in `engine/engine.go`:

```go
var (
    ytdlpInstallOnce sync.Once
    ytdlpInstallErr  error
)

func ensureYtdlp(ctx context.Context) error {
    ytdlpInstallOnce.Do(func() {
        _, ytdlpInstallErr = ytdlp.Install(ctx, &ytdlp.InstallOptions{
            DisableSystem: true, // Never use system yt-dlp
        })
    })
    return ytdlpInstallErr
}
```

Call `ensureYtdlp()` before any yt-dlp operation (`downloadAudio`, `GetMediaDuration`).
The bundled binary is cached at `~/Library/Caches/go-ytdlp/`.

## Related

- Issue: Loom downloads failing with bundled yt-dlp v2025.09.05
- Fix: Updated go-ytdlp v1.2.4 → v1.2.7
- Issue: YouTube Shorts 403 Forbidden (SABR streaming protocol)
- Fix: Updated go-ytdlp v1.2.7 → v1.3.1, added `ensureYtdlp()` with `DisableSystem: true`
