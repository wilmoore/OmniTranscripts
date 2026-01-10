package jobs

import (
	"time"

	"github.com/google/uuid"

	"omnitranscripts/models"
)

type JobStatus string

const (
	StatusPending  JobStatus = "pending"
	StatusRunning  JobStatus = "running"
	StatusComplete JobStatus = "complete"
	StatusError    JobStatus = "error"
)

type Job struct {
	ID          string           `json:"id"`
	URL         string           `json:"url"`
	Status      JobStatus        `json:"status"`
	Transcript  string           `json:"transcript,omitempty"`
	Segments    []models.Segment `json:"segments,omitempty"`
	Error       string           `json:"error,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

func NewJob(url string) *Job {
	return &Job{
		ID:        uuid.New().String(),
		URL:       url,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
}

func (j *Job) MarkRunning() {
	j.Status = StatusRunning
}

func (j *Job) MarkComplete(transcript string, segments []models.Segment) {
	j.Status = StatusComplete
	j.Transcript = transcript
	j.Segments = segments
	now := time.Now()
	j.CompletedAt = &now
}

func (j *Job) MarkError(err error) {
	j.Status = StatusError
	j.Error = err.Error()
	now := time.Now()
	j.CompletedAt = &now
}

func (j *Job) IsComplete() bool {
	return j.Status == StatusComplete || j.Status == StatusError
}
