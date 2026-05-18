package filter

import (
	"context"

	"github.com/a-ZINC/aggregator/internal/domain"
)

type Evaluation struct {
	Eligible bool
	Reason  string
}
type Evaluator interface {
	Evaluate(ctx context.Context, job *domain.Job) (Evaluation, error)
}
