package orchestrator

import (
	"context"
	"log"
	"log/slog"
	"sync"

	"github.com/a-ZINC/aggregator/internal/domain"
	"github.com/a-ZINC/aggregator/internal/filter"
	"github.com/a-ZINC/aggregator/internal/notifier"
	"github.com/a-ZINC/aggregator/internal/store"
	"golang.org/x/sync/semaphore"
)

type ConcurrentJobProcessor struct {
	jobsChannel chan []domain.Job
	jobStore *store.JobStore
	filter filter.Evaluator
	notifier notifier.Notifier
	wg sync.WaitGroup
	fetchWorkers int
}

func NewConcurrentJobProcessor(jobsChannel chan []domain.Job, jobStore *store.JobStore, filter filter.Evaluator, notifier notifier.Notifier, fetchWorkers int) *ConcurrentJobProcessor {
	return &ConcurrentJobProcessor{
		jobsChannel: jobsChannel,
		jobStore: jobStore,
		filter: filter,
		fetchWorkers: fetchWorkers,
		notifier: notifier,
		wg: sync.WaitGroup{},
	}
}

func (jp *ConcurrentJobProcessor) ProcessJobs(ctx context.Context) {
	slog.InfoContext(ctx, "🚀 Starting job processor pipeline")
	for batch := range jp.jobsChannel {
		slog.DebugContext(ctx, "📥 Received batch from channel", "batch_size", len(batch))
		jp.wg.Add(1)
		go func(jobs []domain.Job) {
			defer jp.wg.Done()
			jp.processBatch(ctx, jobs)
		}(batch)
	}
	jp.wg.Wait()
	slog.InfoContext(ctx, "🏁 Job processor pipeline completed cleanly")
}

func (jp *ConcurrentJobProcessor) processBatch(ctx context.Context, jobs []domain.Job) {
	semaphore := make(chan struct{}, jp.fetchWorkers)
	for _, job := range jobs {
		jp.wg.Add(1)
		semaphore <- struct{}{}
		go func(j domain.Job) {
			defer jp.wg.Done()
			defer func() { <-semaphore }()
			err := jp.processJob(ctx, j)
			if err != nil {
				slog.ErrorContext(ctx, "❌ Error processing job", "job", j, "error", err)
				jp.jobStore.UpdateStatus(ctx, j.UrlHash, domain.Expired)
				return
			}
		}(job)
	}
}

func (jp *ConcurrentJobProcessor) processJob(ctx context.Context, job domain.Job) (error) {
	evaluation, err := jp.filter.Evaluate(nil, &job)
	if err != nil {
		return err
	}
	if evaluation.Eligible {
		err := jp.jobStore.Save(ctx, &job)
		if err != nil {
			return err
		}
		slog.InfoContext(ctx, "💾 Job evaluated as eligible and saved", "url_hash", job.UrlHash)
	} else {
		slog.InfoContext(ctx, "⏭ Job evaluated as ineligible", "url_hash", job.UrlHash)
		err := jp.jobStore.UpdateStatus(ctx, job.UrlHash, domain.Skipped)
		if err != nil {
			return err
		}
		return nil
	}

	err = jp.notifier.Notify(ctx, &job, evaluation.Reason)
	if err != nil {
		slog.ErrorContext(ctx, "❌ Error notifying job", "job", job, "error", err)
		return err
	}

	err = jp.jobStore.UpdateStatus(ctx, job.UrlHash, domain.Notified)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "📢 Job notified successfully", "url_hash", job.UrlHash)
	return nil
}