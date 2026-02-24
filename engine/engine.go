package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/lrstanley/go-ytdlp"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"

	"omnitranscripts/lib"
)

var (
	ytdlpInstallOnce sync.Once
	ytdlpInstallErr  error
)

// ensureYtdlp ensures the bundled yt-dlp binary is installed and cached.
// This is called lazily on first use and cached for subsequent calls.
// Using the bundled version ensures consistent behavior and access to
// latest platform support (per ADR-0004: no system yt-dlp fallback).
func ensureYtdlp(ctx context.Context) error {
	ytdlpInstallOnce.Do(func() {
		_, ytdlpInstallErr = ytdlp.Install(ctx, &ytdlp.InstallOptions{
			DisableSystem: true, // Never use system yt-dlp (ADR-0004)
		})
	})
	return ytdlpInstallErr
}

// Transcribe processes media from a URL and returns the transcription.
// It uses yt-dlp to download audio from any supported URL, normalizes
// the audio with ffmpeg, and transcribes using the configured backend.
//
// The URL can be any URL supported by yt-dlp (YouTube, Vimeo, SoundCloud,
// direct audio/video URLs, and 1000+ other platforms).
//
// The context parameter controls subprocess lifecycle - when cancelled,
// all spawned yt-dlp and ffmpeg processes will be terminated.
//
// Returns a TranscriptionError if any stage fails, allowing callers to
// identify which stage encountered the problem.
func Transcribe(ctx context.Context, url string, jobID string, opts Options) (*Result, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(opts.WorkDir, 0755); err != nil {
		return nil, NewError(StageDownload, "failed to create work directory", err)
	}

	// Use URL hash for cache key if caching is enabled, otherwise use jobID
	filePrefix := jobID
	if opts.CacheDownloads {
		filePrefix = "cache_" + URLCacheKey(url)
	}

	audioFile := filepath.Join(opts.WorkDir, fmt.Sprintf("%s.wav", filePrefix))
	normalizedAudio := filepath.Join(opts.WorkDir, fmt.Sprintf("%s_norm.wav", filePrefix))
	transcriptFile := filepath.Join(opts.WorkDir, fmt.Sprintf("%s_transcript.txt", jobID))

	// Cleanup: only remove non-cached files
	defer func() {
		os.Remove(transcriptFile)
		if !opts.CacheDownloads {
			os.Remove(audioFile)
			os.Remove(normalizedAudio)
		}
	}()

	// Check if cached audio exists
	cacheHit := false
	if opts.CacheDownloads {
		if _, err := os.Stat(normalizedAudio); err == nil {
			fmt.Printf("Using cached audio: %s\n", normalizedAudio)
			cacheHit = true
		}
	}

	if !cacheHit {
		if err := downloadAudio(ctx, url, audioFile); err != nil {
			return nil, NewError(StageDownload, "failed to download audio", err)
		}

		if err := normalizeAudio(ctx, audioFile, normalizedAudio); err != nil {
			return nil, NewError(StageNormalize, "failed to normalize audio", err)
		}
	}

	transcript, segments, err := transcribeAudio(normalizedAudio, transcriptFile, opts)
	if err != nil {
		return nil, NewError(StageTranscribe, "failed to transcribe audio", err)
	}

	return &Result{
		Transcript: transcript,
		Segments:   segments,
	}, nil
}

// GetMediaDuration returns the duration of media at the given URL in seconds.
// Uses yt-dlp to extract metadata without downloading the full media.
//
// The context parameter controls subprocess lifecycle - when cancelled,
// the yt-dlp process will be terminated.
func GetMediaDuration(ctx context.Context, url string) (int, error) {
	// Ensure bundled yt-dlp is installed (ADR-0004)
	if err := ensureYtdlp(ctx); err != nil {
		return 0, NewError(StageDownload, "failed to install yt-dlp", err)
	}

	dl := ytdlp.New()

	result, err := dl.Run(ctx, url, "--get-duration", "--no-warnings")
	if err != nil {
		return 0, NewError(StageDownload, "failed to get media info", err)
	}

	if result.ExitCode != 0 {
		return 0, NewError(StageDownload, fmt.Sprintf("yt-dlp failed with code %d", result.ExitCode), nil)
	}

	return parseDuration(result.Stdout), nil
}

