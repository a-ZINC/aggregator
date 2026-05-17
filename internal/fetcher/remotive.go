package fetcher

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a-ZINC/aggregator/internal/domain"
)

const (
	remotiveApiUrl = "https://remotive.io/api/remote-jobs"
	remotiveDateFmt = "2006-01-02T15:04:05"
	remotiveTimeout = 15 * time.Second
)

type RemotiveFetcher struct {
	Client *http.Client
}

type remotiveJob struct {
	Id                        int      `json:"id"`
	Url                       string   `json:"url"`
	Title                     string   `json:"title"`
	Company                   string   `json:"company_name"`
	CompanyLogo               string   `json:"company_logo"`
	Category                  string   `json:"category"`
	Tags                      []string `json:"tags"`
	Description               string   `json:"description"`
	JobType                   string   `json:"job_type"`
	PublicationDate           string   `json:"publication_date"`
	CandidateRequiredLocation string   `json:"candidate_required_location"`
	Salary                    string   `json:"salary"`
}

type remotiveResponse struct {
	Jobs []remotiveJob `json:"jobs"`
	JobCount      int `json:"job-count"`
	TotalJobCount int `json:"total_job_count"`
}

func NewRemotiveFetcher() *RemotiveFetcher {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		DisableCompression: false,
	}
	return &RemotiveFetcher{
		Client: &http.Client{Timeout: remotiveTimeout, Transport: transport},

	}
}

func (f *RemotiveFetcher) Fetch(ctx context.Context) ([]domain.Job, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remotiveApiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "aggregator-bot/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remotive API returned status %d", resp.StatusCode)
	}

	var rResp remotiveResponse
	if err := json.NewDecoder(resp.Body).Decode(&rResp); err != nil {
		return nil, err
	}

	var jobs []domain.Job
	for _, rJob := range rResp.Jobs {
		jobs = append(jobs, mappperRemotiveToDomainJob(rJob))
	}
	return jobs, nil
}

func (f *RemotiveFetcher) Name() string {
	return "Remotive"
}

func normalizeTags(tags []string) []string {
	var normalized []string
	for _, tag := range tags {
		normalized = append(normalized, strings.ToLower(strings.TrimSpace(tag)))
	}
	return normalized
}

func mappperRemotiveToDomainJob(rJob remotiveJob) domain.Job {
	rawPayload, _ := json.Marshal(rJob)
	var postedAt *time.Time
	if rJob.PublicationDate != "" {
		t, err := time.Parse(remotiveDateFmt, rJob.PublicationDate)
		if err == nil {
			postedAt = &t
		}
	}
	return domain.Job{
		Title:       strings.TrimSpace(rJob.Title),
		Company:     strings.TrimSpace(rJob.Company),
		Location:    strings.TrimSpace(rJob.CandidateRequiredLocation),
		Description: strings.TrimSpace(rJob.Description),
		ApplyUrl:    strings.TrimSpace(rJob.Url),
		Source:      "Remotive",
		Tags:        normalizeTags(rJob.Tags),
		SalaryMin:   0,
		SalaryMax:   0,
		Status:      domain.New,
		FetchedAt: time.Now().UTC(),
		PostedAt:  postedAt,
		RawPayload: rawPayload,
		UrlHash: domain.HashUrl(rJob.Url),
	}
}
