# Short-Form Video Transcription

Examples for transcribing short-form video content from YouTube Shorts, TikTok, and Instagram Reels.

## Supported Platforms

| Platform | URL Format | Notes |
|----------|------------|-------|
| YouTube Shorts | `youtube.com/shorts/VIDEO_ID` | Full support |
| TikTok | `tiktok.com/@user/video/ID` | Public videos only |
| Instagram Reels | `instagram.com/reel/ID/` | Public reels only |

## Limitations

- **Private content**: Authentication-gated or friends-only content is not supported
- **Region locks**: Some content may be unavailable in certain regions
- **Rate limits**: Platforms may rate-limit requests; consider adding delays for batch processing

## Examples

Run these scripts after setting your API key:
```bash
export API_KEY="your-api-key-here"
```

- `youtube-shorts.sh` - Transcribe a YouTube Short
- `tiktok.sh` - Transcribe a TikTok video
- `instagram-reels.sh` - Transcribe an Instagram Reel
