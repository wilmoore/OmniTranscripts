// CLI tool for transcribing any URL or local file using OmniTranscripts
//
// Supports:
//   - YouTube, Instagram, TikTok, Vimeo, and 1000+ platforms via yt-dlp
//   - Local audio/video files (mp4, mp3, wav, etc.)
//
// Usage:
//
//	go run main.go <url_or_file_path>
//	make transcribe URL="https://youtube.com/watch?v=..."
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"omnitranscripts/engine"
)

type transcribeFunc func(context.Context, string, string, engine.Options) (*engine.Result, error)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, engine.Transcribe))
}

func run(args []string, stdout, stderr io.Writer, transcribe transcribeFunc) int {
	if len(args) < 1 {
		printUsage(stderr)
		return 1
	}

	input := args[0]

	// Determine if input is a URL or local file
	isURL := strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")

	if !isURL {
		// Check if local file exists
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fmt.Fprintf(stderr, "Error: File not found: %s\n", input)
			return 1
		}
	}

	// Print progress to stderr (not part of transcript output)
	fmt.Fprintf(stderr, "Transcribing: %s\n", input)
	if isURL {
		fmt.Fprintln(stderr, "Type: URL (downloading via yt-dlp)")
	} else {
		fmt.Fprintln(stderr, "Type: Local file")
	}
	fmt.Fprintln(stderr)

	// Create a context with timeout for the transcription
	// ADR-0003: Context propagation with appropriate timeouts
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	startTime := time.Now()

	// Use engine.Transcribe - the public library interface (ADR-0001)
	// URLs go through: yt-dlp download -> FFmpeg normalize -> Whisper transcribe
	// Local files go through: FFmpeg normalize -> Whisper transcribe
	opts := engine.DefaultOptions()
	opts.CacheDownloads = true // Cache downloads for CLI usage
	result, err := transcribe(ctx, input, "cli-transcribe", opts)
	if err != nil {
		writeTranscriptionError(stderr, err)
		return 1
	}

	elapsed := time.Since(startTime)

	// Output ONLY transcript to stdout (clean for piping)
	fmt.Fprint(stdout, result.Transcript)

	// Print diagnostic information to stderr
	fmt.Fprintln(stderr, "\n--- Summary ---")
	fmt.Fprintf(stderr, "Duration: %s\n", elapsed.Round(time.Second))
	fmt.Fprintf(stderr, "Segments: %d\n", len(result.Segments))
	return 0
}

func writeTranscriptionError(stderr io.Writer, err error) {
	var tErr *engine.TranscriptionError
	if !errors.As(err, &tErr) {
		fmt.Fprintf(stderr, "Transcription failed: %v\n", err)
		return
	}

	fmt.Fprintf(stderr, "Transcription failed at stage '%s': %s\n", tErr.Stage, tErr.Message)

	var downloaderErr *engine.DownloaderError
	if errors.As(err, &downloaderErr) && downloaderErr.IsCompatibilityFailure() {
		fmt.Fprintf(stderr, "YouTube download compatibility failure (bundled yt-dlp %s).\n", downloaderErr.Version)
		fmt.Fprintln(stderr, "Update OmniTranscripts to refresh its pinned downloader, or retry later.")
		if downloaderErr.Details != "" {
			fmt.Fprintln(stderr, "Details:")
			for _, line := range strings.Split(downloaderErr.Details, "\n") {
				fmt.Fprintf(stderr, "  %s\n", line)
			}
		}
		return
	}

	if tErr.Err != nil {
		fmt.Fprintf(stderr, "  Cause: %v\n", tErr.Err)
	}
}

func printUsage(stderr io.Writer) {
	fmt.Fprintln(stderr, "Usage: go run main.go <url_or_file_path>")
	fmt.Fprintln(stderr, "\nExamples:")
	fmt.Fprintln(stderr, "  # Transcribe a YouTube video")
	fmt.Fprintln(stderr, "  go run main.go https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	fmt.Fprintln(stderr, "\n  # Transcribe an Instagram reel")
	fmt.Fprintln(stderr, "  go run main.go https://www.instagram.com/reel/ABC123/")
	fmt.Fprintln(stderr, "\n  # Transcribe a local file")
	fmt.Fprintln(stderr, "  go run main.go /path/to/video.mp4")
	fmt.Fprintln(stderr, "\nOr use the Makefile:")
	fmt.Fprintln(stderr, "  make transcribe URL=\"https://youtube.com/watch?v=...\"")
}
