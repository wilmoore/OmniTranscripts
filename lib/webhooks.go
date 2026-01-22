package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"omnitranscripts/models"
)

// WebhookPayload represents the data sent to webhook URLs
type WebhookPayload struct {
	Event     string           `json:"event"`
	JobID     string           `json:"job_id"`
	URL       string           `json:"url"`
	Status    string           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Data      *WebhookJobData  `json:"data,omitempty"`
	Error     string           `json:"error,omitempty"`
	Metadata  *WebhookMetadata `json:"metadata,omitempty"`
}

// WebhookJobData contains the job results
type WebhookJobData struct {
	Transcript    string           `json:"transcript,omitempty"`
	Segments      []models.Segment `json:"segments,omitempty"`
	SegmentCount  int              `json:"segment_count"`
	Duration      float64          `json:"duration_seconds"`
	SubtitleFiles *SubtitleFiles   `json:"subtitle_files,omitempty"`
}

// SubtitleFiles contains paths to generated subtitle files
type SubtitleFiles struct {
	SRTPath string `json:"srt_path,omitempty"`
	VTTPath string `json:"vtt_path,omitempty"`
	SRTURL  string `json:"srt_url,omitempty"`
	VTTURL  string `json:"vtt_url,omitempty"`
}

// WebhookMetadata contains processing metadata
type WebhookMetadata struct {
	ProcessingTimeMs int64  `json:"processing_time_ms"`
	AudioFormat      string `json:"audio_format"`
	WhisperModel     string `json:"whisper_model"`
	Language         string `json:"language"`
	WordTimestamps   bool   `json:"word_timestamps"`
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Timeout time.Duration     `json:"timeout"`
	Retries int               `json:"retries"`
	Events  []string          `json:"events"` // job.started, job.completed, job.failed
}

// WebhookManager handles webhook notifications
type WebhookManager struct {
	client     *http.Client
	config     WebhookConfig
	retryDelay time.Duration
}

// NewWebhookManager creates a new webhook manager
func NewWebhookManager(config WebhookConfig) *WebhookManager {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.Retries == 0 {
		config.Retries = 3
	}

	return &WebhookManager{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config:     config,
		retryDelay: 2 * time.Second,
	}
}

// SendJobStarted sends a webhook when a job starts processing
func (wm *WebhookManager) SendJobStarted(ctx context.Context, job *models.Job) error {
	if !wm.shouldSendEvent("job.started") {
		return nil
	}

	payload := WebhookPayload{
		Event:     "job.started",
		JobID:     job.ID,
		URL:       job.URL,
		Status:    string(job.Status),
		Timestamp: time.Now(),
		Metadata: &WebhookMetadata{
			AudioFormat:    "wav",
			WhisperModel:   "base.en",
			Language:       "en",
			WordTimestamps: true,
		},
	}

	return wm.sendWebhook(ctx, payload)
}

// SendJobCompleted sends a webhook when a job completes successfully
func (wm *WebhookManager) SendJobCompleted(ctx context.Context, job *models.Job, srtPath, vttPath string, processingTime time.Duration) error {
	if !wm.shouldSendEvent("job.completed") {
		return nil
	}

	duration := 0.0
	if len(job.Segments) > 0 {
		duration = job.Segments[len(job.Segments)-1].End
	}

	subtitleFiles := &SubtitleFiles{}
	if srtPath != "" {
		subtitleFiles.SRTPath = srtPath
		// In production, these would be public URLs
		subtitleFiles.SRTURL = fmt.Sprintf("https://api.videotranscript.app/files/%s.srt", job.ID)
	}
	if vttPath != "" {
		subtitleFiles.VTTPath = vttPath
		subtitleFiles.VTTURL = fmt.Sprintf("https://api.videotranscript.app/files/%s.vtt", job.ID)
	}

	payload := WebhookPayload{
		Event:     "job.completed",
		JobID:     job.ID,
		URL:       job.URL,
		Status:    string(job.Status),
		Timestamp: time.Now(),
		Data: &WebhookJobData{
			Transcript:    job.Transcript,
			Segments:      job.Segments,
			SegmentCount:  len(job.Segments),
			Duration:      duration,
			SubtitleFiles: subtitleFiles,
		},
		Metadata: &WebhookMetadata{
			ProcessingTimeMs: processingTime.Milliseconds(),
			AudioFormat:      "wav",
			WhisperModel:     "base.en",
			Language:         "en",
			WordTimestamps:   true,
		},
	}

	return wm.sendWebhook(ctx, payload)
}

