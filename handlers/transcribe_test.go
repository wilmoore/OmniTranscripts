package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"omnitranscripts/jobs"
	"omnitranscripts/models"
)

func setupTestApp() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	jobs.Initialize()

	app.Post("/transcribe", PostTranscribe)
	app.Get("/transcribe/:job_id", GetTranscribeJob)

	return app
}

func TestPostTranscribe_ValidationErrors(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name         string
		body         interface{}
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "Invalid JSON",
			body:         "invalid json",
			expectedCode: 400,
			expectedMsg:  "Invalid request body",
		},
		{
			name:         "Missing URL",
			body:         map[string]string{},
			expectedCode: 400,
			expectedMsg:  "Invalid YouTube URL",
		},
		{
			name:         "Invalid URL",
			body:         map[string]string{"url": "not-a-youtube-url"},
			expectedCode: 400,
			expectedMsg:  "Invalid YouTube URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			if str, ok := tt.body.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/transcribe", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-key")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result["error"], tt.expectedMsg)
		})
	}
}

func TestGetTranscribeJob_NotFound(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/transcribe/non-existent-id", nil)
	req.Header.Set("Authorization", "Bearer test-key")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, 404, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "Job not found", result["error"])
}

func TestJobQueue_Operations(t *testing.T) {
	jobs.Initialize()
	queue := jobs.GetQueue()

	job := jobs.NewJob("https://youtube.com/watch?v=test")

	queue.AddJob(job)

	retrievedJob, err := queue.GetJob(job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.ID, retrievedJob.ID)
	assert.Equal(t, jobs.StatusPending, retrievedJob.Status)

	job.MarkRunning()
	queue.UpdateJob(job)

	retrievedJob, err = queue.GetJob(job.ID)
	require.NoError(t, err)
	assert.Equal(t, jobs.StatusRunning, retrievedJob.Status)

	segments := []models.Segment{
		{Start: 0.0, End: 5.0, Text: "Test segment"},
	}
	job.MarkComplete("Test transcript", segments)
	queue.UpdateJob(job)

	retrievedJob, err = queue.GetJob(job.ID)
	require.NoError(t, err)
	assert.Equal(t, jobs.StatusComplete, retrievedJob.Status)
	assert.Equal(t, "Test transcript", retrievedJob.Transcript)
	assert.Len(t, retrievedJob.Segments, 1)
}

// Benchmark tests
func BenchmarkPostTranscribe_Validation(b *testing.B) {
	app := setupTestApp()

	reqBody, _ := json.Marshal(models.TranscribeRequest{
		URL: "https://youtube.com/watch?v=test",
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/transcribe", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")

		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}
}

func BenchmarkJobQueue_AddAndRetrieve(b *testing.B) {
	jobs.Initialize()
	queue := jobs.GetQueue()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job := jobs.NewJob("https://youtube.com/watch?v=test")
		queue.AddJob(job)

		_, err := queue.GetJob(job.ID)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkJobQueue_Update(b *testing.B) {
	jobs.Initialize()
	queue := jobs.GetQueue()

	job := jobs.NewJob("https://youtube.com/watch?v=test")
	queue.AddJob(job)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job.MarkRunning()
		queue.UpdateJob(job)

		job.MarkComplete("Test transcript", []models.Segment{})
		queue.UpdateJob(job)
	}
}

func BenchmarkJobQueue_ConcurrentAccess(b *testing.B) {
	jobs.Initialize()
	queue := jobs.GetQueue()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			job := jobs.NewJob("https://youtube.com/watch?v=test")
			queue.AddJob(job)

			_, err := queue.GetJob(job.ID)
			if err != nil {
				b.Error(err)
			}

			job.MarkComplete("Test", []models.Segment{})
			queue.UpdateJob(job)
		}
	})
}

// Performance test for API endpoints
func BenchmarkAPI_HealthEndpoint(b *testing.B) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "OmniTranscripts API is running",
		})
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}
}

// Load test simulation
func TestAPI_LoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	app := setupTestApp()

	concurrency := 10
	requestsPerWorker := 100

	reqBody, _ := json.Marshal(models.TranscribeRequest{
		URL: "https://youtube.com/watch?v=test",
	})

	startTime := time.Now()

	type result struct {
		statusCode int
		duration   time.Duration
		err        error
	}

	results := make(chan result, concurrency*requestsPerWorker)

	for worker := 0; worker < concurrency; worker++ {
		go func() {
			for req := 0; req < requestsPerWorker; req++ {
				start := time.Now()

				httpReq := httptest.NewRequest(http.MethodPost, "/transcribe", bytes.NewReader(reqBody))
				httpReq.Header.Set("Content-Type", "application/json")
				httpReq.Header.Set("Authorization", "Bearer test-key")

				resp, err := app.Test(httpReq, 5000) // 5 second timeout
				duration := time.Since(start)

				statusCode := 0
				if resp != nil {
					statusCode = resp.StatusCode
					resp.Body.Close()
				}

				results <- result{
					statusCode: statusCode,
					duration:   duration,
					err:        err,
				}
			}
		}()
	}

	totalRequests := concurrency * requestsPerWorker
	successCount := 0
	var totalDuration time.Duration
	var maxDuration time.Duration

	for i := 0; i < totalRequests; i++ {
		res := <-results

		if res.err == nil && res.statusCode > 0 {
			successCount++
		}

		totalDuration += res.duration
		if res.duration > maxDuration {
			maxDuration = res.duration
		}
	}

	testDuration := time.Since(startTime)
	avgDuration := totalDuration / time.Duration(totalRequests)
	requestsPerSecond := float64(totalRequests) / testDuration.Seconds()

	t.Logf("Load Test Results:")
	t.Logf("  Total Requests: %d", totalRequests)
	t.Logf("  Successful: %d (%.1f%%)", successCount, float64(successCount)/float64(totalRequests)*100)
	t.Logf("  Test Duration: %v", testDuration)
	t.Logf("  Requests/sec: %.1f", requestsPerSecond)
	t.Logf("  Avg Response Time: %v", avgDuration)
	t.Logf("  Max Response Time: %v", maxDuration)

	assert.True(t, float64(successCount)/float64(totalRequests) > 0.95, "Success rate should be > 95%")
	assert.True(t, avgDuration < 100*time.Millisecond, "Average response time should be < 100ms")
}
