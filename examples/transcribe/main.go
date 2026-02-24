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
	"fmt"
	"os"
	"strings"
	"time"

	"omnitranscripts/engine"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	input := os.Args[1]

	// Determine if input is a URL or local file
	isURL := strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")

	if !isURL {
		// Check if local file exists
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fmt.Printf("Error: File not found: %s\n", input)
			os.Exit(1)
		}
	}

	fmt.Printf("Transcribing: %s\n", input)
	if isURL {
		fmt.Println("Type: URL (downloading via yt-dlp)")
	} else {
		fmt.Println("Type: Local file")
	}
	fmt.Println()

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
	result, err := engine.Transcribe(ctx, input, "cli-transcribe", opts)
	if err != nil {
		// Provide stage-specific error context
		if tErr, ok := err.(*engine.TranscriptionError); ok {
			fmt.Printf("Transcription failed at stage '%s': %s\n", tErr.Stage, tErr.Message)
			if tErr.Err != nil {
				fmt.Printf("  Cause: %v\n", tErr.Err)
			}
		} else {
			fmt.Printf("Transcription failed: %v\n", err)
		}
		os.Exit(1)
	}

	elapsed := time.Since(startTime)

	// Print the transcript
	fmt.Println("--- Transcript ---")
	fmt.Println(result.Transcript)
	fmt.Println()

	// Print segments with timestamps
	if len(result.Segments) > 0 {
		fmt.Println("--- Segments ---")
		for _, seg := range result.Segments {
			fmt.Printf("[%0.1fs - %0.1fs] %s\n", seg.Start, seg.End, seg.Text)
		}
		fmt.Println()
	}

	// Print summary
	fmt.Println("--- Summary ---")
	fmt.Printf("Duration: %s\n", elapsed.Round(time.Second))
	fmt.Printf("Segments: %d\n", len(result.Segments))
}

func printUsage() {
	fmt.Println("Usage: go run main.go <url_or_file_path>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Transcribe a YouTube video")
	fmt.Println("  go run main.go https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	fmt.Println()
	fmt.Println("  # Transcribe an Instagram reel")
	fmt.Println("  go run main.go https://www.instagram.com/reel/ABC123/")
	fmt.Println()
	fmt.Println("  # Transcribe a local file")
	fmt.Println("  go run main.go /path/to/video.mp4")
	fmt.Println()
	fmt.Println("Or use the Makefile:")
	fmt.Println("  make transcribe URL=\"https://youtube.com/watch?v=...\"")
}
