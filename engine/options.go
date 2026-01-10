package engine

import "os"

// Options configures the transcription engine behavior.
type Options struct {
	// WorkDir is the directory for temporary files during processing.
	// Defaults to /tmp/omnitranscripts if empty.
	WorkDir string

	// WhisperModelPath is the path to the whisper.cpp model file.
	// If set, native whisper.cpp transcription is attempted first.
	WhisperModelPath string

	// AssemblyAIKey is the API key for AssemblyAI transcription service.
	// Used as fallback if native whisper is unavailable.
	AssemblyAIKey string

	// WhisperServerURL is the URL of a whisper.cpp HTTP server.
	// Used as fallback if AssemblyAI is unavailable.
	WhisperServerURL string
}

// DefaultOptions returns Options populated from environment variables.
// This is the recommended way to configure the engine in most cases.
func DefaultOptions() Options {
	workDir := os.Getenv("WORK_DIR")
	if workDir == "" {
		workDir = "/tmp/omnitranscripts"
	}

	return Options{
		WorkDir:          workDir,
		WhisperModelPath: os.Getenv("WHISPER_MODEL_PATH"),
		AssemblyAIKey:    os.Getenv("ASSEMBLYAI_API_KEY"),
		WhisperServerURL: os.Getenv("WHISPER_SERVER_URL"),
	}
}

// Validate checks that the options are valid.
// Returns an error if WorkDir is empty.
func (o Options) Validate() error {
	if o.WorkDir == "" {
		return NewError(StageDownload, "work directory is required", nil)
	}
	return nil
}

// HasTranscriptionBackend returns true if at least one transcription
// backend is configured (native whisper, AssemblyAI, or whisper server).
func (o Options) HasTranscriptionBackend() bool {
	return o.WhisperModelPath != "" || o.AssemblyAIKey != "" || o.WhisperServerURL != ""
}
