package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"omnitranscripts/engine"
)

func TestRunIsQuietByDefault(t *testing.T) {
	var stdout, stderr bytes.Buffer
	transcribe := func(_ context.Context, _ string, _ string, opts engine.Options) (*engine.Result, error) {
		if opts.Verbose {
			t.Fatal("default run enabled verbose engine diagnostics")
		}
		return &engine.Result{Transcript: "hello world", Segments: []engine.Segment{{}}}, nil
	}

	code := run([]string{"https://youtube.com/watch?v=abcdefghijk"}, &stdout, &stderr, transcribe)
	if code != 0 {
		t.Fatalf("run() code = %d", code)
	}
	if stdout.String() != "hello world" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunVerboseWritesDiagnosticsAndKeepsStdoutClean(t *testing.T) {
	var stdout, stderr bytes.Buffer
	transcribe := func(_ context.Context, _ string, _ string, opts engine.Options) (*engine.Result, error) {
		if !opts.Verbose {
			t.Fatal("verbose run did not enable engine diagnostics")
		}
		return &engine.Result{Transcript: "hello world", Segments: []engine.Segment{{}}}, nil
	}

	code := run([]string{"--verbose", "https://youtube.com/watch?v=abcdefghijk"}, &stdout, &stderr, transcribe)
	if code != 0 {
		t.Fatalf("run() code = %d", code)
	}
	if stdout.String() != "hello world" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	for _, want := range []string{"Transcribing:", "Type: URL", "--- Summary ---", "Segments: 1"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr lacks %q: %q", want, stderr.String())
		}
	}
}

func TestRunPassesLoomShareURLToEngineUnchanged(t *testing.T) {
	const loomURL = "https://www.loom.com/share/4347ba627d6a483db2e55927d1ecb2bc"
	var stdout, stderr bytes.Buffer
	transcribe := func(_ context.Context, input, _ string, _ engine.Options) (*engine.Result, error) {
		if input != loomURL {
			t.Fatalf("input = %q, want %q", input, loomURL)
		}
		return &engine.Result{Transcript: "loom transcript"}, nil
	}

	code := run([]string{loomURL}, &stdout, &stderr, transcribe)
	if code != 0 {
		t.Fatalf("run() code = %d", code)
	}
	if stdout.String() != "loom transcript" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunHelpDocumentsStandaloneCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	transcribe := func(context.Context, string, string, engine.Options) (*engine.Result, error) {
		t.Fatal("help should not start transcription")
		return nil, nil
	}

	code := run([]string{"--help"}, &stdout, &stderr, transcribe)
	if code != 0 {
		t.Fatalf("run() code = %d", code)
	}
	for _, want := range []string{"Usage: build/transcribe", "--verbose", "loom.com/share/VIDEO_ID", "make build-transcribe"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("help lacks %q: %q", want, stderr.String())
		}
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
	if strings.Contains(stderr.String(), "HTTP Error 403") {
		t.Fatalf("default stderr contains raw diagnostics: %q", stderr.String())
	}
}

func TestRunVerboseIncludesCompatibilityFailureDetails(t *testing.T) {
	var stdout, stderr bytes.Buffer
	downloadErr := &engine.DownloaderError{
		Version: "2026.07.04",
		Details: "ERROR: HTTP Error 403: Forbidden",
		Kind:    engine.DownloadFailureCompatibility,
	}
	transcribe := func(context.Context, string, string, engine.Options) (*engine.Result, error) {
		return nil, engine.NewError(engine.StageDownload, "failed to download audio", downloadErr)
	}

	code := run([]string{"-v", "https://youtube.com/watch?v=abcdefghijk"}, &stdout, &stderr, transcribe)
	if code != 1 {
		t.Fatalf("run() code = %d", code)
	}
	for _, want := range []string{"Details:", "HTTP Error 403: Forbidden"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr lacks %q: %q", want, stderr.String())
		}
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
	if strings.Contains(stderr.String(), "Cause: ffmpeg failed") {
		t.Fatalf("default stderr contains raw cause: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "failed to normalize audio") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunVerbosePreservesGenericStageErrorCause(t *testing.T) {
	var stdout, stderr bytes.Buffer
	transcribe := func(context.Context, string, string, engine.Options) (*engine.Result, error) {
		return nil, engine.NewError(engine.StageNormalize, "failed to normalize audio", errors.New("ffmpeg failed"))
	}

	code := run([]string{"--verbose", "https://example.com/media"}, &stdout, &stderr, transcribe)
	if code != 1 {
		t.Fatalf("run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "Cause: ffmpeg failed") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
