package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

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

func fetchReceivedEmail(emailID string, apiKey string) (*ReceivedEmail, error) {
	url := fmt.Sprintf("https://api.resend.com/emails/receiving/%s", emailID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching email: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("resend API returned %d: %s", resp.StatusCode, string(body))
	}

	var email ReceivedEmail
	if err := json.NewDecoder(resp.Body).Decode(&email); err != nil {
		return nil, fmt.Errorf("decoding email: %w", err)
	}
	return &email, nil
}

func downloadRawEmail(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading raw email: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("raw download returned %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func deliverToStalwart(from string, to []string, rawMessage []byte) error {
	smtpAddr := os.Getenv("STALWART_SMTP_ADDR")
	if smtpAddr == "" {
		smtpAddr = "localhost:25"
	}

	// Extract just the email address from "Name <email>" format
	sender := extractEmail(from)

	c, err := smtp.Dial(smtpAddr)
	if err != nil {
		return fmt.Errorf("connecting to stalwart: %w", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("error closing smtp connection: %v", err)
		}
	}()

	if err := c.Mail(sender); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}

	for _, recipient := range to {
		if err := c.Rcpt(extractEmail(recipient)); err != nil {
			return fmt.Errorf("RCPT TO %s: %w", recipient, err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}

	if _, err := w.Write(rawMessage); err != nil {
		return fmt.Errorf("writing message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("closing data: %w", err)
	}

	return c.Quit()
}

func extractEmail(addr string) string {
	if idx := strings.Index(addr, "<"); idx != -1 {
		end := strings.Index(addr, ">")
		if end > idx {
			return addr[idx+1 : end]
		}
	}
	return strings.TrimSpace(addr)
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading body: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("error closing request body: %v", err)
		}
	}()

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("error parsing webhook: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if payload.Type != "email.received" {
		log.Printf("ignoring event type: %s", payload.Type)
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("received email %s from %s to %v subject=%q",
		payload.Data.EmailID, payload.Data.From, payload.Data.To, payload.Data.Subject)

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Printf("RESEND_API_KEY not set")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// Fetch full email content from Resend API
	email, err := fetchReceivedEmail(payload.Data.EmailID, apiKey)
	if err != nil {
		log.Printf("error fetching email %s: %v", payload.Data.EmailID, err)
		http.Error(w, "failed to fetch email", http.StatusInternalServerError)
		return
	}

	// Download raw RFC822 message if available
	var rawMessage []byte
	if email.Raw != nil && email.Raw.DownloadURL != "" {
		rawMessage, err = downloadRawEmail(email.Raw.DownloadURL)
		if err != nil {
			log.Printf("error downloading raw email: %v", err)
			http.Error(w, "failed to download raw email", http.StatusInternalServerError)
			return
		}
	} else {
		log.Printf("no raw email available for %s, skipping", payload.Data.EmailID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Deliver to Stalwart via local SMTP
	if err := deliverToStalwart(email.From, email.To, rawMessage); err != nil {
		log.Printf("error delivering to stalwart: %v", err)
		http.Error(w, "delivery failed", http.StatusInternalServerError)
		return
	}

	log.Printf("delivered email %s to stalwart", payload.Data.EmailID)
	w.WriteHeader(http.StatusOK)
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/webhook/inbound", webhookHandler)
	http.HandleFunc("/healthz", healthHandler)

	log.Printf("inbound webhook handler listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
