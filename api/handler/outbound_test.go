package handler

import (
	"errors"
	"net/http"
	"testing"

	"tempmail/outbound"
)

func TestSendFailureHTTPStatusPreservesProviderClientErrors(t *testing.T) {
	err := &outbound.ProviderError{StatusCode: http.StatusUnprocessableEntity}
	if got := sendFailureHTTPStatus(err); got != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", got)
	}
}

func TestSendFailureHTTPStatusUsesBadGatewayForGenericErrors(t *testing.T) {
	if got := sendFailureHTTPStatus(errors.New("network failed")); got != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", got)
	}
}
