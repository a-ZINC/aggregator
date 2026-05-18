package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/a-ZINC/aggregator/internal/domain"
	"github.com/lib/pq"
)

type JobStore struct {
	db *sql.DB
}

func NewJobStore(db *sql.DB) *JobStore {
	return &JobStore{db: db}
}

func (s *JobStore) Close() error {
	return s.db.Close()
}

func (s *JobStore) Exists(ctx context.Context, urlHash string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM jobs WHERE url_hash = $1)", urlHash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("store: exists check: %w", err)
	}
	return exists, nil
}

func (s *JobStore) Save(ctx context.Context, job *domain.Job) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO jobs (url_hash, title, company, location, description, apply_url, source, tags, salary_min, salary_max, status, posted_at, fetched_at, notified_at, raw_payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, job.UrlHash, job.Title, job.Company, job.Location, job.Description, job.ApplyUrl, job.Source, pq.Array(job.Tags), job.SalaryMin, job.SalaryMax, job.Status, job.PostedAt, job.FetchedAt, job.NotifiedAt, job.RawPayload)
	if err != nil {
		return fmt.Errorf("store: insert job: %w", err)
	}
	return nil
}

func (s *JobStore) UpdateStatus(ctx context.Context, urlHash string, status domain.JobStatus) error {
	if status == domain.Notified {
		_, err := s.db.ExecContext(ctx, "UPDATE jobs SET status = $1, notified_at = $2 WHERE url_hash = $3", status, time.Now().UTC(), urlHash)
		if err != nil {
			return fmt.Errorf("store: update job status: %w", err)
		}
		return nil
	}
	_, err := s.db.ExecContext(ctx, "UPDATE jobs SET status = $1 WHERE url_hash = $2", status, urlHash)
	if err != nil {
		return fmt.Errorf("store: update job status: %w", err)
	}
	return nil
}

