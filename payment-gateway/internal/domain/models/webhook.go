package models

// WebhookResult represents the outcome of processing a webhook event.
type WebhookResult string

const (
	WebhookCredited  WebhookResult = "credited"
	WebhookDuplicate WebhookResult = "duplicate"
	WebhookIgnored   WebhookResult = "ignored"
)
