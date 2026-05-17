package domain

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

type JobStatus string

const (
	New      JobStatus = "new"
	Notified JobStatus = "notified"
	Expired  JobStatus = "expired"
	Skipped  JobStatus = "skipped"
)

type Job struct {
	Id          string
	UrlHash     string
	Title       string
	Company     string
	Location    string
	Description string
	ApplyUrl    string
	Source      string
	Tags        []string
	SalaryMin   int
	SalaryMax   int
	Status      JobStatus
	PostedAt    *time.Time
	FetchedAt   time.Time
	NotifiedAt  *time.Time
	RawPayload  []byte
}

func HashUrl(url string) string {
	normUrl := strings.ToLower(strings.TrimSpace(url))
	sum := sha256.Sum256([]byte(normUrl))
	return fmt.Sprintf("%x", sum)
}