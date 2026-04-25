package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"tempmail/middleware"
	"tempmail/model"
	"tempmail/outbound"
	"tempmail/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OutboundHandler struct {
	store         *store.Store
	sender        outbound.Sender
	enabled       bool
	fromAddress   string
	maxRecipients int
}

func NewOutboundHandler(s *store.Store, sender outbound.Sender, enabled bool, fromAddress string, maxRecipients int) *OutboundHandler {
	if maxRecipients <= 0 {
		maxRecipients = outbound.DefaultMaxRecipients
	}
	return &OutboundHandler{s, sender, enabled, fromAddress, maxRecipients}
}

func (h *OutboundHandler) Send(c *gin.Context) {
	if !h.enabled || h.sender == nil || strings.TrimSpace(h.fromAddress) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "outbound email is not configured"})
		return
	}
	account := middleware.GetAccount(c)
	mailbox, ok := h.loadMailbox(c, account)
	if !ok {
		return
	}
	req, ok := h.parseRequest(c)
	if !ok {
		return
	}
	record, ok := h.createRecord(c, account, mailbox, req)
	if !ok {
		return
	}
	result, err := h.sender.Send(c.Request.Context(), messageFrom(mailbox, req))
	if err != nil {
		h.respondSendFailure(c, record.ID, err)
		return
	}
	sent, err := h.store.MarkOutboundEmailSent(c.Request.Context(), record.ID, result.ProviderMessageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "sent", "outbound_email": sent})
}

func (h *OutboundHandler) loadMailbox(c *gin.Context, account *model.Account) (*model.Mailbox, bool) {
	mailboxID, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mailbox id"})
		return nil, false
	}
	mailbox, err := h.store.GetMailbox(c.Request.Context(), mailboxID, account.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "mailbox not found"})
		return nil, false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, false
	}
	if mailboxExpired(mailbox) {
		c.JSON(http.StatusGone, gin.H{"error": "mailbox expired"})
		return nil, false
	}
	return mailbox, true
}

func (h *OutboundHandler) parseRequest(c *gin.Context) (outbound.SendRequest, bool) {
	var req outbound.SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return req, false
	}
	req, err := outbound.NormalizeSendRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return req, false
	}
	if err := outbound.ValidateSendRequest(req, h.maxRecipients); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return req, false
	}
	return req, true
}

func (h *OutboundHandler) createRecord(c *gin.Context, account *model.Account, mailbox *model.Mailbox, req outbound.SendRequest) (*model.OutboundEmail, bool) {
	record, err := h.store.CreateOutboundEmail(c.Request.Context(), model.CreateOutboundEmailInput{
		AccountID: account.ID, MailboxID: mailbox.ID, FromAddr: h.fromAddress,
		ReplyTo: mailbox.FullAddress, ToAddrs: req.To, CcAddrs: req.Cc, BccAddrs: req.Bcc,
		Subject: req.Subject, BodyText: req.BodyText, BodyHTML: req.BodyHTML, Provider: "resend",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, false
	}
	return record, true
}

func (h *OutboundHandler) respondSendFailure(c *gin.Context, recordID uuid.UUID, sendErr error) {
	record, err := h.store.MarkOutboundEmailFailed(c.Request.Context(), recordID, sendErr.Error())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "send failed", "detail": sendErr.Error()})
		return
	}
	c.JSON(http.StatusBadGateway, gin.H{"error": "send failed", "detail": sendErr.Error(), "outbound_email": record})
}

func messageFrom(mailbox *model.Mailbox, req outbound.SendRequest) outbound.Message {
	return outbound.Message{
		FromMailbox: mailbox.FullAddress,
		To:          req.To,
		Cc:          req.Cc,
		Bcc:         req.Bcc,
		Subject:     req.Subject,
		BodyText:    req.BodyText,
		BodyHTML:    req.BodyHTML,
	}
}

func mailboxExpired(mailbox *model.Mailbox) bool {
	return !mailbox.IsPermanent && mailbox.ExpiresAt != nil && time.Now().After(*mailbox.ExpiresAt)
}
