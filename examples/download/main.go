// CLI tool for downloading media from any URL using yt-dlp
//
// Supports:
//   - YouTube, YouTube Music, Instagram, TikTok, Vimeo, and 1000+ platforms
//   - Auto-detects best format (audio for music URLs, video otherwise)
//
// Usage:
//
//	go run main.go <url> [output_path]
//	make download URL="https://music.youtube.com/watch?v=..."
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lrstanley/go-ytdlp"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	url := os.Args[1]
	outputPath := ""
	if len(os.Args) >= 3 && os.Args[2] != "" {
		outputPath = os.Args[2]
	}

	// Default to ~/Downloads if no output specified
	if outputPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error: Could not determine home directory: %v\n", err)
			os.Exit(1)
		}
		outputPath = filepath.Join(homeDir, "Downloads")
	}

	// Ensure output directory exists
	outputDir := outputPath
	if !isDirectory(outputPath) {
		outputDir = filepath.Dir(outputPath)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error: Could not create output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Output: %s\n", outputPath)

	// Detect if this is likely an audio-only source
	isAudioSource := isAudioURL(url)
	if isAudioSource {
		fmt.Println("Type: Audio (music source detected)")
	} else {
		fmt.Println("Type: Video")
	}
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Install bundled yt-dlp
	if _, err := ytdlp.Install(ctx, &ytdlp.InstallOptions{DisableSystem: true}); err != nil {
		fmt.Printf("Error: Failed to install yt-dlp: %v\n", err)
		os.Exit(1)
	}

	// Configure yt-dlp based on content type
	dl := ytdlp.New()

	if isAudioSource {
		// Audio: extract best audio format
		dl = dl.ExtractAudio().AudioQuality("0")
	}

	// Set output template
	if isDirectory(outputPath) {
		// Directory: use default filename
		dl = dl.Output(filepath.Join(outputPath, "%(title)s.%(ext)s"))
	} else {
		// Specific file path
		dl = dl.Output(outputPath)
	}

	// Run download
	fmt.Println("Downloading...")
	result, err := dl.Run(ctx, url)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		os.Exit(1)
	}

	if result.ExitCode != 0 {
		fmt.Printf("Download failed (exit code %d): %s\n", result.ExitCode, result.Stderr)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Download complete!")
}

// isAudioURL detects if the URL is likely an audio-only source
func isAudioURL(url string) bool {
	audioPatterns := []string{
		"music.youtube.com",
		"soundcloud.com",
		"spotify.com",
		"bandcamp.com",
		"audiomack.com",
		"mixcloud.com",
	}

	urlLower := strings.ToLower(url)
	for _, pattern := range audioPatterns {
		if strings.Contains(urlLower, pattern) {
			return true
		}
	}
	return false
}

// isDirectory checks if path is an existing directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func printUsage() {
	fmt.Println("Usage: go run main.go <url> [output_path]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Download to ~/Downloads/")
	fmt.Println("  go run main.go https://music.youtube.com/watch?v=...")
	fmt.Println()
	fmt.Println("  # Download to specific location")
	fmt.Println("  go run main.go https://www.youtube.com/watch?v=... ./video.mp4")
	fmt.Println()
	fmt.Println("Or use the Makefile:")
	fmt.Println("  make download URL=\"https://...\"")
	fmt.Println("  make download URL=\"https://...\" OUT=./my-file.mp4")
}
