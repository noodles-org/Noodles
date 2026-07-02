package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	maxRetries     = 4
	retryBaseDelay = 500 * time.Millisecond
	maxTotalWait   = 10 * time.Second
)

var (
	emailsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webhook_emails_received_total",
		Help: "Total email.received events received from Resend",
	})
	emailsDelivered = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webhook_emails_delivered_total",
		Help: "Total emails successfully delivered to Stalwart",
	})
	emailsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_emails_failed_total",
		Help: "Total emails that failed to process",
	}, []string{"stage", "retryable"})
	retryAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_retry_attempts_total",
		Help: "Total retry attempts made, by operation",
	}, []string{"operation"})
	processingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "webhook_processing_duration_seconds",
		Help:    "End-to-end time to process an inbound email",
		Buckets: prometheus.DefBuckets,
	})
)

func isRetryable(err error) bool {
	var ae *apiError
	if errors.As(err, &ae) {
		if ae.StatusCode == http.StatusTooManyRequests {
			return true
		}
		if ae.StatusCode >= 400 && ae.StatusCode < 500 {
			return false
		}
		return true
	}
	return true
}

func withRetries(operation string, fn func() error) error {
	start := time.Now()
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if !isRetryable(lastErr) {
			log.Printf("%s: non-retryable error, giving up: %v", operation, lastErr)
			return lastErr
		}

		if attempt == maxRetries {
			break
		}

		delay := retryBaseDelay * time.Duration(1<<(attempt-1))
		if time.Since(start)+delay > maxTotalWait {
			log.Printf("%s: giving up, would exceed max wait budget", operation)
			break
		}

		retryAttempts.WithLabelValues(operation).Inc()
		log.Printf("%s: attempt %d/%d failed: %v (retrying in %s)", operation, attempt, maxRetries, lastErr, delay)
		time.Sleep(delay)
	}
	log.Printf("%s: exhausted retries, giving up: %v", operation, lastErr)
	return lastErr
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
		return nil, &apiError{StatusCode: resp.StatusCode, Body: string(body)}
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
		body, _ := io.ReadAll(resp.Body)
		return nil, &apiError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	return io.ReadAll(resp.Body)
}

func deliverToStalwart(from string, to []string, rawMessage []byte) error {
	smtpAddr := os.Getenv("STALWART_SMTP_ADDR")
	if smtpAddr == "" {
		smtpAddr = "localhost:25"
	}

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

	ehloHost := os.Getenv("SMTP_EHLO_HOST")
	if ehloHost == "" {
		ehloHost = "mail.noodles.quest"
	}
	if err := c.Hello(ehloHost); err != nil {
		return fmt.Errorf("EHLO %s: %w", ehloHost, err)
	}

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

func processInboundEmail(payload WebhookPayload, apiKey string) error {
	start := time.Now()
	defer func() { processingDuration.Observe(time.Since(start).Seconds()) }()

	var email *ReceivedEmail
	err := withRetries("fetch_received_email", func() error {
		e, err := fetchReceivedEmail(payload.Data.EmailID, apiKey)
		if err != nil {
			return err
		}
		email = e
		return nil
	})
	if err != nil {
		emailsFailed.WithLabelValues("fetch", strconv.FormatBool(isRetryable(err))).Inc()
		return fmt.Errorf("fetching email %s: %w", payload.Data.EmailID, err)
	}

	if email.Raw == nil || email.Raw.DownloadURL == "" {
		log.Printf("no raw email available for %s, skipping", payload.Data.EmailID)
		return nil
	}

	var rawMessage []byte
	err = withRetries("download_raw_email", func() error {
		b, err := downloadRawEmail(email.Raw.DownloadURL)
		if err != nil {
			return err
		}
		rawMessage = b
		return nil
	})
	if err != nil {
		emailsFailed.WithLabelValues("download", strconv.FormatBool(isRetryable(err))).Inc()
		return fmt.Errorf("downloading raw email %s: %w", payload.Data.EmailID, err)
	}

	err = withRetries("deliver_to_stalwart", func() error {
		return deliverToStalwart(email.From, email.To, rawMessage)
	})
	if err != nil {
		emailsFailed.WithLabelValues("deliver", strconv.FormatBool(isRetryable(err))).Inc()
		return fmt.Errorf("delivering email %s to stalwart: %w", payload.Data.EmailID, err)
	}

	emailsDelivered.Inc()
	log.Printf("delivered email %s to stalwart", payload.Data.EmailID)
	return nil
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
	emailsReceived.Inc()
	log.Printf("received email %s from %s to %v subject=%q",
		payload.Data.EmailID, payload.Data.From, payload.Data.To, payload.Data.Subject)

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Printf("CONFIG ERROR: RESEND_API_KEY not set, cannot process email %s", payload.Data.EmailID)
		emailsFailed.WithLabelValues("config", "false").Inc()
		w.WriteHeader(http.StatusOK)
		return
	}

	// Always 200 to Resend regardless of outcome. Retries and handling are done in our logic.
	if err := processInboundEmail(payload, apiKey); err != nil {
		log.Printf("FAILED to process email %s after retries: %v", payload.Data.EmailID, err)
	}

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
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("inbound webhook handler listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
