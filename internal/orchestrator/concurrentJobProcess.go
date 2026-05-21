package orchestrator

import (
	"context"
	"log/slog"
	"sync"

	"github.com/a-ZINC/aggregator/internal/domain"
	"github.com/a-ZINC/aggregator/internal/filter"
	"github.com/a-ZINC/aggregator/internal/notifier"
	"github.com/a-ZINC/aggregator/internal/store"
)

type ConcurrentJobProcessor struct {
	jobStore     *store.JobStore
	filter       filter.Evaluator
	notifier     notifier.Notifier
	wg           sync.WaitGroup
	fetchWorkers int
}

func NewConcurrentJobProcessor(jobStore *store.JobStore, filter filter.Evaluator, notifier notifier.Notifier, fetchWorkers int) *ConcurrentJobProcessor {
	return &ConcurrentJobProcessor{
		jobStore:     jobStore,
		filter:       filter,
		fetchWorkers: fetchWorkers,
		notifier:     notifier,
		wg:           sync.WaitGroup{},
	}
}

func (jp *ConcurrentJobProcessor) ProcessJobs(ctx context.Context, jobsChannel chan []domain.Job) {
	slog.InfoContext(ctx, "🚀 Starting job processor pipeline")
	for batch := range jobsChannel {
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
				slog.ErrorContext(ctx, "❌ Error processing job", "jobHash", j.UrlHash, "error", err)
				jp.jobStore.UpdateStatus(ctx, j.UrlHash, domain.Expired)
				return
			}
		}(job)
	}
}

func (jp *ConcurrentJobProcessor) processJob(ctx context.Context, job domain.Job) error {
	exist, err := jp.jobStore.Exists(ctx, job.UrlHash)
	if err != nil {
		return err
	}
	if exist {
		slog.InfoContext(ctx, "⚠️ Job already exists, skipping", "url_hash", job.UrlHash)
		return nil
	}

	err = jp.jobStore.Save(ctx, &job)
	if err != nil {
		slog.ErrorContext(ctx, "❌ Error saving job", "jobHash", job.UrlHash, "error", err)
		return err
	}

	evaluation, err := jp.filter.Evaluate(ctx, &job)
	if err != nil {
		slog.ErrorContext(ctx, "❌ Error evaluating job", "jobHash", job.UrlHash, "error", err)
		return err
	}
	if evaluation.Eligible {
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
		slog.ErrorContext(ctx, "❌ Error notifying job", "jobHash", job.UrlHash, "error", err)
		return err
	}

	err = jp.jobStore.UpdateStatus(ctx, job.UrlHash, domain.Notified)
	if err != nil {
		slog.ErrorContext(ctx, "❌ Error updating job status", "jobHash", job.UrlHash, "error", err)
		return err
	}
	slog.InfoContext(ctx, "📢 Job notified successfully", "url_hash", job.UrlHash)
	return nil
}
