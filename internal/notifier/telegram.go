package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/a-ZINC/aggregator/internal/domain"
)

type TelegramNotifier struct {
	token  string
	chatId string
	httpClient *http.Client
	ratelimiter <-chan time.Time
}

func NewTelegramNotifier(token, chatId string, rateLimitSeconds int) *TelegramNotifier {
	return &TelegramNotifier{
		token:  token,
		chatId: chatId,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		ratelimiter: time.Tick(time.Duration(rateLimitSeconds) * time.Second),
	}
}

func (t *TelegramNotifier) Notify(ctx context.Context, job *domain.Job, reason string) error {
	apiUrl := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	formattedText := formatMessage(job)
	data := url.Values{}
	data.Set("chat_id", t.chatId)
	data.Set("text", formattedText)
	data.Set("parse_mode", "HTML")
	data.Set("disable_web_page_preview", "true")

	// Wait for the rate limiter before sending the request
	<-t.ratelimiter

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	// CRITICAL: Set content-type to urlencoded form matching the body structure
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var payload map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&payload)
		return fmt.Errorf("telegram API returned status %d: %v", resp.StatusCode, payload)
	}
	return nil
}

func formatMessage(job *domain.Job) string {
	var sb strings.Builder

	sb.WriteString("🎯 <b>New Remote Job Match</b>\n\n")
	sb.WriteString(fmt.Sprintf("<b>Role:</b> %s\n", html.EscapeString(job.Title)))
	sb.WriteString(fmt.Sprintf("<b>Company:</b> %s\n", html.EscapeString(job.Company)))
	sb.WriteString(fmt.Sprintf("<b>Source:</b> %s\n", html.EscapeString(job.Source)))

	if job.Location != "" {
		sb.WriteString(fmt.Sprintf("<b>Location:</b> %s\n", html.EscapeString(job.Location)))
	}

	if job.SalaryMin > 0 {
		sb.WriteString(fmt.Sprintf("<b>Salary:</b> $%d — $%d\n", job.SalaryMin, job.SalaryMax))
	}

	if len(job.Tags) > 0 {
		// Show max 5 tags — enough signal without flooding the message
		tags := job.Tags
		if len(tags) > 5 {
			tags = tags[:5]
		}
		sb.WriteString(fmt.Sprintf("<b>Tags:</b> %s\n", strings.Join(tags, ", ")))
	}

	sb.WriteString(fmt.Sprintf("\n🔗 <a href=\"%s\">Apply Now</a>", job.ApplyUrl))

	return sb.String()
}