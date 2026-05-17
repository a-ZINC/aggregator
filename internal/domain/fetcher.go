package domain

import (
	"context"
)

type Fetcher interface {
	Fetch(context context.Context) ([]Job, error)
	Name() string
}