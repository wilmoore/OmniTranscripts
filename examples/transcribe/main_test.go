package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"omnitranscripts/engine"
)

func TestRunWritesOneProgressMessageAndCleanTranscript(t *testing.T) {
	var stdout, stderr bytes.Buffer
	transcribe := func(context.Context, string, string, engine.Options) (*engine.Result, error) {
		return &engine.Result{Transcript: "hello world", Segments: []engine.Segment{{}}}, nil
	}

	code := run([]string{"https://youtube.com/watch?v=abcdefghijk"}, &stdout, &stderr, transcribe)
	if code != 0 {
		t.Fatalf("run() code = %d", code)
	}
	if stdout.String() != "hello world" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if count := strings.Count(stderr.String(), "Transcribing:"); count != 1 {
		t.Fatalf("progress count = %d; stderr = %q", count, stderr.String())
	}
}

func TestRunExplainsYouTubeCompatibilityFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	downloadErr := &engine.DownloaderError{
		Version:  "2026.07.04",
		ExitCode: 1,
		Details:  "WARNING: YouTube is forcing SABR streaming\nERROR: HTTP Error 403: Forbidden",
		Kind:     engine.DownloadFailureCompatibility,
		Err:      errors.New("exit status 1"),
	}
	transcribe := func(context.Context, string, string, engine.Options) (*engine.Result, error) {
		return nil, engine.NewError(engine.StageDownload, "failed to download audio", downloadErr)
	}

	code := run([]string{"https://youtube.com/watch?v=abcdefghijk"}, &stdout, &stderr, transcribe)
	if code != 1 {
		t.Fatalf("run() code = %d", code)
	}
	for _, want := range []string{
		"YouTube download compatibility failure",
		"bundled yt-dlp 2026.07.04",
		"Update OmniTranscripts",
		"Details:",
		"HTTP Error 403: Forbidden",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr lacks %q: %q", want, stderr.String())
		}
	}
	if strings.Contains(stderr.String(), "yt-dlp -U") {
		t.Fatalf("stderr contains unsupported self-update advice: %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunPreservesGenericStageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	transcribe := func(context.Context, string, string, engine.Options) (*engine.Result, error) {
		return nil, engine.NewError(engine.StageNormalize, "failed to normalize audio", errors.New("ffmpeg failed"))
	}

	code := run([]string{"https://example.com/media"}, &stdout, &stderr, transcribe)
	if code != 1 {
		t.Fatalf("run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "Cause: ffmpeg failed") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
