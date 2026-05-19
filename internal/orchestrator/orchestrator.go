package orchestrator

import (
	"context"

	"github.com/a-ZINC/aggregator/internal/domain"
	"github.com/a-ZINC/aggregator/internal/filter"
	"github.com/a-ZINC/aggregator/internal/notifier"
	"github.com/a-ZINC/aggregator/internal/store"
)

type Orchestrator struct {
	fetcher *ConcurrentFetcher
	processor *ConcurrentJobProcessor
	jobStore *store.JobStore
	filter filter.Evaluator
	notifier notifier.Notifier
	jobsChannel chan []domain.Job
}

func NewOrchestrator(jobStore *store.JobStore, filter filter.Evaluator, notifier notifier.Notifier) *Orchestrator {
	jobsChannel := make(chan []domain.Job, 10) // Buffered channel to handle bursts of job batches
	return &Orchestrator{
		fetcher: NewConcurrentFetcher(jobsChannel),
		processor: NewConcurrentJobProcessor(jobsChannel, jobStore, filter, notifier, 5),
	}
}

func (o *Orchestrator) Run() {
	// 1. Fetch jobs concurrently
	go o.fetcher.FetchAll()

	// 2. Process fetched jobs
	o.processor.ProcessJobs(context.Background())
}