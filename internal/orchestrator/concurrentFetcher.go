package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/a-ZINC/aggregator/internal/domain"
)

type ConcurrentFetcher struct {
	fetchers []domain.Fetcher
	wg sync.WaitGroup
}

func NewConcurrentFetcher() *ConcurrentFetcher {
	return &ConcurrentFetcher{
		fetchers: []domain.Fetcher{},
		wg: sync.WaitGroup{},
	}
}

func (cf *ConcurrentFetcher) AddFetcher(f domain.Fetcher) {
	cf.fetchers = append(cf.fetchers, f)
}

func (cf *ConcurrentFetcher) FetchAll(jobsChannel chan []domain.Job) (error) {
	defer close(jobsChannel)
	var wg sync.WaitGroup
	for _, fetcher := range cf.fetchers {
		wg.Add(1)
		go cf.fetchAndSend(fetcher, jobsChannel, &wg)
	}
	wg.Wait()
	return nil
}

func (cf *ConcurrentFetcher) fetchAndSend(fetcher domain.Fetcher, jobs chan <- []domain.Job, wg *sync.WaitGroup) {
	defer wg.Done()
	fetchedJobs, err := fetcher.Fetch(context.Background())
	if err != nil {
		fmt.Printf("Error fetching from %T: %v\n", fetcher, err)
		return
	}
	jobs <- fetchedJobs
}
