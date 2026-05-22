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
	StackKeywords   []string
	DenyKeywords    []string
	PollInterval     time.Duration
	MaxOpenConns     int32
	MaxIdleConns     int32
	ConnMaxLifetime  time.Duration
	MaxFetchWorkers  int
	RateLimitSeconds int
}

var DefaultConfig = Config{
	StackKeywords: []string{
		"java", "spring", "spring boot", "kubernetes", "k8s", "prometheus", "grafana",
		"opentelemetry", "observability", "monitoring", "ci/cd", "github actions",
		"helm", "istio", "service mesh", "alertmanager",
		"kafka", "confluent", "aws", "lambda", "s3", "sqs", "sns", "ec2", "ses",
        "microservices", "event-driven architecture", "serverless", "audit trail", 
        "webhook", "goroutine", "concurrency",
        "postgresql", "postgres", "redis", "mongodb", "sql", "prisma", "sequelize", 
        "database management", "dbms", "query optimization", "indexing",
        "container runtime", "namespaces", "cgroups", "cgroups v2", "iptables", "nat", 
        "veth", "ipam", "network isolation", "wsl2", "dockerfile", "docker-compose",
        "javascript", "typescript", "node.js", "express.js", "react", "react.js", 
        "next.js", "trpc", "langchain", "webpack", 
        "gulp", "lazy loading", "stripe", "oauth2",
		"aws", "kubernetes", "k8s", "docker", "kafka", "microservices", "distributed",
        "postgresql", "postgres", "sql", "redis", "nosql", "ci/cd",
    },

    DenyKeywords: []string{
        "ios", "android", "flutter", "react native", "swift", "kotlin developer",
        "wordpress", "shopify", "magento", "php", "laravel", "web designer", "ui/ux",
        "scrum master", "project manager", "product manager", "product owner", 
        "qa engineer", "qa tester", "manual tester", "business analyst",
        "helpdesk", "it support", "hardware technician", "system administrator",
    },
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

func getInt32Env(key string, defaultVal int32) int32 {
	if val := os.Getenv(key); val != "" {
		var intVal int32
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

	cfg.PollInterval = time.Duration(getIntEnv("POLL_INTERVAL", 12)) * time.Hour
	cfg.DbUrl = getStringEnv("DB_URL", "sqlite:///jobs.db")
	cfg.MaxOpenConns = getInt32Env("DB_MAX_OPEN_CONNS", 10)
	cfg.MaxIdleConns = getInt32Env("DB_MAX_IDLE_CONNS", 5)
	cfg.ConnMaxLifetime = time.Duration(getIntEnv("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute
	cfg.TelegramBotToken = getStringEnv("TELEGRAM_BOT_TOKEN", "")
	cfg.TelegramChatId = getStringEnv("TELEGRAM_CHAT_ID", "")
	cfg.StackKeywords = getStringSliceEnv("STACK_KEYWORDS", DefaultConfig.StackKeywords)
	cfg.DenyKeywords = getStringSliceEnv("DENY_KEYWORDS", DefaultConfig.DenyKeywords)
	cfg.MaxFetchWorkers = getIntEnv("MAX_FETCH_WORKERS", 10)
	cfg.RateLimitSeconds = getIntEnv("RATELIMIT_SECONDS", 10)
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
