package engine

import (
	"errors"
	"strings"
	"testing"

	"github.com/lrstanley/go-ytdlp"
)

func TestPinnedYTDLPVersion(t *testing.T) {
	if ytdlp.Version != "2026.08.19" {
		t.Fatalf("go-ytdlp pins yt-dlp %s, want 2026.08.19", ytdlp.Version)
	}
}

func TestClassifyDownloadFailure(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		details string
		want    DownloadFailureKind
	}{
		{name: "SABR", url: "https://youtube.com/watch?v=abcdefghijk", details: "YouTube is forcing SABR streaming", want: DownloadFailureCompatibility},
		{name: "HTTP 403", url: "https://www.youtube.com/watch?v=abcdefghijk", details: "HTTP Error 403: Forbidden", want: DownloadFailureCompatibility},
		{name: "unavailable format", url: "https://youtu.be/abcdefghijk", details: "Requested format is not available", want: DownloadFailureCompatibility},
		{name: "stale downloader", url: "https://music.youtube.com/watch?v=abcdefghijk", details: "older than 90 days", want: DownloadFailureCompatibility},
		{name: "generic YouTube error", url: "https://youtube.com/watch?v=abcdefghijk", details: "network timeout", want: DownloadFailureGeneric},
		{name: "non YouTube 403", url: "https://example.com/media", details: "HTTP Error 403: Forbidden", want: DownloadFailureGeneric},
		{name: "lookalike host", url: "https://youtube.com.example.test/media", details: "SABR", want: DownloadFailureGeneric},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := classifyDownloadFailure(test.url, test.details); got != test.want {
				t.Fatalf("classifyDownloadFailure() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDownloaderError(t *testing.T) {
	cause := errors.New("exit status 1")
	raw := "\x1b[31mWARNING: Your yt-dlp version (2026.02.21) is older than 90 days!\x1b[0m\n" +
		"         Run yt-dlp -U to update.\n\n" +
		"WARNING: YouTube is forcing SABR streaming\n" +
		"WARNING: YouTube is forcing SABR streaming\n" +
		"ERROR: HTTP Error 403: Forbidden  \n"

	err := newDownloaderError("https://youtube.com/watch?v=abcdefghijk", "2026.02.21", 1, raw, cause)
	if err.Error() != "yt-dlp 2026.02.21 failed with exit code 1" {
		t.Fatalf("Error() = %q", err.Error())
	}
	if !errors.Is(err, cause) {
		t.Fatal("DownloaderError does not unwrap its cause")
	}
	if !err.IsCompatibilityFailure() {
		t.Fatal("expected compatibility failure")
	}
	if strings.Contains(err.Details, "older than 90 days") || strings.Contains(err.Details, "yt-dlp -U") {
		t.Fatalf("stale self-update warning was not removed: %q", err.Details)
	}
	if strings.Contains(err.Details, "\x1b") {
		t.Fatalf("ANSI escape was not removed: %q", err.Details)
	}
	if strings.Count(err.Details, "forcing SABR") != 1 {
		t.Fatalf("duplicate detail was not removed: %q", err.Details)
	}
	if !strings.Contains(err.Details, "HTTP Error 403: Forbidden") {
		t.Fatalf("useful diagnostic detail was lost: %q", err.Details)
	}
}
