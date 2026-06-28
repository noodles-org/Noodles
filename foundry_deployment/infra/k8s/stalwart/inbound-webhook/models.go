package main

import "fmt"

type WebhookPayload struct {
	Type string      `json:"type"`
	Data WebhookData `json:"data"`
}

type WebhookData struct {
	EmailID string   `json:"email_id"`
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
}

type ReceivedEmail struct {
	From    string            `json:"from"`
	To      []string          `json:"to"`
	Subject string            `json:"subject"`
	HTML    string            `json:"html"`
	Text    *string           `json:"text"`
	Headers map[string]string `json:"headers"`
	Raw     *RawEmail         `json:"raw"`
}

type RawEmail struct {
	DownloadURL string `json:"download_url"`
	ExpiresAt   string `json:"expires_at"`
}
type apiError struct {
	StatusCode int
	Body       string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Body)
}
