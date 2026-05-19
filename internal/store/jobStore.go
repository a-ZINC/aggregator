package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/a-ZINC/aggregator/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobStore struct {
	db *pgxpool.Pool
}

func NewJobStore(db *pgxpool.Pool) *JobStore {
	return &JobStore{db: db}
}

func (s *JobStore) Close() {
	s.db.Close() // Note: pgxpool.Close() doesn't return an error
}

func (s *JobStore) Exists(ctx context.Context, urlHash string) (bool, error) {
	var exists bool
	// .QueryRowContext changes to .QueryRow (context is the first parameter)
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM jobs WHERE url_hash = $1)", urlHash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("store: exists check: %w", err)
	}
	return exists, nil
}

func (s *JobStore) Save(ctx context.Context, job *domain.Job) error {
	var validatedPayload string

	switch v := any(job.RawPayload).(type) {
	case string:
		validatedPayload = v
		if validatedPayload == "" {
			validatedPayload = "{}"
		}
	case []byte:
		validatedPayload = string(v)
		if validatedPayload == "" {
			validatedPayload = "{}"
		}
	default:
		// If RawPayload is a struct, map, or slice, explicitly marshal it to valid JSON
		bytes, err := json.Marshal(job.RawPayload)
		if err != nil {
			return fmt.Errorf("store: marshal raw payload: %w", err)
		}
		validatedPayload = string(bytes)
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO jobs (url_hash, title, company, location, description, apply_url, source, tags, salary_min, salary_max, status, posted_at, fetched_at, notified_at, raw_payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, job.UrlHash, job.Title, job.Company, job.Location, job.Description, job.ApplyUrl, job.Source, job.Tags, job.SalaryMin, job.SalaryMax, job.Status, job.PostedAt, job.FetchedAt, job.NotifiedAt, validatedPayload)
	if err != nil {
		return fmt.Errorf("store: insert job: %w", err)
	}
	return nil
}

func (s *JobStore) UpdateStatus(ctx context.Context, urlHash string, status domain.JobStatus) error {
	if status == domain.Notified {
		_, err := s.db.Exec(ctx, "UPDATE jobs SET status = $1, notified_at = $2 WHERE url_hash = $3", status, time.Now().UTC(), urlHash)
		if err != nil {
			return fmt.Errorf("store: update job status: %w", err)
		}
		return nil
	}
	_, err := s.db.Exec(ctx, "UPDATE jobs SET status = $1 WHERE url_hash = $2", status, urlHash)
	if err != nil {
		return fmt.Errorf("store: update job status: %w", err)
	}
	return nil
}
