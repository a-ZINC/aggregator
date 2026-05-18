package notifier

import (
	"context"

	"github.com/a-ZINC/aggregator/internal/domain"
)

type Notifier interface {
	Notify(ctx context.Context, job *domain.Job, reason string) error
}