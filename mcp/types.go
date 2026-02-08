// Package mcp provides an MCP (Model Context Protocol) server for OmniTranscripts,
// enabling integration with ChatGPT via the OpenAI Apps SDK.
package mcp

import (
	"fmt"

	"omnitranscripts/models"
)

// TranscribeURLInput is the input schema for the transcribe_url tool.
type TranscribeURLInput struct {
	URL string `json:"url" jsonschema:"description=URL of media to transcribe (YouTube, Vimeo, podcast, etc.),required"`
}

// TranscribeURLOutput is the output for the transcribe_url tool.
type TranscribeURLOutput struct {
	JobID           string `json:"job_id"`
	Status          string `json:"status"`
	Title           string `json:"title,omitempty"`
	DurationSeconds int    `json:"duration_seconds,omitempty"`
	Message         string `json:"message"`
}

// GetTranscriptionInput is the input schema for the get_transcription tool.
type GetTranscriptionInput struct {
	JobID string `json:"job_id" jsonschema:"description=Job ID returned from transcribe_url,required"`
}

// GetTranscriptionOutput is the output for the get_transcription tool.
type GetTranscriptionOutput struct {
	JobID           string           `json:"job_id"`
	Status          string           `json:"status"`
	Progress        string           `json:"progress,omitempty"`
	Message         string           `json:"message,omitempty"`
	Transcript      string           `json:"transcript,omitempty"`
	Segments        []models.Segment `json:"segments,omitempty"`
	WordCount       int              `json:"word_count,omitempty"`
	DurationSeconds int              `json:"duration_seconds,omitempty"`
	Error           string           `json:"error,omitempty"`
}

// estimateProcessingTime returns a human-readable estimate based on duration.
func estimateProcessingTime(durationSeconds int) string {
	if durationSeconds <= 120 {
		return "1-2 minutes"
	}
	if durationSeconds <= 600 {
		return "2-5 minutes"
	}
	if durationSeconds <= 1800 {
		return "5-10 minutes"
	}
	if durationSeconds <= 3600 {
		return "10-15 minutes"
	}
	// For very long content, estimate ~1 min processing per 5 min of content
	estimatedMinutes := durationSeconds / 300
	if estimatedMinutes < 15 {
		estimatedMinutes = 15
	}
	return fmt.Sprintf("%d-%d minutes", estimatedMinutes, estimatedMinutes+10)
}

// countWords returns the number of words in a string.
func countWords(s string) int {
	if s == "" {
		return 0
	}
	words := 0
	inWord := false
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			inWord = false
		} else if !inWord {
			inWord = true
			words++
		}
	}
	return words
}
