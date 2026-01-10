//go:build !cgo

package lib

import "fmt"

// WhisperContext is a stub for non-CGO builds
type WhisperContext struct{}

// InitWhisper returns an error on non-CGO builds
func InitWhisper(modelPath string) (*WhisperContext, error) {
	return nil, fmt.Errorf("whisper.cpp requires CGO; build with CGO_ENABLED=1")
}

// Free is a no-op for the stub
func (w *WhisperContext) Free() {}

// TranscribeAudio returns an error on non-CGO builds
func (w *WhisperContext) TranscribeAudio(samples []float32) ([]TranscriptSegment, error) {
	return nil, fmt.Errorf("whisper.cpp requires CGO; build with CGO_ENABLED=1")
}

// IsWhisperAvailable returns false for non-CGO builds
func IsWhisperAvailable() bool {
	return false
}
