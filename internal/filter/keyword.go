package filter

import (
	"context"
	"crypto/des"
	"fmt"
	"strings"

	"github.com/a-ZINC/aggregator/internal/domain"
)

type config struct {
	AllowKeywords   []string
	ExcludeKeywords []string
}

type KeywordEvaluator struct {
	allowList []string
	denyList  []string
}

func NewKeywordEvaluator(cfg config) *KeywordEvaluator {
	allow := make([]string, len(cfg.AllowKeywords))
	for i, v := range cfg.AllowKeywords {
		allow[i] = strings.ToLower(v)
	}
	deny := make([]string, len(cfg.ExcludeKeywords))
	for i, v := range cfg.ExcludeKeywords {
		deny[i] = strings.ToLower(v)
	}
	return &KeywordEvaluator{
		allowList: allow,
		denyList:  deny,
	}
}

func (e *KeywordEvaluator) Evaluate(ctx context.Context, job *domain.Job) (Evaluation, error) {
	title := strings.ToLower(job.Title)
	desc := strings.ToLower(job.Description)

	for _, word := range e.denyList {
		if strings.Contains(title, word) {
			return Evaluation{Eligible: false,
				Reason: fmt.Sprintf("Disqualified: Found blocklisted term '%s' in job title.", word)
				}, nil
		}
	}

	for _, word := range e.allowList {
		if strings.Contains(title, word) {
			return Result{
				Eligible: true,
				Reason:   fmt.Sprintf("allowed: matched %q in title", word),
			}, nil
		}
		if tagsContain(job.Tags, word) {
			return Result{
				Eligible: true,
				Reason:   fmt.Sprintf("allowed: matched %q in tags", word),
			}, nil
		}
		if strings.Contains(desc, word) {
			return Result{
				Eligible: true,
				Reason:   fmt.Sprintf("allowed: matched %q in description", word),
			}, nil
		}
	}

	return Evaluation{Eligible: false, Reason: "Skipped: No keywords found."}, nil
}

func tagsContain(tags []string, keyword string) bool {
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), keyword) {
			return true
		}
	}
	return false
}
