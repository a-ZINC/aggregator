package main

import (
	"log/slog"
	"os"

	"github.com/a-ZINC/aggregator/config"
	"github.com/a-ZINC/aggregator/internal/db"
	"github.com/a-ZINC/aggregator/internal/domain"
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
	notifier := notifier.NewTelegramNotifier(config.TelegramBotToken, config.TelegramChatId)
	filter := filter.NewKeywordEvaluator(filterConfig)

	jobsChannel := make(chan []domain.Job, 100)
	orchestrator := orchestrator.NewOrchestrator(jobstore, filter, notifier)
orchestrator.Run()
}