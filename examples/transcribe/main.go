// CLI tool for transcribing any URL or local file using OmniTranscripts
//
// Supports:
//   - YouTube, Loom, Instagram, TikTok, Vimeo, and 1000+ platforms via yt-dlp
//   - Local audio/video files (mp4, mp3, wav, etc.)
//
// Usage:
//
//	build/transcribe [--verbose] <url_or_file_path>
//	make transcribe URL="https://youtube.com/watch?v=..."
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"omnitranscripts/engine"
)

type transcribeFunc func(context.Context, string, string, engine.Options) (*engine.Result, error)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, engine.Transcribe))
}

func run(args []string, stdout, stderr io.Writer, transcribe transcribeFunc) int {
	flags := flag.NewFlagSet("transcribe", flag.ContinueOnError)
	flags.SetOutput(stderr)
	verbose := flags.Bool("verbose", false, "show progress and diagnostic output")
	flags.BoolVar(verbose, "v", false, "show progress and diagnostic output")
	flags.Usage = func() { printUsage(stderr) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if flags.NArg() != 1 {
		printUsage(stderr)
		return 1
	}

	input := flags.Arg(0)

	// Determine if input is a URL or local file
	isURL := strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")

	if !isURL {
		// Check if local file exists
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fmt.Fprintf(stderr, "Error: File not found: %s\n", input)
			return 1
		}
	}

	if *verbose {
		fmt.Fprintf(stderr, "Transcribing: %s\n", input)
		if isURL {
			fmt.Fprintln(stderr, "Type: URL (downloading via yt-dlp)")
		} else {
			fmt.Fprintln(stderr, "Type: Local file")
		}
		fmt.Fprintln(stderr)
	}

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
	opts.Verbose = *verbose
	if opts.WhisperModelPath == "" {
		opts.WhisperModelPath = defaultModelPath()
	}
	result, err := transcribe(ctx, input, "cli-transcribe", opts)
	if err != nil {
		writeTranscriptionError(stderr, err, *verbose)
		return 1
	}

	elapsed := time.Since(startTime)

	// Output ONLY transcript to stdout (clean for piping)
	fmt.Fprint(stdout, result.Transcript)

	if *verbose {
		fmt.Fprintln(stderr, "\n--- Summary ---")
		fmt.Fprintf(stderr, "Duration: %s\n", elapsed.Round(time.Second))
		fmt.Fprintf(stderr, "Segments: %d\n", len(result.Segments))
	}
	return 0
}

func defaultModelPath() string {
	candidates := []string{"models/ggml-base.en.bin"}
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), "..", "models", "ggml-base.en.bin"))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func writeTranscriptionError(stderr io.Writer, err error, verbose bool) {
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
		if verbose && downloaderErr.Details != "" {
			fmt.Fprintln(stderr, "Details:")
			for _, line := range strings.Split(downloaderErr.Details, "\n") {
				fmt.Fprintf(stderr, "  %s\n", line)
			}
		}
		return
	}

	if verbose && tErr.Err != nil {
		fmt.Fprintf(stderr, "  Cause: %v\n", tErr.Err)
	}
}

func printUsage(stderr io.Writer) {
	fmt.Fprintln(stderr, "Usage: build/transcribe [--verbose] <url_or_file_path>")
	fmt.Fprintln(stderr, "\nOptions:")
	fmt.Fprintln(stderr, "  -v, --verbose  Show progress and diagnostic output")
	fmt.Fprintln(stderr, "\nExamples:")
	fmt.Fprintln(stderr, "  # Transcribe a YouTube video")
	fmt.Fprintln(stderr, "  build/transcribe https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	fmt.Fprintln(stderr, "\n  # Transcribe an Instagram reel")
	fmt.Fprintln(stderr, "  build/transcribe https://www.instagram.com/reel/ABC123/")
	fmt.Fprintln(stderr, "\n  # Transcribe a public Loom share")
	fmt.Fprintln(stderr, "  build/transcribe https://www.loom.com/share/VIDEO_ID")
	fmt.Fprintln(stderr, "\n  # Transcribe a local file")
	fmt.Fprintln(stderr, "  build/transcribe /path/to/video.mp4")
	fmt.Fprintln(stderr, "\nBuild with: make build-transcribe")
	fmt.Fprintln(stderr, "\nMake convenience wrapper:")
	fmt.Fprintln(stderr, "  make transcribe URL=\"https://youtube.com/watch?v=...\"")
	fmt.Fprintln(stderr, "  make transcribe URL=\"https://youtube.com/watch?v=...\" VERBOSE=1")
	fmt.Fprintln(stderr, "  make transcribe URL=\"https://youtube.com/watch?v=...\" | pbcopy")
}
