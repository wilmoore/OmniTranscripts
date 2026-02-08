package mcp

import (
	"context"
	"fmt"
	"time"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"omnitranscripts/jobs"
	"omnitranscripts/lib"
	"omnitranscripts/models"
)

// Timeout for async job processing (matches handlers/transcribe.go)
const asyncProcessingTimeout = 30 * time.Minute

// TranscribeURLTool creates the MCP tool definition for transcribe_url.
func TranscribeURLTool() mcplib.Tool {
	return mcplib.NewTool("transcribe_url",
		mcplib.WithDescription(
			"Starts transcription of a video or audio URL. Returns a job_id immediately. "+
				"Processing happens in the background. For videos over 2 minutes, wait 1-2 minutes "+
				"then call get_transcription to check if complete. Longer videos take proportionally longer. "+
				"Supports YouTube, Vimeo, SoundCloud, direct media URLs, and 1000+ platforms via yt-dlp.",
		),
		mcplib.WithString("url",
			mcplib.Required(),
			mcplib.Description("URL of media to transcribe (YouTube, Vimeo, podcast, etc.)"),
		),
	)
}

// GetTranscriptionTool creates the MCP tool definition for get_transcription.
func GetTranscriptionTool() mcplib.Tool {
	return mcplib.NewTool("get_transcription",
		mcplib.WithDescription(
			"Check the status of a transcription job. Returns 'processing' if still running, "+
				"'complete' with the full transcript and segments, or 'error' with details. "+
				"Call this after transcribe_url to retrieve results.",
		),
		mcplib.WithString("job_id",
			mcplib.Required(),
			mcplib.Description("Job ID returned from transcribe_url"),
		),
	)
}

// HandleTranscribeURL handles the transcribe_url tool invocation.
func HandleTranscribeURL(ctx context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	url, err := request.RequireString("url")
	if err != nil {
		return mcplib.NewToolResultError("Missing required parameter: url"), nil
	}

	// Validate URL
	if !models.ValidateURL(url) {
		return mcplib.NewToolResultError("Invalid URL. Must be a valid HTTP/HTTPS URL"), nil
	}

	// Create job
	job := jobs.NewJobWithSource(url, jobs.SourceTypeURL, "")
	queue := jobs.GetQueue()
	queue.AddJob(job)

	// Get duration for better UX messaging (but don't fail if it fails)
	var durationSeconds int
	durationCtx, durationCancel := context.WithTimeout(ctx, 30*time.Second)
	defer durationCancel()

	duration, durationErr := lib.GetVideoDuration(durationCtx, url)
	if durationErr == nil {
		durationSeconds = duration
	}

	// Start async processing
	go processTranscriptionJob(job)

	// Build response message
	var message string
	if durationSeconds > 0 {
		estimate := estimateProcessingTime(durationSeconds)
		durationMinutes := durationSeconds / 60
		message = fmt.Sprintf(
			"Transcription started. For a %d-minute video, expect %s processing time. "+
				"Use get_transcription with job_id to check status.",
			durationMinutes, estimate,
		)
	} else {
		message = "Transcription started. Use get_transcription with job_id to check status."
	}

	output := TranscribeURLOutput{
		JobID:           job.ID,
		Status:          "processing",
		DurationSeconds: durationSeconds,
		Message:         message,
	}

	return mcplib.NewToolResultText(formatTranscribeURLOutput(output)), nil
}

// HandleGetTranscription handles the get_transcription tool invocation.
func HandleGetTranscription(ctx context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	jobID, err := request.RequireString("job_id")
	if err != nil {
		return mcplib.NewToolResultError("Missing required parameter: job_id"), nil
	}

	queue := jobs.GetQueue()
	job, err := queue.GetJob(jobID)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("Job not found: %s", jobID)), nil
	}

	output := GetTranscriptionOutput{
		JobID:  job.ID,
		Status: string(job.Status),
	}

	switch job.Status {
	case jobs.StatusPending:
		output.Progress = "Waiting to start"
		output.Message = "Job is queued. Check again in a moment."

	case jobs.StatusRunning:
		output.Progress = "Processing (downloading, normalizing, or transcribing)"
		output.Message = "Still processing. Check again in 1-2 minutes."

	case jobs.StatusComplete:
		output.Transcript = job.Transcript
		output.Segments = job.Segments
		output.WordCount = countWords(job.Transcript)
		if job.Meta != nil {
			output.DurationSeconds = int(job.Meta.ProcessingTimeMs / 1000)
		}
		output.Message = "Transcription complete."

	case jobs.StatusError:
		output.Error = job.Error
		output.Message = "Transcription failed."
	}

	return mcplib.NewToolResultText(formatGetTranscriptionOutput(output)), nil
}

// processTranscriptionJob runs the transcription pipeline for a job.
func processTranscriptionJob(job *jobs.Job) {
	queue := jobs.GetQueue()

	ctx, cancel := context.WithTimeout(context.Background(), asyncProcessingTimeout)
	defer cancel()

	job.MarkRunning()
	queue.UpdateJob(job)

	transcript, segments, err := lib.ProcessTranscription(ctx, job.URL, job.ID)
	if err != nil {
		job.MarkError(err)
		queue.UpdateJob(job)
		return
	}

	job.MarkComplete(transcript, segments)
	queue.UpdateJob(job)
}

// formatTranscribeURLOutput formats the output as a readable string for LLMs.
func formatTranscribeURLOutput(o TranscribeURLOutput) string {
	result := fmt.Sprintf("Job ID: %s\nStatus: %s\n", o.JobID, o.Status)
	if o.DurationSeconds > 0 {
		result += fmt.Sprintf("Duration: %d seconds (%d minutes)\n", o.DurationSeconds, o.DurationSeconds/60)
	}
	if o.Title != "" {
		result += fmt.Sprintf("Title: %s\n", o.Title)
	}
	result += fmt.Sprintf("\n%s", o.Message)
	return result
}

// formatGetTranscriptionOutput formats the output as a readable string for LLMs.
func formatGetTranscriptionOutput(o GetTranscriptionOutput) string {
	result := fmt.Sprintf("Job ID: %s\nStatus: %s\n", o.JobID, o.Status)

	if o.Progress != "" {
		result += fmt.Sprintf("Progress: %s\n", o.Progress)
	}

	if o.Error != "" {
		result += fmt.Sprintf("Error: %s\n", o.Error)
	}

	if o.Transcript != "" {
		result += fmt.Sprintf("\nWord Count: %d\n", o.WordCount)
		result += fmt.Sprintf("\n--- TRANSCRIPT ---\n%s\n--- END TRANSCRIPT ---\n", o.Transcript)

		if len(o.Segments) > 0 {
			result += fmt.Sprintf("\nSegments: %d timestamped segments available\n", len(o.Segments))
		}
	}

	if o.Message != "" {
		result += fmt.Sprintf("\n%s", o.Message)
	}

	return result
}