// SendJobFailed sends a webhook when a job fails
func (wm *WebhookManager) SendJobFailed(ctx context.Context, job *models.Job, errorMsg string, processingTime time.Duration) error {
	if !wm.shouldSendEvent("job.failed") {
		return nil
	}

	payload := WebhookPayload{
		Event:     "job.failed",
		JobID:     job.ID,
		URL:       job.URL,
		Status:    string(job.Status),
		Timestamp: time.Now(),
		Error:     errorMsg,
		Metadata: &WebhookMetadata{
			ProcessingTimeMs: processingTime.Milliseconds(),
			AudioFormat:      "wav",
			WhisperModel:     "base.en",
			Language:         "en",
			WordTimestamps:   true,
		},
	}

	return wm.sendWebhook(ctx, payload)
}

// sendWebhook sends the webhook with retry logic
func (wm *WebhookManager) sendWebhook(ctx context.Context, payload WebhookPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= wm.config.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wm.retryDelay * time.Duration(attempt)):
				// Exponential backoff
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", wm.config.URL, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create webhook request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "OmniTranscripts/1.0")
		req.Header.Set("X-Webhook-Event", payload.Event)
		req.Header.Set("X-Webhook-Job-ID", payload.JobID)

		// Add custom headers
		for key, value := range wm.config.Headers {
			req.Header.Set(key, value)
		}

		resp, err := wm.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("webhook request failed (attempt %d): %w", attempt+1, err)
			continue
		}

		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("✅ Webhook sent successfully: %s (status: %d)\n", payload.Event, resp.StatusCode)
			return nil
		}

		lastErr = fmt.Errorf("webhook returned non-2xx status: %d (attempt %d)", resp.StatusCode, attempt+1)
	}

	fmt.Printf("❌ Webhook failed after %d attempts: %v\n", wm.config.Retries+1, lastErr)
	return lastErr
}

// shouldSendEvent checks if the event should be sent based on configuration
func (wm *WebhookManager) shouldSendEvent(event string) bool {
	if len(wm.config.Events) == 0 {
		return true // Send all events if none specified
	}

	for _, configEvent := range wm.config.Events {
		if configEvent == event || configEvent == "*" {
			return true
		}
	}

	return false
}

// ValidateWebhookConfig validates webhook configuration
func ValidateWebhookConfig(config WebhookConfig) error {
	if config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	// Basic URL validation
	if !strings.HasPrefix(config.URL, "http://") && !strings.HasPrefix(config.URL, "https://") {
		return fmt.Errorf("webhook URL must start with http:// or https://")
	}

	if config.Timeout < 0 {
		return fmt.Errorf("webhook timeout must be positive")
	}

	if config.Retries < 0 {
		return fmt.Errorf("webhook retries must be non-negative")
	}

	return nil
}

// WebhookTestPayload creates a test payload for webhook validation
func WebhookTestPayload(webhookURL string) WebhookPayload {
	return WebhookPayload{
		Event:     "webhook.test",
		JobID:     "test-job-id",
		URL:       "https://www.youtube.com/watch?v=test",
		Status:    "test",
		Timestamp: time.Now(),
		Data: &WebhookJobData{
			Transcript:   "This is a test webhook payload.",
			SegmentCount: 1,
			Duration:     5.0,
		},
		Metadata: &WebhookMetadata{
			ProcessingTimeMs: 1000,
			AudioFormat:      "wav",
			WhisperModel:     "base.en",
			Language:         "en",
			WordTimestamps:   true,
		},
	}
}
