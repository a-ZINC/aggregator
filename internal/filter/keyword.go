package filter

import (
	"context"
	"fmt"
	"strings"

	"github.com/a-ZINC/aggregator/internal/domain"
)

type Config struct {
	StackKeywords []string
	DenyKeywords  []string
}

type KeywordEvaluator struct {
	stackList []string
	denyList  []string
}

func NewKeywordEvaluator(cfg Config) *KeywordEvaluator {
	stack := make([]string, len(cfg.StackKeywords))
	for i, v := range cfg.StackKeywords {
		stack[i] = strings.ToLower(v)
	}
	deny := make([]string, len(cfg.DenyKeywords))
	for i, v := range cfg.DenyKeywords {
		deny[i] = strings.ToLower(v)
	}
	return &KeywordEvaluator{
		stackList: stack,
		denyList:  deny,
	}
}

func (e *KeywordEvaluator) Evaluate(ctx context.Context, job *domain.Job) (Evaluation, error) {
	title := strings.ToLower(job.Title)
	desc := strings.ToLower(job.Description)

	for _, word := range e.denyList {
		if strings.Contains(title, word) {
			return Evaluation{Eligible: false,
				Reason: fmt.Sprintf("Disqualified: Found blocklisted term '%s' in job title.", word),
				}, nil
		}
	}

	for _, word := range e.stackList {
		if strings.Contains(title, word) {
			return Evaluation{
				Eligible: true,
				Reason:   fmt.Sprintf("allowed: matched %q in title", word),
			}, nil
		}
		if tagsContain(job.Tags, word) {
			return Evaluation{
				Eligible: true,
				Reason:   fmt.Sprintf("allowed: matched %q in tags", word),
			}, nil
		}
		if strings.Contains(desc, word) {
			return Evaluation{
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
