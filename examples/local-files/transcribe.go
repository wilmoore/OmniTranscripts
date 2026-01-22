// Example: Local file transcription using the Go library
//
// Run with: go run transcribe.go /path/to/file.mp4
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"omnitranscripts/engine"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run transcribe.go <file_path>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Error: File not found: %s\n", filePath)
		os.Exit(1)
	}

	fmt.Printf("Transcribing: %s\n", filePath)

	// Create a context with timeout for the transcription
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Transcribe the local file
	// Local files bypass the download stage and go directly through FFmpeg → Whisper
	result, err := engine.Transcribe(
		ctx,
		filePath,
		"local-file-example",
		engine.DefaultOptions(),
	)
	if err != nil {
		var tErr *engine.TranscriptionError
		if errors.As(err, &tErr) {
			fmt.Printf("Transcription failed at stage '%s': %s\n", tErr.Stage, tErr.Message)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	// Print the transcript
	fmt.Printf("\n--- Transcript ---\n%s\n", result.Transcript)

	// Print segments with timestamps
	fmt.Printf("\n--- Segments ---\n")
	for _, seg := range result.Segments {
		fmt.Printf("[%0.1fs - %0.1fs] %s\n", seg.Start, seg.End, seg.Text)
	}
}
