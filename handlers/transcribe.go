package handlers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"omnitranscripts/config"
	"omnitranscripts/jobs"
	"omnitranscripts/lib"
	"omnitranscripts/models"
)

func PostTranscribe(c *fiber.Ctx) error {
	contentType := string(c.Request().Header.ContentType())

	// Handle multipart file upload
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return handleFileUpload(c)
	}

	// Handle JSON URL request (existing flow)
	return handleURLTranscribe(c)
}

func handleFileUpload(c *fiber.Ctx) error {
	cfg := config.Load()

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided. Use 'file' field in multipart/form-data",
		})
	}

	// Validate file extension
	if !models.ValidateFileExtension(file.Filename) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":           "Unsupported file type",
			"supported_audio": models.SupportedAudioExtensions,
			"supported_video": models.SupportedVideoExtensions,
		})
	}

	// Check file size
	if file.Size > cfg.MaxUploadSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "File too large",
			"max_size": cfg.MaxUploadSize,
		})
	}

	// Create work directory
	if err := os.MkdirAll(cfg.WorkDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create work directory",
		})
	}

	// Create job first to get ID for filename
	ext := filepath.Ext(file.Filename)
	job := jobs.NewJobWithSource(
		fmt.Sprintf("file://%s", file.Filename),
		jobs.SourceTypeFile,
		strings.TrimPrefix(ext, "."),
	)
	queue := jobs.GetQueue()
	queue.AddJob(job)

	// Save uploaded file with job ID prefix
	uploadPath := filepath.Join(cfg.WorkDir, fmt.Sprintf("%s_upload%s", job.ID, ext))

	src, err := file.Open()
	if err != nil {
		job.MarkError(fmt.Errorf("failed to open uploaded file"))
		queue.UpdateJob(job)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process uploaded file",
		})
	}
	defer src.Close()

	dst, err := os.Create(uploadPath)
	if err != nil {
		job.MarkError(fmt.Errorf("failed to save uploaded file"))
		queue.UpdateJob(job)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save uploaded file",
		})
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		job.MarkError(fmt.Errorf("failed to write uploaded file"))
		queue.UpdateJob(job)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to write uploaded file",
		})
	}

	// Update job URL to the local file path
	job.URL = uploadPath

	// Process file uploads asynchronously (file size unknown in terms of duration)
	go processFileTranscription(job, uploadPath)

	return c.JSON(models.TranscribeResponse{
		JobID: job.ID,
	})
}

func handleURLTranscribe(c *fiber.Ctx) error {
	var req models.TranscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if !models.ValidateURL(req.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL. Must be a valid HTTP/HTTPS URL",
		})
	}

	// Extract format from URL if possible
	urlFormat := ""
	if idx := strings.LastIndex(req.URL, "."); idx != -1 {
		urlFormat = strings.TrimPrefix(req.URL[idx:], ".")
		// Only keep if it looks like a media format
		if len(urlFormat) > 5 || strings.Contains(urlFormat, "/") {
			urlFormat = ""
		}
	}

	job := jobs.NewJobWithSource(req.URL, jobs.SourceTypeURL, urlFormat)
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

	if job.Meta != nil {
		response["meta"] = job.Meta
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

func processFileTranscription(job *jobs.Job, uploadPath string) {
	queue := jobs.GetQueue()

	// Clean up uploaded file when done
	defer os.Remove(uploadPath)

	job.MarkRunning()
	queue.UpdateJob(job)

	// For local files, we pass the file path directly to yt-dlp
	// yt-dlp handles local files natively
	transcript, segments, err := lib.ProcessTranscription(uploadPath, job.ID)
	if err != nil {
		job.MarkError(err)
		queue.UpdateJob(job)
		return
	}

	job.MarkComplete(transcript, segments)
	queue.UpdateJob(job)
}
