package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/a-ZINC/aggregator/internal/db"
)

type Config struct {
	DbUrl            string
	TelegramBotToken string
	TelegramChatId   string
	AllowKeywords    []string
	ExcludeKeywords  []string
	PollInterval     time.Duration
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
	maxFetchWorkers   int
}

func getStringEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var intVal int
		_, err := fmt.Sscanf(val, "%d", &intVal)
		if err == nil {
			return intVal
		}
	}
	return defaultVal
}

// func getBoolEnv(key string, defaultVal bool) bool {
// 	if val := os.Getenv(key); val != "" {
// 		var boolVal bool
// 		_, err := fmt.Sscanf(val, "%t", &boolVal)
// 		if err == nil {
// 			return boolVal
// 		}
// 	}
// 	return defaultVal
// }

func getStringSliceEnv(key string, defaultVal []string) []string {
	if val := os.Getenv(key); val != "" {
		return splitAndTrim(val, ",")
	}
	return defaultVal
}

func splitAndTrim(s string, sep string) []string {
	var result []string
	for _, v := range strings.Split(s, sep) {
		result = append(result, strings.TrimSpace(v))
	}
	return result
}

func Load() *Config {
	var cfg Config

	cfg.PollInterval = time.Duration(getIntEnv("POLL_INTERVAL", 30)) * time.Minute
	cfg.DbUrl = getStringEnv("DB_URL", "sqlite:///jobs.db")
	cfg.MaxOpenConns = getIntEnv("DB_MAX_OPEN_CONNS", 10)
	cfg.MaxIdleConns = getIntEnv("DB_MAX_IDLE_CONNS", 5)
	cfg.ConnMaxLifetime = time.Duration(getIntEnv("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute
	cfg.TelegramBotToken = getStringEnv("TELEGRAM_BOT_TOKEN", "")
	cfg.TelegramChatId = getStringEnv("TELEGRAM_CHAT_ID", "")
	cfg.AllowKeywords = getStringSliceEnv("ALLOW_KEYWORDS", []string{"golang", "fullstack", "full-stack", "backend engineer", "go", "backend", "remote", "software", "engineer", "sde", "developer", "distributed", "systems", "cloud", "aws", "gcp", "azure", "kubernetes", "docker", "microservices", "api", "rest", "grpc", "sql", "nosql", "postgres", "mongodb", "redis", "elasticsearch", "ci/cd", "devops", "site reliability", "sre", "observability", "monitoring", "logging", "tracing", "security", "performance", "scalability", "high availability", "low latency", "concurrency", "parallelism", "message queues", "kafka", "rabbitmq", "aws sns", "aws sqs", "gcp pubsub", "azure service bus"})
	cfg.ExcludeKeywords = getStringSliceEnv("EXCLUDE_KEYWORDS", []string{"senior", "manager", "director", "lead", "principal", "staff", "vp", "c-level", "cto", "ceo", "founder", "co-founder"})
	cfg.maxFetchWorkers = getIntEnv("MAX_FETCH_WORKERS", 10)
	return &cfg
}

func BuildDbConfig(cfg *Config) db.Config {
	return db.Config{
		DbUrl:           cfg.DbUrl,
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	}
}
