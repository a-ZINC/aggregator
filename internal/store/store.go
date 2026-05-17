package store

import (
	"context"

	"github.com/a-ZINC/aggregator/internal/domain"
)

type JobStoreImpl interface {
	Exists(ctx context.Context, urlHash string) (bool, error)
	Save(ctx context.Context, job *domain.Job) error
	UpdateStatus(ctx context.Context, urlHash string, status domain.JobStatus) error
}

type Store struct {
	JobStore JobStoreImpl
}

func NewStore(jobStore JobStoreImpl) *Store {
	return &Store{
		JobStore: jobStore,
	}
}