package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultResendAPIURL = "https://api.resend.com/emails"

type ResendConfig struct {
	APIKey      string
	APIURL      string
	FromAddress string
	FromName    string
	HTTPClient  *http.Client
}

type Message struct {
	FromMailbox string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	BodyText    string
	BodyHTML    string
}

type SendResult struct {
	ProviderMessageID string
}

type Sender interface {
	Send(ctx context.Context, msg Message) (*SendResult, error)
}

type ResendSender struct {
	cfg    ResendConfig
	client *http.Client
}

func NewResendSender(cfg ResendConfig) *ResendSender {
	if cfg.APIURL == "" {
		cfg.APIURL = defaultResendAPIURL
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &ResendSender{cfg: cfg, client: client}
}

func (s *ResendSender) Send(ctx context.Context, msg Message) (*SendResult, error) {
	if strings.TrimSpace(s.cfg.APIKey) == "" {
		return nil, fmt.Errorf("resend api key is not configured")
	}
	payload := s.buildPayload(msg)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.APIURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	return s.do(req)
}

func (s *ResendSender) buildPayload(msg Message) resendPayload {
	return resendPayload{
		From:    formatFrom(s.cfg.FromName, s.cfg.FromAddress),
		To:      msg.To,
		Cc:      msg.Cc,
		Bcc:     msg.Bcc,
		Subject: msg.Subject,
		Text:    msg.BodyText,
		HTML:    msg.BodyHTML,
		ReplyTo: strings.TrimSpace(msg.FromMailbox),
	}
}

func (s *ResendSender) do(req *http.Request) (*SendResult, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("resend status %d: %s", resp.StatusCode, readLimited(resp.Body))
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &SendResult{ProviderMessageID: result.ID}, nil
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Cc      []string `json:"cc,omitempty"`
	Bcc     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject"`
	Text    string   `json:"text,omitempty"`
	HTML    string   `json:"html,omitempty"`
	ReplyTo string   `json:"reply_to,omitempty"`
}

func formatFrom(name, address string) string {
	address = strings.TrimSpace(address)
	name = strings.TrimSpace(name)
	if name == "" {
		return address
	}
	return fmt.Sprintf("%s <%s>", name, address)
}

func readLimited(r io.Reader) string {
	data, _ := io.ReadAll(io.LimitReader(r, 4096))
	return strings.TrimSpace(string(data))
}
