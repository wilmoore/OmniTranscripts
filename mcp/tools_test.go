package mcp

import (
	"context"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"omnitranscripts/jobs"
)

func init() {
	// Initialize the job queue for tests
	jobs.Initialize()
}

func TestTranscribeURLTool_Definition(t *testing.T) {
	tool := TranscribeURLTool()

	if tool.Name != "transcribe_url" {
		t.Errorf("expected tool name 'transcribe_url', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("expected non-empty tool description")
	}
}

func TestGetTranscriptionTool_Definition(t *testing.T) {
	tool := GetTranscriptionTool()

	if tool.Name != "get_transcription" {
		t.Errorf("expected tool name 'get_transcription', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("expected non-empty tool description")
	}
}

func TestHandleTranscribeURL_MissingURL(t *testing.T) {
	ctx := context.Background()
	request := mcplib.CallToolRequest{
		Params: mcplib.CallToolParams{
			Name:      "transcribe_url",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := HandleTranscribeURL(ctx, request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Check that it's an error result
	if !result.IsError {
		t.Error("expected error result for missing URL")
	}
}

func TestHandleTranscribeURL_InvalidURL(t *testing.T) {
	ctx := context.Background()
	request := mcplib.CallToolRequest{
		Params: mcplib.CallToolParams{
			Name: "transcribe_url",
			Arguments: map[string]interface{}{
				"url": "not-a-valid-url",
			},
		},
	}

	result, err := HandleTranscribeURL(ctx, request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Check that it's an error result
	if !result.IsError {
		t.Error("expected error result for invalid URL")
	}
}

func TestHandleGetTranscription_MissingJobID(t *testing.T) {
	ctx := context.Background()
	request := mcplib.CallToolRequest{
		Params: mcplib.CallToolParams{
			Name:      "get_transcription",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := HandleGetTranscription(ctx, request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Check that it's an error result
	if !result.IsError {
		t.Error("expected error result for missing job_id")
	}
}

func TestHandleGetTranscription_NotFound(t *testing.T) {
	ctx := context.Background()
	request := mcplib.CallToolRequest{
		Params: mcplib.CallToolParams{
			Name: "get_transcription",
			Arguments: map[string]interface{}{
				"job_id": "non-existent-job-id",
			},
		},
	}

	result, err := HandleGetTranscription(ctx, request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Check that it's an error result
	if !result.IsError {
		t.Error("expected error result for non-existent job")
	}
}

func TestEstimateProcessingTime(t *testing.T) {
	tests := []struct {
		duration int
		want     string
	}{
		{60, "1-2 minutes"},
		{120, "1-2 minutes"},
		{300, "2-5 minutes"},
		{600, "2-5 minutes"},
		{1200, "5-10 minutes"},
		{1800, "5-10 minutes"},
		{3600, "10-15 minutes"},
	}

	for _, tt := range tests {
		got := estimateProcessingTime(tt.duration)
		if got != tt.want {
			t.Errorf("estimateProcessingTime(%d) = %s, want %s", tt.duration, got, tt.want)
		}
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"  hello   world  ", 2},
		{"one\ntwo\tthree", 3},
	}

	for _, tt := range tests {
		got := countWords(tt.input)
		if got != tt.want {
			t.Errorf("countWords(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestNewServer(t *testing.T) {
	s := NewServer()
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}