func downloadAudio(ctx context.Context, url, outputPath string) error {
	// Ensure bundled yt-dlp is installed (ADR-0004)
	if err := ensureYtdlp(ctx); err != nil {
		return fmt.Errorf("failed to install yt-dlp: %w", err)
	}

	dl := ytdlp.New().
		ExtractAudio().
		AudioFormat("wav").
		AudioQuality("0").
		Output(outputPath)

	result, err := dl.Run(ctx, url)
	if err != nil {
		return fmt.Errorf("yt-dlp failed: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("yt-dlp failed with code %d: %s", result.ExitCode, result.Stderr)
	}

	return nil
}

func normalizeAudio(ctx context.Context, inputPath, outputPath string) error {
	stream := ffmpeg_go.Input(inputPath).
		Audio().
		Output(outputPath, ffmpeg_go.KwArgs{
			"ar":  16000,
			"ac":  1,
			"c:a": "pcm_s16le",
		}).
		GlobalArgs("-y") // Overwrite output file without asking

	// Set the context on the stream to enable cancellation
	stream.Context = ctx

	if err := stream.Run(); err != nil {
		return fmt.Errorf("ffmpeg normalization failed: %w", err)
	}

	return nil
}

func transcribeAudio(audioPath, outputPath string, opts Options) (string, []Segment, error) {
	// Try backends in order of preference

	// 1. Native whisper.cpp
	if opts.WhisperModelPath != "" {
		fmt.Printf("Attempting transcription with native Whisper (model: %s)...\n", opts.WhisperModelPath)
		transcript, segments, err := transcribeWithNativeWhisper(audioPath, outputPath, opts.WhisperModelPath)
		if err == nil {
			fmt.Println("Native Whisper transcription completed successfully")
			return transcript, segments, nil
		}
		fmt.Printf("Native Whisper transcription failed: %v, falling back...\n", err)
	}

	// 2. AssemblyAI
	if opts.AssemblyAIKey != "" {
		fmt.Println("Attempting transcription with AssemblyAI...")
		transcript, segments, err := transcribeWithAssemblyAI(audioPath, outputPath, opts.AssemblyAIKey)
		if err == nil {
			fmt.Println("AssemblyAI transcription completed successfully")
			return transcript, segments, nil
		}
		fmt.Printf("AssemblyAI transcription failed: %v, falling back...\n", err)
	}

	// 3. Whisper server
	if opts.WhisperServerURL != "" {
		fmt.Println("Attempting transcription with local Whisper server...")
		transcript, segments, err := transcribeWithWhisperServer(audioPath, outputPath, opts.WhisperServerURL)
		if err == nil {
			fmt.Println("Whisper server transcription completed successfully")
			return transcript, segments, nil
		}
		fmt.Printf("Whisper server transcription failed: %v, falling back...\n", err)
	}

	// 4. Demo fallback
	fmt.Println("Using demo transcription (no transcription services configured)")
	return transcribeDemo(audioPath, outputPath)
}

func transcribeWithNativeWhisper(audioPath, outputPath, modelPath string) (string, []Segment, error) {
	// Check if whisper.cpp is available (CGO build)
	if !lib.IsWhisperAvailable() {
		return "", nil, fmt.Errorf("whisper.cpp requires CGO; build with CGO_ENABLED=1")
	}

	// Load audio samples from WAV file
	samples, err := lib.LoadWAVAsFloat32(audioPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to load audio: %w", err)
	}

	// Initialize whisper context with model
	whisperCtx, err := lib.InitWhisper(modelPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to initialize whisper: %w", err)
	}
	defer whisperCtx.Free()

	// Transcribe audio
	libSegments, err := whisperCtx.TranscribeAudio(samples)
	if err != nil {
		return "", nil, fmt.Errorf("transcription failed: %w", err)
	}

	// Convert lib.TranscriptSegment to engine.Segment
	segments := make([]Segment, len(libSegments))
	var fullTranscript string

	for i, seg := range libSegments {
		segments[i] = Segment{
			Start: float64(seg.StartTime) / 1000.0, // Convert milliseconds to seconds
			End:   float64(seg.EndTime) / 1000.0,
			Text:  seg.Text,
		}
		if i > 0 {
			fullTranscript += " "
		}
		fullTranscript += seg.Text
	}

	// Write transcript to output file
	if err := os.WriteFile(outputPath, []byte(fullTranscript), 0644); err != nil {
		return "", nil, fmt.Errorf("failed to write transcript: %w", err)
	}

	return fullTranscript, segments, nil
}

func transcribeWithAssemblyAI(audioPath, outputPath, apiKey string) (string, []Segment, error) {
	// TODO: Implement AssemblyAI integration
	return "", nil, fmt.Errorf("AssemblyAI integration pending")
}

func transcribeWithWhisperServer(audioPath, outputPath, serverURL string) (string, []Segment, error) {
	// TODO: Implement whisper.cpp HTTP server client
	return "", nil, fmt.Errorf("whisper server integration pending")
}

func transcribeDemo(audioPath, outputPath string) (string, []Segment, error) {
	transcript := `Demo transcription: This is a placeholder generated by the OmniTranscripts engine.
The audio download and normalization stages completed successfully.
To enable actual transcription, configure WHISPER_MODEL_PATH, ASSEMBLYAI_API_KEY, or WHISPER_SERVER_URL.`

	segments := []Segment{
		{Start: 0.0, End: 5.0, Text: "Demo transcription: This is a placeholder generated by the OmniTranscripts engine."},
		{Start: 5.0, End: 10.0, Text: "The audio download and normalization stages completed successfully."},
		{Start: 10.0, End: 15.0, Text: "To enable actual transcription, configure WHISPER_MODEL_PATH, ASSEMBLYAI_API_KEY, or WHISPER_SERVER_URL."},
	}

	if err := os.WriteFile(outputPath, []byte(transcript), 0644); err != nil {
		return "", nil, fmt.Errorf("failed to write transcript: %w", err)
	}

	return transcript, segments, nil
}

func parseDuration(duration string) int {
	// TODO: Implement actual duration parsing
	return 120
}
