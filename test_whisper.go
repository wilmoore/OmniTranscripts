//go:build ignore

package main

import (
	"fmt"

	"omnitranscripts/lib"
)

func main() {
	// Test whisper availability
	fmt.Println("Testing Whisper.cpp integration...")

	if lib.IsWhisperAvailable() {
		fmt.Println("✓ Whisper.cpp is available")
	} else {
		fmt.Println("✗ Whisper.cpp is not available")
		return
	}

	// Try to initialize with a test model path
	modelPath := "models/ggml-base.en.bin"
	fmt.Printf("Attempting to initialize whisper with model: %s\n", modelPath)

	ctx, err := lib.InitWhisper(modelPath)
	if err != nil {
		fmt.Printf("✗ Failed to initialize whisper: %v\n", err)
		fmt.Println("Note: This is expected if the model file doesn't exist yet")
		return
	}
	defer ctx.Free()

	fmt.Println("✓ Whisper context initialized successfully!")
	fmt.Println("CGO integration is working correctly")
}