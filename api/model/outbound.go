package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboundStatusQueued = "queued"
	OutboundStatusSent   = "sent"
	OutboundStatusFailed = "failed"
)

type OutboundEmail struct {
	ID                uuid.UUID  `json:"id"`
	AccountID         uuid.UUID  `json:"account_id"`
	MailboxID         uuid.UUID  `json:"mailbox_id"`
	FromAddr          string     `json:"from_addr"`
	ReplyTo           string     `json:"reply_to"`
	ToAddrs           []string   `json:"to_addrs"`
	CcAddrs           []string   `json:"cc_addrs"`
	BccAddrs          []string   `json:"bcc_addrs,omitempty"`
	Subject           string     `json:"subject"`
	BodyText          string     `json:"body_text,omitempty"`
	BodyHTML          string     `json:"body_html,omitempty"`
	Provider          string     `json:"provider"`
	ProviderMessageID string     `json:"provider_message_id,omitempty"`
	Status            string     `json:"status"`
	Error             string     `json:"error,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	SentAt            *time.Time `json:"sent_at,omitempty"`
}

type CreateOutboundEmailInput struct {
	AccountID uuid.UUID
	MailboxID uuid.UUID
	FromAddr  string
	ReplyTo   string
	ToAddrs   []string
	CcAddrs   []string
	BccAddrs  []string
	Subject   string
	BodyText  string
	BodyHTML  string
	Provider  string
}
