package orchestrator

import (
	"context"

	"github.com/a-ZINC/aggregator/internal/domain"
	"github.com/a-ZINC/aggregator/internal/fetcher"
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

func NewOrchestrator(jobStore *store.JobStore, filter filter.Evaluator, notifier notifier.Notifier, fetchWorkers int) *Orchestrator { // Buffered channel to handle bursts of job batches
	return &Orchestrator{
		fetcher: NewConcurrentFetcher(),
		processor: NewConcurrentJobProcessor(jobStore, filter, notifier, fetchWorkers),
	}
}

func (o *Orchestrator) buildChain() {
	o.fetcher.AddFetcher(fetcher.NewRemoteOkFetcher())
	o.fetcher.AddFetcher(fetcher.NewRemotiveFetcher())
}

func (o *Orchestrator) Run() {
	jobsChannel := make(chan []domain.Job, 10)
	// 1. Make the chain of responsibility
	o.buildChain()
	// 1. Fetch jobs concurrently
	go o.fetcher.FetchAll(jobsChannel)

	// 2. Process fetched jobs
	o.processor.ProcessJobs(context.Background(), jobsChannel)
}