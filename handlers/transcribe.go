package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"omnitranscripts/jobs"
	"omnitranscripts/lib"
	"omnitranscripts/models"
)

func PostTranscribe(c *fiber.Ctx) error {
	var req models.TranscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if !models.ValidateURL(req.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid YouTube URL",
		})
	}

	job := jobs.NewJob(req.URL)
	queue := jobs.GetQueue()
	queue.AddJob(job)

	duration, err := lib.GetVideoDuration(req.URL)
	if err != nil {
		job.MarkError(err)
		queue.UpdateJob(job)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to get video information",
		})
	}

	if duration <= 120 {
		go processTranscriptionSync(job)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return c.JSON(models.TranscribeResponse{
					JobID: job.ID,
				})
			default:
				currentJob, _ := queue.GetJob(job.ID)
				if currentJob.IsComplete() {
					if currentJob.Status == jobs.StatusError {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"error": currentJob.Error,
						})
					}
					return c.JSON(models.TranscribeResponse{
						Transcript: currentJob.Transcript,
						Segments:   currentJob.Segments,
					})
				}
				time.Sleep(1 * time.Second)
			}
		}
	} else {
		go processTranscriptionAsync(job)
		return c.JSON(models.TranscribeResponse{
			JobID: job.ID,
		})
	}
}

func GetTranscribeJob(c *fiber.Ctx) error {
	jobID := c.Params("job_id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Job ID is required",
		})
	}

	queue := jobs.GetQueue()
	job, err := queue.GetJob(jobID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	response := fiber.Map{
		"id":         job.ID,
		"status":     job.Status,
		"created_at": job.CreatedAt,
	}

	if job.Status == jobs.StatusComplete {
		response["transcript"] = job.Transcript
		response["segments"] = job.Segments
		response["completed_at"] = job.CompletedAt
	} else if job.Status == jobs.StatusError {
		response["error"] = job.Error
		response["completed_at"] = job.CompletedAt
	}

	return c.JSON(response)
}

func processTranscriptionSync(job *jobs.Job) {
	processTranscription(job)
}

func processTranscriptionAsync(job *jobs.Job) {
	processTranscription(job)
}

func processTranscription(job *jobs.Job) {
	queue := jobs.GetQueue()

	job.MarkRunning()
	queue.UpdateJob(job)

	transcript, segments, err := lib.ProcessTranscription(job.URL, job.ID)
	if err != nil {
		job.MarkError(err)
		queue.UpdateJob(job)
		return
	}

	job.MarkComplete(transcript, segments)
	queue.UpdateJob(job)
}
