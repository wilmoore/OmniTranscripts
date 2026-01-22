package models

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TranscribeRequest struct {
	URL string `json:"url" validate:"required"`
}

type TranscribeResponse struct {
	JobID      string    `json:"job_id,omitempty"`
	Transcript string    `json:"transcript,omitempty"`
	Segments   []Segment `json:"segments,omitempty"`
}

// JobStatus represents the status of a transcription job
type JobStatus string

const (
	StatusPending  JobStatus = "pending"
	StatusRunning  JobStatus = "running"
	StatusComplete JobStatus = "complete"
	StatusError    JobStatus = "error"
)

// Job represents a transcription job
type Job struct {
	ID          string     `json:"id"`
	URL         string     `json:"url"`
	Status      JobStatus  `json:"status"`
	Transcript  string     `json:"transcript,omitempty"`
	Segments    []Segment  `json:"segments,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Segment represents a timestamped segment of transcribed text
type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// NewJob creates a new transcription job
func NewJob(url string) *Job {
	return &Job{
		ID:        uuid.New().String(),
		URL:       url,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
}

// MarkRunning marks the job as running
func (j *Job) MarkRunning() {
	j.Status = StatusRunning
}

// MarkComplete marks the job as complete with transcript and segments
func (j *Job) MarkComplete(transcript string, segments []Segment) {
	j.Status = StatusComplete
	j.Transcript = transcript
	j.Segments = segments
	now := time.Now()
	j.CompletedAt = &now
}

// MarkError marks the job as failed with an error
func (j *Job) MarkError(err error) {
	j.Status = StatusError
	j.Error = err.Error()
	now := time.Now()
	j.CompletedAt = &now
}

func ValidateURL(url string) bool {
	// OmniTranscripts supports 1000+ platforms via yt-dlp
	// Accept any valid HTTP/HTTPS URL with a host
	urlRegex := regexp.MustCompile(`^https?://[^/\s]+`)
	return urlRegex.MatchString(url)
}

// SupportedAudioExtensions lists audio file extensions supported for upload
var SupportedAudioExtensions = []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac"}

// SupportedVideoExtensions lists video file extensions supported for upload
var SupportedVideoExtensions = []string{".mp4", ".mkv", ".webm", ".avi", ".mov"}

// ValidateFileExtension checks if the filename has a supported audio/video extension
func ValidateFileExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}

	for _, supported := range SupportedAudioExtensions {
		if ext == supported {
			return true
		}
	}
	for _, supported := range SupportedVideoExtensions {
		if ext == supported {
			return true
		}
	}
	return false
}

// GetSupportedExtensions returns all supported file extensions
func GetSupportedExtensions() []string {
	return append(SupportedAudioExtensions, SupportedVideoExtensions...)
}

func LoadTranscript(filePath string) (string, []Segment, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	var transcript strings.Builder
	var segments []Segment

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if segment := parseWhisperLine(line); segment != nil {
			segments = append(segments, *segment)
			transcript.WriteString(segment.Text + " ")
		} else {
			transcript.WriteString(line + " ")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}

	return strings.TrimSpace(transcript.String()), segments, nil
}

func parseWhisperLine(line string) *Segment {
	timestampRegex := regexp.MustCompile(`^\[(\d+:\d+:\d+\.\d+) --> (\d+:\d+:\d+\.\d+)\]\s*(.*)`)
	matches := timestampRegex.FindStringSubmatch(line)

	if len(matches) != 4 {
		return nil
	}

	start := parseTimestamp(matches[1])
	end := parseTimestamp(matches[2])
	text := strings.TrimSpace(matches[3])

	if text == "" {
		return nil
	}

	return &Segment{
		Start: start,
		End:   end,
		Text:  text,
	}
}

func parseTimestamp(timestamp string) float64 {
	parts := strings.Split(timestamp, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return float64(hours*3600) + float64(minutes*60) + seconds
}
