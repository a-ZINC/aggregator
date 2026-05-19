package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-ZINC/aggregator/config"
	"github.com/a-ZINC/aggregator/internal/db"
	"github.com/a-ZINC/aggregator/internal/filter"
	"github.com/a-ZINC/aggregator/internal/notifier"
	"github.com/a-ZINC/aggregator/internal/orchestrator"
	"github.com/a-ZINC/aggregator/internal/store"
)

func main() {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	config := config.Load();
	dbConfig := db.Config{
		DbUrl: config.DbUrl,
		MaxOpenConns: config.MaxOpenConns,
		MaxIdleConns: config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
	}
	filterConfig := filter.Config{
		AllowKeywords: config.AllowKeywords,
		ExcludeKeywords: config.ExcludeKeywords,
	}

	db, err := db.NewConnectionPool(dbConfig)
	if err != nil {
		slog.Error("❌ Failed to connect to database", "error", err)
		return
	}
	jobstore := store.NewJobStore(db)
	notifier := notifier.NewTelegramNotifier(config.TelegramBotToken, config.TelegramChatId, config.RateLimitSeconds)
	filter := filter.NewKeywordEvaluator(filterConfig)

	orchestrator := orchestrator.NewOrchestrator(jobstore, filter, notifier, config.MaxFetchWorkers)


	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	interval := config.PollInterval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("🚀 Running initial startup job aggregation sequence...")
	orchestrator.Run()
	slog.Info("💤 Startup sequence complete. Scheduler entering standby state.", "interval", interval.String())

	for {
		select {
		case <-ticker.C:
			slog.Info("⏰ 12-Hour interval reached. Triggering aggregation cycle...")
			
			startTime := time.Now()
			
			orchestrator.Run()
			
			slog.Info("✅ Aggregation cycle finished successfully", 
				"duration", time.Since(startTime).Round(time.Second).String(),
			)
		case <-ctx.Done():
			slog.Info("👋 Shutdown signal received. Exiting gracefully...")
			return
		}
	}
	
}