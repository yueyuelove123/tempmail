package store

import (
	"context"
	"time"

	"tempmail/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const outboundColumns = `
	id, account_id, mailbox_id, from_addr, reply_to,
	to_addrs, cc_addrs, bcc_addrs, subject, body_text, body_html,
	provider, provider_message_id, status, error, created_at, sent_at`

func (s *Store) CreateOutboundEmail(ctx context.Context, in model.CreateOutboundEmailInput) (*model.OutboundEmail, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO outbound_emails (
			account_id, mailbox_id, from_addr, reply_to,
			to_addrs, cc_addrs, bcc_addrs, subject, body_text, body_html,
			provider, status
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING `+outboundColumns,
		in.AccountID, in.MailboxID, in.FromAddr, in.ReplyTo,
		in.ToAddrs, in.CcAddrs, in.BccAddrs, in.Subject, in.BodyText, in.BodyHTML,
		in.Provider, model.OutboundStatusQueued,
	)
	return scanOutboundEmail(row)
}

func (s *Store) MarkOutboundEmailSent(ctx context.Context, id uuid.UUID, providerID string) (*model.OutboundEmail, error) {
	now := time.Now()
	row := s.pool.QueryRow(ctx, `
		UPDATE outbound_emails
		SET status = $2, provider_message_id = $3, error = '', sent_at = $4
		WHERE id = $1
		RETURNING `+outboundColumns,
		id, model.OutboundStatusSent, providerID, now,
	)
	return scanOutboundEmail(row)
}

func (s *Store) MarkOutboundEmailFailed(ctx context.Context, id uuid.UUID, reason string) (*model.OutboundEmail, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE outbound_emails
		SET status = $2, error = $3
		WHERE id = $1
		RETURNING `+outboundColumns,
		id, model.OutboundStatusFailed, reason,
	)
	return scanOutboundEmail(row)
}

func scanOutboundEmail(row pgx.Row) (*model.OutboundEmail, error) {
	var e model.OutboundEmail
	err := row.Scan(
		&e.ID, &e.AccountID, &e.MailboxID, &e.FromAddr, &e.ReplyTo,
		&e.ToAddrs, &e.CcAddrs, &e.BccAddrs, &e.Subject, &e.BodyText, &e.BodyHTML,
		&e.Provider, &e.ProviderMessageID, &e.Status, &e.Error, &e.CreatedAt, &e.SentAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}
