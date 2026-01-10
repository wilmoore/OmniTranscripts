//go:build encore

package transcribe

import (
	"context"
	"encoding/json"

	"encore.dev/storage/sqldb"

	"omnitranscripts/models"
)

// Database for job storage
var db = sqldb.NewDatabase("jobs", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// storeJob stores a job in the database.
func storeJob(ctx context.Context, job *models.Job) error {
	segmentsJSON, err := json.Marshal(job.Segments)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO jobs (id, url, status, transcript, segments, error, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = db.Exec(ctx, query,
		job.ID, job.URL, job.Status, job.Transcript,
		segmentsJSON, job.Error, job.CreatedAt, job.CompletedAt,
	)
	return err
}

// getJob retrieves a job from the database.
func getJob(ctx context.Context, id string) (*models.Job, error) {
	query := `
		SELECT id, url, status, transcript, segments, error, created_at, completed_at
		FROM jobs WHERE id = $1
	`

	var job models.Job
	var segmentsJSON []byte

	err := db.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.URL, &job.Status, &job.Transcript,
		&segmentsJSON, &job.Error, &job.CreatedAt, &job.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(segmentsJSON) > 0 {
		if err := json.Unmarshal(segmentsJSON, &job.Segments); err != nil {
			return nil, err
		}
	}

	return &job, nil
}

// updateJob updates a job in the database.
func updateJob(ctx context.Context, job *models.Job) error {
	segmentsJSON, err := json.Marshal(job.Segments)
	if err != nil {
		return err
	}

	query := `
		UPDATE jobs
		SET status = $2, transcript = $3, segments = $4, error = $5, completed_at = $6
		WHERE id = $1
	`

	_, err = db.Exec(ctx, query,
		job.ID, job.Status, job.Transcript,
		segmentsJSON, job.Error, job.CompletedAt,
	)
	return err
}