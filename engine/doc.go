// Package engine provides universal media transcription capabilities.
//
// OmniTranscripts engine supports transcribing audio and video from any
// URL supported by yt-dlp (1000+ platforms), as well as local audio files.
// The transcription pipeline uses a hybrid approach with multiple backends:
//
//   - Native whisper.cpp (fastest, requires local model)
//   - AssemblyAI (cloud-based, requires API key)
//   - Whisper server (self-hosted whisper.cpp HTTP server)
//   - Demo mode (fallback for development)
//
// Basic usage:
//
//	result, err := engine.Transcribe("https://example.com/audio.mp3", engine.DefaultOptions())
//	if err != nil {
//	    var tErr *engine.TranscriptionError
//	    if errors.As(err, &tErr) {
//	        fmt.Printf("Failed at stage %s: %s\n", tErr.Stage, tErr.Message)
//	    }
//	    return err
//	}
//	fmt.Println(result.Transcript)
//
// The engine is transport-agnostic. For HTTP API access, see the main
// omnitranscripts package which provides a Fiber-based HTTP server.
package engine
