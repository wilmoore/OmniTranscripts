package engine

import "fmt"

// Stage represents a stage in the transcription pipeline.
type Stage string

const (
	// StageDownload is the audio/video download stage using yt-dlp.
	StageDownload Stage = "download"
	// StageNormalize is the audio normalization stage using ffmpeg.
	StageNormalize Stage = "normalize"
	// StageTranscribe is the speech-to-text transcription stage.
	StageTranscribe Stage = "transcribe"
)

// TranscriptionError represents a stage-specific error in the transcription pipeline.
// It wraps the underlying error and provides context about which stage failed.
type TranscriptionError struct {
	// Stage indicates which pipeline stage encountered the error.
	Stage Stage
	// Message provides a human-readable description of the error.
	Message string
	// Err is the underlying error, if any.
	Err error
}

// Error implements the error interface.
func (e *TranscriptionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Stage, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Stage, e.Message)
}

// Unwrap returns the underlying error for errors.Is and errors.As.
func (e *TranscriptionError) Unwrap() error {
	return e.Err
}

// NewError creates a new TranscriptionError for the given stage.
func NewError(stage Stage, message string, err error) *TranscriptionError {
	return &TranscriptionError{
		Stage:   stage,
		Message: message,
		Err:     err,
	}
}

// IsDownloadError returns true if the error occurred during the download stage.
func IsDownloadError(err error) bool {
	var tErr *TranscriptionError
	if ok := errorAs(err, &tErr); ok {
		return tErr.Stage == StageDownload
	}
	return false
}

// IsNormalizeError returns true if the error occurred during the normalize stage.
func IsNormalizeError(err error) bool {
	var tErr *TranscriptionError
	if ok := errorAs(err, &tErr); ok {
		return tErr.Stage == StageNormalize
	}
	return false
}

// IsTranscribeError returns true if the error occurred during the transcribe stage.
func IsTranscribeError(err error) bool {
	var tErr *TranscriptionError
	if ok := errorAs(err, &tErr); ok {
		return tErr.Stage == StageTranscribe
	}
	return false
}

// errorAs is a helper for errors.As to avoid import in this file.
// The actual errors.As is used elsewhere; this is just for the helper functions.
func errorAs(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	// Simple type assertion for our specific case
	if tErr, ok := err.(*TranscriptionError); ok {
		if t, ok := target.(**TranscriptionError); ok {
			*t = tErr
			return true
		}
	}
	return false
}
