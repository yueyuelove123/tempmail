-- ============================================================
-- TempMail v6 迁移 — Resend 发件日志
-- ============================================================

CREATE TABLE IF NOT EXISTS outbound_emails (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id          UUID         NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    mailbox_id          UUID         NOT NULL REFERENCES mailboxes(id) ON DELETE CASCADE,
    from_addr           VARCHAR(320) NOT NULL,
    reply_to            VARCHAR(320) NOT NULL DEFAULT '',
    to_addrs            TEXT[]       NOT NULL DEFAULT '{}',
    cc_addrs            TEXT[]       NOT NULL DEFAULT '{}',
    bcc_addrs           TEXT[]       NOT NULL DEFAULT '{}',
    subject             VARCHAR(998) NOT NULL DEFAULT '',
    body_text           TEXT         NOT NULL DEFAULT '',
    body_html           TEXT         NOT NULL DEFAULT '',
    provider            VARCHAR(32)  NOT NULL DEFAULT 'resend',
    provider_message_id VARCHAR(255) NOT NULL DEFAULT '',
    status              VARCHAR(32)  NOT NULL DEFAULT 'queued',
    error               TEXT         NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    sent_at             TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbound_account_created
ON outbound_emails (account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_outbound_mailbox_created
ON outbound_emails (mailbox_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_outbound_status
ON outbound_emails (status);
