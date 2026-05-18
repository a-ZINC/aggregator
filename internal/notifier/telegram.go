package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
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
}

func NewTelegramNotifier(token, chatId string) *TelegramNotifier {
	return &TelegramNotifier{
		token:  token,
		chatId: chatId,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *TelegramNotifier) Notify(ctx context.Context, job *domain.Job, reason string) error {
	apiUrl := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	safeTitle := html.EscapeString(job.Title)
	safeCompany := html.EscapeString(job.Company)
	safeSource := html.EscapeString(job.Source)
	safeReason := html.EscapeString(reason)
	safeUrl := html.EscapeString(job.ApplyUrl)

	formattedText := fmt.Sprintf(
		"🎯 <b>New High-Signal Remote Match</b>\n\n"+
			"<b>Role:</b> %s\n"+
			"<b>Company:</b> %s\n"+
			"<b>Source:</b> %s\n\n"+
			"💡 <b>Verdict:</b> %s\n\n"+
			"🔗 <a href=\"%s\">Apply to Position</a>",
		safeTitle, safeCompany, safeSource, safeReason, safeUrl,
	)

	log.Printf("[DEBUG] Target Chat ID: %q", t.chatId)
	log.Printf("[DEBUG] Job Title Raw: %q -> Safe: %q", job.Title, safeTitle)
	log.Printf("[DEBUG] Job URL Raw: %q -> Safe: %q", job.ApplyUrl, safeUrl)
	log.Printf("[DEBUG] Final Formatted Text:\n%s\n", formattedText)

	data := url.Values{}
	data.Set("chat_id", t.chatId)
	data.Set("text", formattedText)
	data.Set("parse_mode", "HTML")

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