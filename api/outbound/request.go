package outbound

import (
	"fmt"
	"net/mail"
	"strings"
)

const DefaultMaxRecipients = 50

type SendRequest struct {
	To       []string `json:"to"`
	Cc       []string `json:"cc,omitempty"`
	Bcc      []string `json:"bcc,omitempty"`
	Subject  string   `json:"subject"`
	BodyText string   `json:"body_text,omitempty"`
	BodyHTML string   `json:"body_html,omitempty"`
}

func NormalizeSendRequest(req SendRequest) (SendRequest, error) {
	var err error
	req.To, err = normalizeRecipients(req.To)
	if err != nil {
		return req, fmt.Errorf("invalid to: %w", err)
	}
	req.Cc, err = normalizeRecipients(req.Cc)
	if err != nil {
		return req, fmt.Errorf("invalid cc: %w", err)
	}
	req.Bcc, err = normalizeRecipients(req.Bcc)
	if err != nil {
		return req, fmt.Errorf("invalid bcc: %w", err)
	}
	req.Subject = strings.TrimSpace(req.Subject)
	req.BodyText = strings.TrimSpace(req.BodyText)
	req.BodyHTML = strings.TrimSpace(req.BodyHTML)
	return req, nil
}

func ValidateSendRequest(req SendRequest, maxRecipients int) error {
	if maxRecipients <= 0 {
		maxRecipients = DefaultMaxRecipients
	}
	if err := validateRecipients(req, maxRecipients); err != nil {
		return err
	}
	if strings.TrimSpace(req.Subject) == "" {
		return fmt.Errorf("subject is required")
	}
	if strings.TrimSpace(req.BodyText) == "" && strings.TrimSpace(req.BodyHTML) == "" {
		return fmt.Errorf("body_text or body_html is required")
	}
	return nil
}

func validateRecipients(req SendRequest, maxRecipients int) error {
	total := len(req.To) + len(req.Cc) + len(req.Bcc)
	if total == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if total > maxRecipients {
		return fmt.Errorf("too many recipients: %d > %d", total, maxRecipients)
	}
	return nil
}

func normalizeRecipients(values []string) ([]string, error) {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		addr, err := mail.ParseAddress(value)
		if err != nil {
			return nil, err
		}
		out = append(out, addr.Address)
	}
	return out, nil
}
