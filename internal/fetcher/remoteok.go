package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/a-ZINC/aggregator/internal/domain"
)

const (
	remoteOkApiUrl  = "https://remoteok.com/api"
	remoteOkTimeout = 15 * time.Second
	remoteOkDateFmt = "2006-01-02T15:04:05Z07:00"
)

type RemoteOkFetcher struct {
	client *http.Client
}

type remoteOkJob struct {
	Slug        string    `json:"slug"`
	Id          string       `json:"id"`
	Epoch       int64     `json:"epoch"`
	Date        string    `json:"date"`
	Company     string    `json:"company"`
	CompanyLogo string    `json:"company_logo"`
	Position    string    `json:"position"`
	Tags        []string  `json:"tags"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	ApplyUrl    string    `json:"apply_url"`
	SalaryMin   int       `json:"salary_min"`
	SalaryMax   int       `json:"salary_max"`
	Logo        string    `json:"logo"`
	Url         string    `json:"url"`
}

func NewRemoteOkFetcher() *RemoteOkFetcher {
	return &RemoteOkFetcher{
		client: &http.Client{Timeout: remoteOkTimeout},
	}
}

func (f *RemoteOkFetcher) Fetch(ctx context.Context) ([]domain.Job, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remoteOkApiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "aggregator-bot/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remoteok API returned status %d", resp.StatusCode)
	}

	var raw []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	
	jobs := make([]domain.Job, 0, len(raw)-1)
	for _, item := range raw[1:] {
		var rj remoteOkJob
		if err := json.Unmarshal(item, &rj); err != nil {
			log.Printf("failed to unmarshal remoteok job: %v", err)
			continue
		}
		jobs = append(jobs, mappingRemoteOkJobToDomainJob(rj))
	}

	return jobs, nil
}

func (f *RemoteOkFetcher) Name() string {
	return "RemoteOk"
}

func dateResolver(dateStr string) *time.Time {
	t, err := time.Parse(remoteOkDateFmt, dateStr)
	if err != nil {
		return nil
	}
	return &t
}


func mappingRemoteOkJobToDomainJob(rj remoteOkJob) domain.Job {
	rawPayload, _ := json.Marshal(rj)
	return domain.Job{
		Title:       strings.TrimSpace(rj.Position),
		Company:     strings.TrimSpace(rj.Company),
		Location:    strings.TrimSpace(rj.Location),
		Description: strings.TrimSpace(rj.Description),
		ApplyUrl:    strings.TrimSpace(rj.ApplyUrl),
		Source:      "RemoteOk",
		Tags:        normalizeTags(rj.Tags),
		SalaryMin:   rj.SalaryMin,
		SalaryMax:   rj.SalaryMax,
		Status:      domain.New,
		FetchedAt:   time.Now().UTC(),
		PostedAt:    dateResolver(rj.Date),
		UrlHash:     domain.HashUrl(rj.Url),
		RawPayload:  rawPayload,
	}
}
