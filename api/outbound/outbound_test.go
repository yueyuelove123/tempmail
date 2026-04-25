package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateSendRequestRequiresBodyAndRecipientLimit(t *testing.T) {
	emptyBody := SendRequest{
		To:      []string{"to@example.com"},
		Subject: "hello",
	}
	if err := ValidateSendRequest(emptyBody, 50); err == nil {
		t.Fatalf("expected empty body request to fail")
	}

	tooMany := SendRequest{
		To:       make([]string, 51),
		Subject:  "hello",
		BodyText: "body",
	}
	for i := range tooMany.To {
		tooMany.To[i] = "to@example.com"
	}
	if err := ValidateSendRequest(tooMany, 50); err == nil {
		t.Fatalf("expected request with too many recipients to fail")
	}
}

func TestNormalizeSendRequestKeepsRecipientsResendCompatible(t *testing.T) {
	req, err := NormalizeSendRequest(SendRequest{
		To:      []string{" plain@example.com "},
		Cc:      []string{"Friend <friend@example.com>"},
		Subject: "hello",
		BodyText: "body",
	})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if req.To[0] != "plain@example.com" {
		t.Fatalf("to = %q, want plain@example.com", req.To[0])
	}
	if req.Cc[0] != "friend@example.com" {
		t.Fatalf("cc = %q, want friend@example.com", req.Cc[0])
	}
}

func TestResendSenderUsesConfiguredFromAndMailboxReplyTo(t *testing.T) {
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer re_test" {
			t.Fatalf("authorization header = %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"email_123"}`))
	}))
	defer server.Close()

	sender := NewResendSender(ResendConfig{
		APIKey:      "re_test",
		APIURL:      server.URL,
		FromAddress: "noreply@655588.xyz",
		FromName:    "TempMail",
		HTTPClient:  server.Client(),
	})
	result, err := sender.Send(context.Background(), Message{
		FromMailbox: "alice@test.655588.xyz",
		To:          []string{"bob@example.com"},
		Subject:     "hello",
		BodyText:    "plain body",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if result.ProviderMessageID != "email_123" {
		t.Fatalf("provider id = %q, want email_123", result.ProviderMessageID)
	}
	if got["from"] != "TempMail <noreply@655588.xyz>" {
		t.Fatalf("from = %v", got["from"])
	}
	if got["reply_to"] != "alice@test.655588.xyz" {
		t.Fatalf("reply_to = %v", got["reply_to"])
	}
	if got["text"] != "plain body" {
		t.Fatalf("text = %v", got["text"])
	}
}

func TestResendSenderReturnsProviderErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"bad to"}`))
	}))
	defer server.Close()

	sender := NewResendSender(ResendConfig{
		APIKey:      "re_test",
		APIURL:      server.URL,
		FromAddress: "noreply@655588.xyz",
		HTTPClient:  server.Client(),
	})
	_, err := sender.Send(context.Background(), Message{
		FromMailbox: "alice@test.655588.xyz",
		To:          []string{"bob@example.com"},
		Subject:     "hello",
		BodyText:    "plain body",
	})
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("error type = %T, want ProviderError", err)
	}
	if providerErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", providerErr.StatusCode)
	}
}
